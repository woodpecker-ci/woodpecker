// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build test

package wasm_test

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm/wasmtest"
)

func newRunner(t *testing.T, change func(*wasm.Limits)) *wasm.Runner {
	t.Helper()

	limits := wasm.DefaultLimits()
	if change != nil {
		change(&limits)
	}

	runner, err := wasm.NewRunner(t.Context(), wasmtest.Build(t, wasmtest.ProbeGuest), limits)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, runner.Close(context.Background())) })

	return runner
}

func TestRunnerPassesData(t *testing.T) {
	runner := newRunner(t, nil)

	out, err := runner.Run(t.Context(), []string{"guest", "echo"}, []byte(`{"some":"input"}`))
	require.NoError(t, err)
	assert.JSONEq(t, `{"some":"input"}`, string(out))

	out, err = runner.Run(t.Context(), []string{"guest", "args", "extra"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "guest,args,extra", string(out))
}

func TestRunnerReportsFailure(t *testing.T) {
	runner := newRunner(t, nil)

	_, err := runner.Run(t.Context(), []string{"guest", "fail"}, nil)

	var exit *wasm.ExitError
	require.ErrorAs(t, err, &exit)
	assert.EqualValues(t, 3, exit.Code)
	assert.Equal(t, "something went wrong", exit.Stderr)
}

func TestRunnerSandbox(t *testing.T) {
	t.Setenv("WOODPECKER_AGENT_SECRET", "must-not-leak")
	runner := newRunner(t, nil)

	out, err := runner.Run(t.Context(), []string{"guest", "probe"}, nil)
	require.NoError(t, err)

	assert.Equal(t, []string{
		"env=false",
		"read_root=false",
		"read_cwd=false",
		"write_file=false",
		"extra_descriptor=false",
	}, strings.Fields(string(out)))
}

func TestRunnerKeepsNoState(t *testing.T) {
	runner := newRunner(t, nil)

	for range 3 {
		out, err := runner.Run(t.Context(), []string{"guest", "calls"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "1", string(out))
	}
}

func TestRunnerIsolatesParallelCalls(t *testing.T) {
	runner := newRunner(t, nil)

	var wg sync.WaitGroup
	outputs := make([]string, wasm.DefaultLimits().Concurrency)
	for i := range outputs {
		wg.Go(func() {
			out, err := runner.Run(t.Context(), []string{"guest", "echo"}, []byte(strings.Repeat(strconv.Itoa(i), 1<<16)))
			assert.NoError(t, err)
			outputs[i] = string(out)
		})
	}
	wg.Wait()

	for i, output := range outputs {
		assert.Equal(t, strings.Repeat(strconv.Itoa(i), 1<<16), output, "call %d got the data of another one", i)
	}
}

func TestRunnerLimits(t *testing.T) {
	t.Run("run time", func(t *testing.T) {
		runner := newRunner(t, func(l *wasm.Limits) { l.Timeout = 300 * time.Millisecond })

		start := time.Now()
		_, err := runner.Run(t.Context(), []string{"guest", "spin"}, nil)
		require.ErrorIs(t, err, context.DeadlineExceeded)
		assert.Less(t, time.Since(start), 5*time.Second)
	})

	t.Run("caller gives up", func(t *testing.T) {
		runner := newRunner(t, nil)
		ctx, cancel := context.WithCancelCause(t.Context())
		time.AfterFunc(300*time.Millisecond, func() { cancel(nil) })

		_, err := runner.Run(ctx, []string{"guest", "spin"}, nil)
		require.ErrorIs(t, err, context.Canceled)
	})

	t.Run("memory", func(t *testing.T) {
		runner := newRunner(t, func(l *wasm.Limits) { l.MemoryBytes = 32 << 20 })

		_, err := runner.Run(t.Context(), []string{"guest", "alloc"}, nil)
		var exit *wasm.ExitError
		require.ErrorAs(t, err, &exit, "the guest must run out of memory before the time is up")
		assert.Contains(t, exit.Stderr, "out of memory")
	})

	t.Run("stdout", func(t *testing.T) {
		runner := newRunner(t, func(l *wasm.Limits) { l.StdoutBytes = 1 << 20 })

		_, err := runner.Run(t.Context(), []string{"guest", "flood-stdout"}, nil)
		require.ErrorIs(t, err, wasm.ErrOutputTooLarge)
	})

	t.Run("stderr is cut but does not fail the call", func(t *testing.T) {
		runner := newRunner(t, func(l *wasm.Limits) { l.StderrBytes = 1 << 10 })

		out, err := runner.Run(t.Context(), []string{"guest", "flood-stderr"}, nil)
		require.NoError(t, err)
		assert.Equal(t, "done", string(out))
	})

	t.Run("parallel calls", func(t *testing.T) {
		runner := newRunner(t, func(l *wasm.Limits) {
			l.Concurrency = 1
			l.Timeout = time.Second
		})

		// occupies the only slot until its time is up
		busy := make(chan error, 1)
		go func() {
			_, err := runner.Run(t.Context(), []string{"guest", "spin"}, nil)
			busy <- err
		}()
		time.Sleep(200 * time.Millisecond)

		ctx, cancel := context.WithTimeout(t.Context(), 200*time.Millisecond)
		defer cancel()
		_, err := runner.Run(ctx, []string{"guest", "echo"}, nil)
		require.ErrorContains(t, err, "waiting for a free wasm slot")
		require.ErrorIs(t, err, context.DeadlineExceeded)

		require.ErrorContains(t, <-busy, "wasm module was stopped")

		// the slot is free again
		_, err = runner.Run(t.Context(), []string{"guest", "echo"}, nil)
		require.NoError(t, err)
	})
}

func TestNewRunnerRejects(t *testing.T) {
	t.Run("not a wasm module", func(t *testing.T) {
		_, err := wasm.NewRunner(t.Context(), []byte("#!/bin/sh\necho no"), wasm.DefaultLimits())
		require.ErrorContains(t, err, "could not compile wasm module")
	})

	t.Run("module without entrypoint", func(t *testing.T) {
		// the smallest valid module: magic number and version only
		_, err := wasm.NewRunner(t.Context(), []byte("\x00asm\x01\x00\x00\x00"), wasm.DefaultLimits())
		require.ErrorContains(t, err, "not a wasi command")
	})

	t.Run("module asking for host functions the sandbox does not offer", func(t *testing.T) {
		// (module (import "env" "host_call" (func)) (func (export "_start")))
		module := []byte("\x00asm\x01\x00\x00\x00" +
			"\x01\x04\x01\x60\x00\x00" + // type section: func() -> ()
			"\x02\x11\x01\x03env\x09host_call\x00\x00" + // import section
			"\x03\x02\x01\x00" + // function section
			"\x07\x0a\x01\x06_start\x00\x01" + // export section
			"\x0a\x04\x01\x02\x00\x0b") // code section
		_, err := wasm.NewRunner(t.Context(), module, wasm.DefaultLimits())
		require.ErrorContains(t, err, `imports "host_call" from "env"`)
	})

	t.Run("module asking for memory of the host", func(t *testing.T) {
		// (module (import "env" "mem" (memory 1)) (func (export "_start")))
		module := []byte("\x00asm\x01\x00\x00\x00" +
			"\x01\x04\x01\x60\x00\x00" + // type section: func() -> ()
			"\x02\x0c\x01\x03env\x03mem\x02\x00\x01" + // import section
			"\x03\x02\x01\x00" + // function section
			"\x07\x0a\x01\x06_start\x00\x00" + // export section
			"\x0a\x04\x01\x02\x00\x0b") // code section
		_, err := wasm.NewRunner(t.Context(), module, wasm.DefaultLimits())
		require.ErrorContains(t, err, `imports memory "mem" from "env"`)
	})

	t.Run("invalid limits", func(t *testing.T) {
		limits := wasm.DefaultLimits()
		limits.Timeout = 0
		_, err := wasm.NewRunner(t.Context(), nil, limits)
		require.ErrorContains(t, err, "timeout must be positive")
	})
}

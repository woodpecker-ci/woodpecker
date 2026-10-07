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

// Package wasm runs extensions that are compiled to WebAssembly inside of the server process.
//
// An extension is a WASI (preview 1) command module, it is the in-process counterpart of an HTTP extension:
// what an HTTP extension gets as request body is written to stdin, what it would send as response body is read from stdout.
//
// The module is treated as untrusted code. It gets no filesystem, no environment variables, no network and no real clock.
// Every call runs in a fresh instance, so nothing can leak from one call into the next one.
// Memory, run time, output size and the number of parallel calls are limited.
package wasm

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

const (
	wasmPageSize  = 64 << 10
	startFunction = "_start"

	defaultMemoryBytes = 256 << 20
	defaultTimeout     = 10 * time.Second
	defaultStdoutBytes = 10 << 20
	defaultStderrBytes = 8 << 10
	defaultConcurrency = 4
)

// compilationCache makes sure the same module is only compiled once per process.
var compilationCache = wazero.NewCompilationCache()

// ErrOutputTooLarge is returned if a module writes more to stdout than it is allowed to.
var ErrOutputTooLarge = errors.New("wasm module exceeded the output limit")

// ExitError is returned if a module stops with a non zero exit code.
type ExitError struct {
	Code   uint32
	Stderr string
}

func (e *ExitError) Error() string {
	if e.Stderr == "" {
		return fmt.Sprintf("wasm module exited with code %d", e.Code)
	}
	return fmt.Sprintf("wasm module exited with code %d: %s", e.Code, e.Stderr)
}

// Limits restricts what a single call of a module can consume.
type Limits struct {
	// MemoryBytes is the maximum linear memory of one instance, rounded down to full wasm pages.
	MemoryBytes uint64
	// Timeout is the maximum run time of one call.
	Timeout time.Duration
	// StdoutBytes is the maximum size of the output, exceeding it fails the call.
	StdoutBytes int
	// StderrBytes is the maximum size kept from stderr, everything above is dropped.
	StderrBytes int
	// Concurrency is the maximum number of calls running at the same time, further calls wait.
	Concurrency int
}

// DefaultLimits returns the limits used if nothing else is requested.
func DefaultLimits() Limits {
	return Limits{
		MemoryBytes: defaultMemoryBytes,
		Timeout:     defaultTimeout,
		StdoutBytes: defaultStdoutBytes,
		StderrBytes: defaultStderrBytes,
		Concurrency: defaultConcurrency,
	}
}

func (l Limits) validate() error {
	switch {
	case l.MemoryBytes < wasmPageSize || l.MemoryBytes > 4<<30:
		return fmt.Errorf("memory limit must be between one wasm page and 4GiB, got %d bytes", l.MemoryBytes)
	case l.Timeout <= 0:
		return errors.New("timeout must be positive")
	case l.StdoutBytes <= 0 || l.StderrBytes <= 0:
		return errors.New("output limits must be positive")
	case l.Concurrency <= 0:
		return errors.New("concurrency must be positive")
	}
	return nil
}

// Runner executes one compiled module.
type Runner struct {
	runtime  wazero.Runtime
	compiled wazero.CompiledModule
	limits   Limits
	slots    chan struct{}
}

// NewRunner compiles the module and checks that it only asks for what the sandbox offers.
// Compiling is expensive, so create a runner once and reuse it.
func NewRunner(ctx context.Context, binary []byte, limits Limits) (_ *Runner, err error) {
	if err := limits.validate(); err != nil {
		return nil, err
	}

	runtime := wazero.NewRuntimeConfig().
		WithCompilationCache(compilationCache).
		WithMemoryLimitPages(uint32(limits.MemoryBytes / wasmPageSize)).
		WithCloseOnContextDone(true)
	r := wazero.NewRuntimeWithConfig(ctx, runtime)
	defer func() {
		if err != nil {
			_ = r.Close(ctx)
		}
	}()

	if _, err := wasi_snapshot_preview1.Instantiate(ctx, r); err != nil {
		return nil, fmt.Errorf("could not set up wasi: %w", err)
	}

	compiled, err := r.CompileModule(ctx, binary)
	if err != nil {
		return nil, fmt.Errorf("could not compile wasm module: %w", err)
	}

	if _, ok := compiled.ExportedFunctions()[startFunction]; !ok {
		return nil, fmt.Errorf("wasm module is not a wasi command, it does not export %q", startFunction)
	}
	for _, fn := range compiled.ImportedFunctions() {
		if module, name, _ := fn.Import(); module != wasi_snapshot_preview1.ModuleName {
			return nil, fmt.Errorf("wasm module imports %q from %q, only %q is available", name, module, wasi_snapshot_preview1.ModuleName)
		}
	}
	for _, memory := range compiled.ImportedMemories() {
		module, name, _ := memory.Import()
		return nil, fmt.Errorf("wasm module imports memory %q from %q, it has to bring its own", name, module)
	}

	return &Runner{
		runtime:  r,
		compiled: compiled,
		limits:   limits,
		slots:    make(chan struct{}, limits.Concurrency),
	}, nil
}

// Run starts a fresh instance of the module with the given arguments, feeds it stdin and returns what it wrote to stdout.
func (r *Runner) Run(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, r.limits.Timeout)
	defer cancel()

	select {
	case r.slots <- struct{}{}:
		defer func() { <-r.slots }()
	case <-ctx.Done():
		return nil, fmt.Errorf("gave up waiting for a free wasm slot: %w", ctx.Err())
	}

	stdout := &cappedBuffer{limit: r.limits.StdoutBytes, strict: true}
	stderr := &cappedBuffer{limit: r.limits.StderrBytes}

	config := wazero.NewModuleConfig().
		WithName(""). // anonymous, so instances can run in parallel
		WithArgs(args...).
		WithStdin(bytes.NewReader(stdin)).
		WithStdout(stdout).
		WithStderr(stderr).
		WithRandSource(rand.Reader)

	module, err := r.runtime.InstantiateModule(ctx, r.compiled, config)
	if module != nil {
		defer module.Close(ctx)
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("wasm module was stopped: %w", ctx.Err())
	}
	if stdout.exceeded {
		return nil, ErrOutputTooLarge
	}

	// a wasi command ends by calling proc_exit, which is reported as error even for exit code 0
	var exit *sys.ExitError
	if err != nil && !errors.As(err, &exit) {
		return nil, fmt.Errorf("wasm module crashed: %w", err)
	}
	if exit != nil && exit.ExitCode() != 0 {
		return nil, &ExitError{Code: exit.ExitCode(), Stderr: strings.TrimSpace(stderr.String())}
	}

	return stdout.Bytes(), nil
}

// Close releases the compiled module.
func (r *Runner) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}

// cappedBuffer keeps at most limit bytes and remembers if more was written.
// A strict buffer fails the write that exceeds the limit, otherwise the surplus is dropped silently.
type cappedBuffer struct {
	bytes.Buffer
	limit    int
	strict   bool
	exceeded bool
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	free := b.limit - b.Len()
	if len(p) <= free {
		return b.Buffer.Write(p)
	}

	b.exceeded = true
	b.Buffer.Write(p[:free])
	if b.strict {
		return free, ErrOutputTooLarge
	}
	return len(p), nil
}

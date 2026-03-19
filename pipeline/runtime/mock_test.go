//go:build test

// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runtime

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/urfave/cli/v3"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
)

// compile-time interface checks.
var (
	_ backend.Backend = (*mockEngine)(nil)
	_ tracing.Tracer  = (*mockTracer)(nil)
)

// ---------------------------------------------------------------------------
// mockEngine — only for edge cases that the dummy backend cannot simulate
// (e.g. returning nil *State from WaitStep, injecting DestroyStep errors,
// returning context.Canceled from specific methods).
// Prefer the dummy backend for all other tests.
// ---------------------------------------------------------------------------

type mockEngine struct {
	setupWorkflowFn   func(context.Context, *backend.Config, string) error
	destroyWorkflowFn func(context.Context, *backend.Config, string) error
	startStepFn       func(context.Context, *backend.Step, string) error
	waitStepFn        func(context.Context, *backend.Step, string) (*backend.State, error)
	tailStepFn        func(context.Context, *backend.Step, string) (io.ReadCloser, error)
	destroyStepFn     func(context.Context, *backend.Step, string) error
}

func (m *mockEngine) Name() string                                        { return "mock" }
func (m *mockEngine) IsAvailable(_ context.Context) bool                   { return true }
func (m *mockEngine) Flags() []cli.Flag                                    { return nil }
func (m *mockEngine) Load(_ context.Context) (*backend.BackendInfo, error) { return nil, nil }

func (m *mockEngine) SetupWorkflow(ctx context.Context, conf *backend.Config, id string) error {
	if m.setupWorkflowFn != nil {
		return m.setupWorkflowFn(ctx, conf, id)
	}
	return nil
}

func (m *mockEngine) DestroyWorkflow(ctx context.Context, conf *backend.Config, id string) error {
	if m.destroyWorkflowFn != nil {
		return m.destroyWorkflowFn(ctx, conf, id)
	}
	return nil
}

func (m *mockEngine) StartStep(ctx context.Context, step *backend.Step, id string) error {
	if m.startStepFn != nil {
		return m.startStepFn(ctx, step, id)
	}
	return nil
}

func (m *mockEngine) WaitStep(ctx context.Context, step *backend.Step, id string) (*backend.State, error) {
	if m.waitStepFn != nil {
		return m.waitStepFn(ctx, step, id)
	}
	return &backend.State{Exited: true, ExitCode: 0}, nil
}

func (m *mockEngine) TailStep(ctx context.Context, step *backend.Step, id string) (io.ReadCloser, error) {
	if m.tailStepFn != nil {
		return m.tailStepFn(ctx, step, id)
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *mockEngine) DestroyStep(ctx context.Context, step *backend.Step, id string) error {
	if m.destroyStepFn != nil {
		return m.destroyStepFn(ctx, step, id)
	}
	return nil
}

// ---------------------------------------------------------------------------
// mockTracer — records all Trace calls for assertions.
// ---------------------------------------------------------------------------

type mockTracer struct {
	mu    sync.Mutex
	calls []state.State
	fn    func(*state.State) error
}

func (m *mockTracer) Trace(s *state.State) error {
	m.mu.Lock()
	m.calls = append(m.calls, *s)
	m.mu.Unlock()
	if m.fn != nil {
		return m.fn(s)
	}
	return nil
}

func (m *mockTracer) getCalls() []state.State {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]state.State, len(m.calls))
	copy(out, m.calls)
	return out
}

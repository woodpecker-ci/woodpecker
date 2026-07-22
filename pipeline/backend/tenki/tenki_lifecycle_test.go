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

package tenki

import (
	"context"
	"errors"
	"io"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TenkiCloud/tenki-sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// Fakes implementing the SDK seams from client.go.

type fakeWriteCloser struct{ closed atomic.Bool }

func (w *fakeWriteCloser) Write(p []byte) (int, error) { return len(p), nil }
func (w *fakeWriteCloser) Close() error                { w.closed.Store(true); return nil }

type fakeRunHandle struct {
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
	result  *sandbox.Result
	waitErr error
	release chan struct{} // if non-nil, Wait blocks until it is closed
	kills   atomic.Int32
}

func (f *fakeRunHandle) Stdin() io.WriteCloser { return f.stdin }
func (f *fakeRunHandle) Stdout() io.Reader     { return f.stdout }
func (f *fakeRunHandle) Stderr() io.Reader     { return f.stderr }

func (f *fakeRunHandle) Wait() (*sandbox.Result, error) {
	if f.release != nil {
		<-f.release
	}
	return f.result, f.waitErr
}

func (f *fakeRunHandle) Kill() error {
	f.kills.Add(1)
	return nil
}

type fakeCommand struct {
	handle    sandboxRunHandle
	streamErr error
	execRes   *sandbox.Result
	execErr   error
}

func (c *fakeCommand) Stream(_ context.Context) (sandboxRunHandle, error) {
	if c.streamErr != nil {
		return nil, c.streamErr
	}
	return c.handle, nil
}

func (c *fakeCommand) Exec(_ context.Context) (*sandbox.Result, error) {
	if c.execErr != nil {
		return nil, c.execErr
	}
	if c.execRes != nil {
		return c.execRes, nil
	}
	return &sandbox.Result{Status: sandbox.CommandStatusSucceeded}, nil
}

type fakeSession struct {
	id, name string
	cmd      *fakeCommand
	closeErr error
	closes   atomic.Int32
	argvs    [][]string
}

func (s *fakeSession) ID() string   { return s.id }
func (s *fakeSession) Name() string { return s.name }

func (s *fakeSession) Command(argv []string, _ ...sandbox.RunOptions) sandboxCommand {
	s.argvs = append(s.argvs, argv)
	if s.cmd == nil {
		s.cmd = &fakeCommand{handle: &fakeRunHandle{result: &sandbox.Result{Status: sandbox.CommandStatusSucceeded}}}
	}
	return s.cmd
}

func (s *fakeSession) CloseIfOpen(_ context.Context) error {
	s.closes.Add(1)
	return s.closeErr
}

type fakeClient struct {
	createSession sandboxSession
	createErr     error
	listSessions  []sandboxSession
	listErr       error
}

func (c *fakeClient) CreateAndWait(_ context.Context, _ time.Duration, _ ...sandbox.CreateOption) (sandboxSession, error) {
	if c.createErr != nil {
		return nil, c.createErr
	}
	return c.createSession, nil
}

func (c *fakeClient) ListProjectSandboxes(_ context.Context, _ string, _ ...sandbox.ListOption) ([]sandboxSession, error) {
	return c.listSessions, c.listErr
}

func (c *fakeClient) WhoAmI(_ context.Context) (*sandbox.Identity, error) {
	return nil, errors.New("not implemented")
}

// Lifecycle tests.

func TestReapOrphanSandbox(t *testing.T) {
	match := &fakeSession{id: "sb-match", name: sandboxName("task-1")}
	other := &fakeSession{id: "sb-other", name: "unrelated"}
	e := &tenki{client: &fakeClient{listSessions: []sandboxSession{other, match}}}
	e.config.projectID = "proj"

	require.NoError(t, e.reapOrphanSandbox(context.Background(), "task-1"))

	assert.Equal(t, int32(1), match.closes.Load(), "matching orphan should be closed")
	assert.Equal(t, int32(0), other.closes.Load(), "unrelated sandbox must not be touched")
}

func TestSetupWorkflowReapsOnCreateFailure(t *testing.T) {
	orphan := &fakeSession{id: "sb-orphan", name: sandboxName("task-x")}
	e := &tenki{client: &fakeClient{
		createErr:    errors.New("boom: readiness timed out"),
		listSessions: []sandboxSession{orphan},
	}}
	e.config.projectID = "proj"

	err := e.SetupWorkflow(context.Background(), &backend_types.Config{}, "task-x")

	require.Error(t, err)
	assert.Equal(t, int32(1), orphan.closes.Load(), "orphan created before failure should be reaped")
	_, stateErr := e.getWorkflowState("task-x")
	assert.ErrorIs(t, stateErr, ErrWorkflowStateNotFound, "no workflow state should be stored on failure")
}

func TestWaitStepCancellation(t *testing.T) {
	handle := &fakeRunHandle{release: make(chan struct{})}
	defer close(handle.release) // let the Wait goroutine finish

	e := &tenki{}
	ws := &workflowState{session: &fakeSession{}}
	ws.stepState.Store("step-1", &stepState{handle: handle})
	e.workflows.Store("task-1", ws)

	ctx, cancel := context.WithTimeout(context.Background(), 0) // already expired
	defer cancel()

	state, err := e.WaitStep(ctx, &backend_types.Step{UUID: "step-1"}, "task-1")
	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Nil(t, state)
	assert.GreaterOrEqual(t, handle.kills.Load(), int32(1), "cancellation should signal the process")
}

func TestWaitStepStatusMapping(t *testing.T) {
	run := func(t *testing.T, status sandbox.CommandStatus, exit int32) *backend_types.State {
		t.Helper()
		handle := &fakeRunHandle{result: &sandbox.Result{Status: status, ExitCode: exit}}
		e := &tenki{}
		ws := &workflowState{session: &fakeSession{}}
		ws.stepState.Store("s", &stepState{handle: handle})
		e.workflows.Store("t", ws)
		state, err := e.WaitStep(context.Background(), &backend_types.Step{UUID: "s"}, "t")
		require.NoError(t, err)
		return state
	}

	t.Run("success keeps exit code", func(t *testing.T) {
		st := run(t, sandbox.CommandStatusSucceeded, 0)
		assert.Equal(t, 0, st.ExitCode)
		assert.NoError(t, st.Error)
	})

	t.Run("killed with exit 0 is coerced to failure", func(t *testing.T) {
		st := run(t, sandbox.CommandStatusFailed, 0)
		assert.NotEqual(t, 0, st.ExitCode)
	})

	t.Run("timed out sets error", func(t *testing.T) {
		st := run(t, sandbox.CommandStatusTimedOut, 0)
		assert.NotEqual(t, 0, st.ExitCode)
		assert.Error(t, st.Error)
	})
}

func TestDestroyStepKillsAndIsIdempotent(t *testing.T) {
	handle := &fakeRunHandle{}
	e := &tenki{}
	ws := &workflowState{session: &fakeSession{}}
	ws.stepState.Store("s", &stepState{handle: handle, output: io.NopCloser(strings.NewReader(""))})
	e.workflows.Store("t", ws)

	require.NoError(t, e.DestroyStep(context.Background(), &backend_types.Step{UUID: "s"}, "t"))
	assert.Equal(t, int32(1), handle.kills.Load())

	// second call must be a no-op, not an error
	require.NoError(t, e.DestroyStep(context.Background(), &backend_types.Step{UUID: "s"}, "t"))
	assert.Equal(t, int32(1), handle.kills.Load())
}

func TestDestroyWorkflowTeardownFailure(t *testing.T) {
	handle := &fakeRunHandle{}
	sess := &fakeSession{name: sandboxName("t"), closeErr: errors.New("terminate failed")}
	orphan := &fakeSession{id: "sb", name: sandboxName("t")}
	e := &tenki{client: &fakeClient{listSessions: []sandboxSession{orphan}}}
	ws := &workflowState{session: sess}
	ws.stepState.Store("s", &stepState{handle: handle, output: io.NopCloser(strings.NewReader(""))})
	e.workflows.Store("t", ws)

	// A failing teardown must not surface as an error, and the sandbox must be
	// retried out-of-band (reaped by name) rather than left running.
	require.NoError(t, e.DestroyWorkflow(context.Background(), &backend_types.Config{}, "t"))
	assert.Equal(t, int32(1), sess.closes.Load(), "direct close attempted")
	assert.Equal(t, int32(1), orphan.closes.Load(), "failed close retried via reaper")
	assert.GreaterOrEqual(t, handle.kills.Load(), int32(1))
	_, err := e.getWorkflowState("t")
	assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
}

func TestDestroyWorkflowKeepsStateWhenTerminationUnconfirmed(t *testing.T) {
	handle := &fakeRunHandle{}
	sess := &fakeSession{name: sandboxName("t"), closeErr: errors.New("close failed")}
	// the reaper retry also fails (cannot even list), so termination is unconfirmed
	e := &tenki{client: &fakeClient{listErr: errors.New("list failed")}}
	ws := &workflowState{session: sess}
	ws.stepState.Store("s", &stepState{handle: handle, output: io.NopCloser(strings.NewReader(""))})
	e.workflows.Store("t", ws)

	err := e.DestroyWorkflow(context.Background(), &backend_types.Config{}, "t")
	require.Error(t, err, "unconfirmed teardown must not be reported as success")

	// state is kept so the sandbox can still be terminated on a later attempt
	_, stateErr := e.getWorkflowState("t")
	require.NoError(t, stateErr, "workflow state must be retained until termination is confirmed")
}

func TestStartStepRejectsEmptyCommands(t *testing.T) {
	e := &tenki{}
	e.workflows.Store("t", &workflowState{session: &fakeSession{}})

	err := e.StartStep(context.Background(), &backend_types.Step{UUID: "s", Name: "clone", Type: backend_types.StepTypeClone}, "t")
	assert.ErrorIs(t, err, ErrNoCommands)
}

func TestStartStepUsesGeneratedScript(t *testing.T) {
	stdin := &fakeWriteCloser{}
	sess := &fakeSession{cmd: &fakeCommand{handle: &fakeRunHandle{
		stdin:  stdin,
		stdout: strings.NewReader(""),
		stderr: strings.NewReader(""),
		result: &sandbox.Result{Status: sandbox.CommandStatusSucceeded},
	}}}
	e := &tenki{}
	e.workflows.Store("t", &workflowState{session: sess})

	err := e.StartStep(context.Background(), &backend_types.Step{
		UUID:       "s",
		Name:       "build",
		Type:       backend_types.StepTypeCommands,
		Commands:   []string{"echo hi"},
		WorkingDir: "/woodpecker/src",
	}, "t")
	require.NoError(t, err)

	// the step must run through the shared CI_SCRIPT entrypoint (/bin/sh -e),
	// not a raw joined `sh -c`.
	require.NotEmpty(t, sess.argvs)
	last := sess.argvs[len(sess.argvs)-1]
	assert.Equal(t, []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, last)

	// stdin must be closed so the SDK's stdin pump goroutine does not leak.
	assert.True(t, stdin.closed.Load(), "step stdin should be closed")
}

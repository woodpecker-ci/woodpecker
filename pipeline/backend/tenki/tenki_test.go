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
	"io"
	"strings"
	"testing"
	"time"

	"github.com/TenkiCloud/tenki-sdk-go/sandbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestName(t *testing.T) {
	assert.Equal(t, "tenki", New().Name())
}

func TestFlags(t *testing.T) {
	assert.NotEmpty(t, New().Flags())
}

func TestIsAvailable(t *testing.T) {
	backend := New()

	t.Run("no cli context", func(t *testing.T) {
		assert.False(t, backend.IsAvailable(context.Background()))
	})

	t.Run("api key set", func(t *testing.T) {
		cmd := &cli.Command{Flags: []cli.Flag{
			&cli.StringFlag{Name: "backend-tenki-api-key", Value: "tk_test"},
		}}
		ctx := context.WithValue(context.Background(), backend_types.CliCommand, cmd)
		assert.True(t, backend.IsAvailable(ctx))
	})

	t.Run("api key empty", func(t *testing.T) {
		cmd := &cli.Command{Flags: []cli.Flag{
			&cli.StringFlag{Name: "backend-tenki-api-key"},
		}}
		ctx := context.WithValue(context.Background(), backend_types.CliCommand, cmd)
		assert.False(t, backend.IsAvailable(ctx))
	})
}

func TestConfigFromCli(t *testing.T) {
	t.Run("missing api key", func(t *testing.T) {
		cmd := &cli.Command{Flags: []cli.Flag{
			&cli.StringFlag{Name: "backend-tenki-api-key"},
		}}
		_, err := configFromCli(cmd)
		assert.ErrorIs(t, err, ErrMissingAPIKey)
	})

	t.Run("valid config", func(t *testing.T) {
		cmd := &cli.Command{Flags: []cli.Flag{
			&cli.StringFlag{Name: "backend-tenki-api-key", Value: "tk_test"},
			&cli.StringFlag{Name: "backend-tenki-endpoint", Value: "https://api.example.com"},
			&cli.StringFlag{Name: "backend-tenki-project-id", Value: "proj-1"},
			&cli.StringFlag{Name: "backend-tenki-workspace-id", Value: "ws-1"},
			&cli.BoolFlag{Name: "backend-tenki-allow-outbound", Value: true},
		}}
		conf, err := configFromCli(cmd)
		require.NoError(t, err)
		assert.Equal(t, "tk_test", conf.apiKey)
		assert.Equal(t, "https://api.example.com", conf.endpoint)
		assert.Equal(t, "proj-1", conf.projectID)
		assert.Equal(t, "ws-1", conf.workspaceID)
		assert.True(t, conf.allowOutbound)
	})
}

func TestSelectProject(t *testing.T) {
	workspaces := []sandbox.IdentityWorkspace{
		{ID: "ws-empty", Projects: nil},
		{ID: "ws-a", Projects: []sandbox.IdentityProject{{ID: "proj-a1"}, {ID: "proj-a2"}}},
		{ID: "ws-b", Projects: []sandbox.IdentityProject{{ID: "proj-b1"}}},
	}

	tests := []struct {
		name        string
		workspaces  []sandbox.IdentityWorkspace
		workspaceID string
		wantWS      string
		wantProject string
		wantOK      bool
	}{
		{
			name:        "first workspace with a project",
			workspaces:  workspaces,
			wantWS:      "ws-a",
			wantProject: "proj-a1",
			wantOK:      true,
		},
		{
			name:        "explicit workspace",
			workspaces:  workspaces,
			workspaceID: "ws-b",
			wantWS:      "ws-b",
			wantProject: "proj-b1",
			wantOK:      true,
		},
		{
			name:        "explicit workspace without projects",
			workspaces:  workspaces,
			workspaceID: "ws-empty",
			wantOK:      false,
		},
		{
			name:        "unknown workspace",
			workspaces:  workspaces,
			workspaceID: "ws-missing",
			wantOK:      false,
		},
		{
			name:       "no workspaces",
			workspaces: nil,
			wantOK:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ws, project, ok := selectProject(tt.workspaces, tt.workspaceID)
			assert.Equal(t, tt.wantOK, ok)
			assert.Equal(t, tt.wantWS, ws)
			assert.Equal(t, tt.wantProject, project)
		})
	}
}

func TestWorkflowStateNotFound(t *testing.T) {
	e, _ := New().(*tenki)

	_, err := e.getWorkflowState("missing")
	assert.ErrorIs(t, err, ErrWorkflowStateNotFound)

	_, err = e.getStepState("missing", "step")
	assert.ErrorIs(t, err, ErrWorkflowStateNotFound)
}

func TestStepStateRoundTrip(t *testing.T) {
	e, _ := New().(*tenki)

	ws := &workflowState{}
	e.workflows.Store("task-1", ws)

	// workflow exists, step does not yet
	_, err := e.getStepState("task-1", "step-1")
	assert.ErrorIs(t, err, ErrStepStateNotFound)

	got, err := e.getWorkflowState("task-1")
	require.NoError(t, err)
	assert.Same(t, ws, got)

	// store a step and read it back
	ss := &stepState{}
	ws.stepState.Store("step-1", ss)
	gotStep, err := e.getStepState("task-1", "step-1")
	require.NoError(t, err)
	assert.Same(t, ss, gotStep)
}

func TestSandboxName(t *testing.T) {
	assert.Equal(t, "woodpecker-abc123", sandboxName("abc123"))
}

func TestWorkflowMetadata(t *testing.T) {
	t.Run("no steps: only tracing keys", func(t *testing.T) {
		md := workflowMetadata(&backend_types.Config{}, "task-1")
		assert.Equal(t, "woodpecker", md["managed-by"])
		assert.Equal(t, "task-1", md["task-uuid"])
		_, hasRepo := md["repo"]
		assert.False(t, hasRepo)
	})

	t.Run("labels taken from first step that has any", func(t *testing.T) {
		conf := &backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{{Name: "a"}}},
			{Steps: []*backend_types.Step{{Name: "b", WorkflowLabels: map[string]string{"repo": "x/y"}}}},
		}}
		md := workflowMetadata(conf, "task-2")
		assert.Equal(t, "x/y", md["repo"])
		assert.Equal(t, "woodpecker", md["managed-by"])
		assert.Equal(t, "task-2", md["task-uuid"])
	})

	t.Run("tracing keys are not shadowed by user labels", func(t *testing.T) {
		conf := &backend_types.Config{Stages: []*backend_types.Stage{
			{Steps: []*backend_types.Step{{Name: "a", WorkflowLabels: map[string]string{
				"managed-by": "spoof",
				"task-uuid":  "spoof",
			}}}},
		}}
		md := workflowMetadata(conf, "real-task")
		assert.Equal(t, "woodpecker", md["managed-by"])
		assert.Equal(t, "real-task", md["task-uuid"])
	})
}

func TestMergeOutput(t *testing.T) {
	t.Run("merges stdout and stderr", func(t *testing.T) {
		h := &fakeRunHandle{
			stdout: strings.NewReader("out-line\n"),
			stderr: strings.NewReader("err-line\n"),
		}
		rc := mergeOutput(h)
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())

		got := string(data)
		assert.Contains(t, got, "out-line")
		assert.Contains(t, got, "err-line")
	})

	t.Run("tolerates a nil stream", func(t *testing.T) {
		h := &fakeRunHandle{stdout: strings.NewReader("only-out")}
		rc := mergeOutput(h)
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		require.NoError(t, rc.Close())
		assert.Equal(t, "only-out", string(data))
	})
}

func TestConfigDurations(t *testing.T) {
	// the documented defaults live in these constants
	assert.Equal(t, time.Minute*time.Duration(3), defaultCreateTimeout)
	assert.Equal(t, time.Hour, defaultMaxDuration)

	// configFromCli reads the duration flags
	cmd := &cli.Command{Flags: []cli.Flag{
		&cli.StringFlag{Name: "backend-tenki-api-key", Value: "tk_test"},
		&cli.DurationFlag{Name: "backend-tenki-create-timeout", Value: defaultCreateTimeout},
		&cli.DurationFlag{Name: "backend-tenki-max-duration", Value: defaultMaxDuration},
	}}
	conf, err := configFromCli(cmd)
	require.NoError(t, err)
	assert.Equal(t, defaultCreateTimeout, conf.createTimeout)
	assert.Equal(t, defaultMaxDuration, conf.maxDuration)
}

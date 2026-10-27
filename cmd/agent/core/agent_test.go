// Copyright 2023 Woodpecker Authors
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

package core

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	agent_rpc "go.woodpecker-ci.org/woodpecker/v3/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/local"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestStringSliceAddToMap(t *testing.T) {
	tests := []struct {
		name     string
		sl       []string
		m        map[string]string
		expected map[string]string
		err      bool
	}{
		{
			name: "add values to map",
			sl:   []string{"foo=bar", "baz=qux=nux"},
			m:    make(map[string]string),
			expected: map[string]string{
				"foo": "bar",
				"baz": "qux=nux",
			},
			err: false,
		},
		{
			name:     "empty slice",
			sl:       []string{},
			m:        make(map[string]string),
			expected: make(map[string]string),
			err:      false,
		},
		{
			name:     "missing value",
			sl:       []string{"foo", "baz=qux"},
			m:        make(map[string]string),
			expected: make(map[string]string),
			err:      true,
		},
		{
			name:     "empty string in slice",
			sl:       []string{"foo=bar", "", "baz=qux"},
			m:        make(map[string]string),
			expected: map[string]string{"foo": "bar", "baz": "qux"},
			err:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stringSliceAddToMap(tt.sl, tt.m)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.EqualValues(t, tt.expected, tt.m)
			}
		})
	}
}

func TestGrpcDialConfigUsesRegisteredKeepaliveFlags(t *testing.T) {
	t.Setenv("WOODPECKER_KEEPALIVE_TIME", "10m")
	t.Setenv("WOODPECKER_KEEPALIVE_TIMEOUT", "30s")

	app := GenApp(nil)
	var dialConfig agent_rpc.DialConfig
	app.Action = func(_ context.Context, c *cli.Command) error {
		dialConfig = grpcDialConfig(c, AgentConfig{}, false)
		return nil
	}
	assert.NoError(t, app.Run(t.Context(), []string{"woodpecker-agent"}))

	assert.Equal(t, 10*time.Minute, dialConfig.KeepaliveTime)
	assert.Equal(t, 30*time.Second, dialConfig.KeepaliveTimeout)
}

func TestAgentSecretNotInEnvironment(t *testing.T) {
	if _, err := exec.LookPath("env"); err != nil {
		t.Skip("env binary not found")
	}
	t.Setenv("WOODPECKER_AGENT_SECRET", "agent-secret-value")

	app := GenApp(nil)
	var token string
	app.Action = func(_ context.Context, c *cli.Command) error {
		token = c.String("grpc-token")
		return nil
	}
	require.NoError(t, app.Run(t.Context(), []string{"woodpecker-agent"}))

	assert.Equal(t, "agent-secret-value", token)
	_, found := os.LookupEnv("WOODPECKER_AGENT_SECRET")
	assert.False(t, found, "WOODPECKER_AGENT_SECRET still in agent environment")

	// a local backend step must not inherit the secret from the agent
	backend := local.New()
	step := &types.Step{UUID: "env", Name: "env", Type: types.StepTypePlugin, Image: "env"}
	require.NoError(t, backend.SetupWorkflow(t.Context(), &types.Config{}, "task"))
	defer func() { assert.NoError(t, backend.DestroyWorkflow(t.Context(), &types.Config{}, "task")) }()
	require.NoError(t, backend.StartStep(t.Context(), step, "task"))
	output, err := backend.TailStep(t.Context(), step, "task")
	require.NoError(t, err)
	stepEnv, err := io.ReadAll(output)
	require.NoError(t, err)
	_, err = backend.WaitStep(t.Context(), step, "task")
	require.NoError(t, err)

	assert.False(t, strings.Contains(string(stepEnv), "WOODPECKER_AGENT_SECRET="), "local step inherited WOODPECKER_AGENT_SECRET")
}

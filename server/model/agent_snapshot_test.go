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

package model

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAgentSnapshot(t *testing.T) {
	agent := &Agent{
		ID:           3,
		Name:         "builder-03",
		OwnerID:      7,
		OrgID:        5,
		Token:        "secret-token",
		Platform:     "linux/amd64",
		Backend:      "docker",
		Version:      "3.19.0",
		CustomLabels: map[string]string{"zone": "eu"},
		Filters:      map[string]string{"repo": "a/b"},
	}

	snapshot := NewAgentSnapshot(agent)

	assert.Equal(t, &AgentSnapshot{
		AgentID:      3,
		OrgID:        5,
		Name:         "builder-03",
		Platform:     "linux/amd64",
		Backend:      "docker",
		CustomLabels: map[string]string{"zone": "eu"},
	}, snapshot)

	// later changes of the agent must not change the snapshot
	agent.CustomLabels["zone"] = "us"
	assert.Equal(t, "eu", snapshot.CustomLabels["zone"])

	raw, err := json.Marshal(snapshot)
	require.NoError(t, err)
	assert.NotContains(t, string(raw), "secret-token")
}

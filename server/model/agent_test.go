// Copyright 2024 Woodpecker Authors
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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNewAgentToken(t *testing.T) {
	token1 := GenerateNewAgentToken()
	token2 := GenerateNewAgentToken()

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
	assert.Len(t, token1, 56)
}

func TestAgent_GetServerLabels(t *testing.T) {
	t.Run("EmptyAgent", func(t *testing.T) {
		agent := &Agent{}
		filters, err := agent.GetServerLabels()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterOrgID: "0",
		}, filters)
	})

	t.Run("GlobalAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID: IDNotSet,
		}
		filters, err := agent.GetServerLabels()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterOrgID: "*",
		}, filters)
	})

	t.Run("OrgAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID: 123,
		}
		filters, err := agent.GetServerLabels()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterOrgID: "123",
		}, filters)
	})
}

func TestAgent_CanAccessRepo(t *testing.T) {
	repo := &Repo{ID: 123, OrgID: 12}
	otherRepo := &Repo{ID: 456, OrgID: 45}

	t.Run("EmptyAgent", func(t *testing.T) {
		agent := &Agent{}
		assert.False(t, agent.CanAccessRepo(repo))
	})

	t.Run("GlobalAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID: IDNotSet,
		}

		assert.True(t, agent.CanAccessRepo(repo))
	})

	t.Run("OrgAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID: 12,
		}
		assert.True(t, agent.CanAccessRepo(repo))
		assert.False(t, agent.CanAccessRepo(otherRepo))
	})
}

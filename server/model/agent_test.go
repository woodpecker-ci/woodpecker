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

func TestAgent_GetServerFilters(t *testing.T) {
	t.Run("EmptyAgent", func(t *testing.T) {
		agent := &Agent{}
		filters, err := agent.GetServerFilters()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterRepoID: "0",
			agentFilterOrgID:  "0",
		}, filters)
	})

	t.Run("SystemAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  SystemAgentOwnerID,
			RepoID: SystemAgentOwnerID,
		}
		filters, err := agent.GetServerFilters()
		assert.NoError(t, err)
		assert.Empty(t, filters)
	})

	t.Run("OrgAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  123,
			RepoID: SystemAgentOwnerID,
		}
		filters, err := agent.GetServerFilters()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterOrgID: "123",
		}, filters)
	})

	t.Run("RepoAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  SystemAgentOwnerID,
			RepoID: 456,
		}
		filters, err := agent.GetServerFilters()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterRepoID: "456",
		}, filters)
	})

	t.Run("OrgAndRepoAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  123,
			RepoID: 456,
		}
		filters, err := agent.GetServerFilters()
		assert.NoError(t, err)
		assert.Equal(t, map[string]string{
			agentFilterOrgID:  "123",
			agentFilterRepoID: "456",
		}, filters)
	})
}

func TestAgent_CanAccessRepo(t *testing.T) {
	t.Run("SystemAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  SystemAgentOwnerID,
			RepoID: SystemAgentOwnerID,
		}
		repo := &Repo{OrgID: 123}
		assert.True(t, agent.CanAccessRepo(repo))
	})

	t.Run("OrgAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  123,
			RepoID: SystemAgentOwnerID,
		}
		repo := &Repo{OrgID: 123}
		assert.True(t, agent.CanAccessRepo(repo))

		repoOtherOrg := &Repo{OrgID: 456}
		assert.False(t, agent.CanAccessRepo(repoOtherOrg))
	})

	t.Run("RepoAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  SystemAgentOwnerID,
			RepoID: 123,
		}
		repo := &Repo{OrgID: 123}
		assert.True(t, agent.CanAccessRepo(repo))

		repoOtherOrg := &Repo{OrgID: 456}
		assert.False(t, agent.CanAccessRepo(repoOtherOrg))
	})

	t.Run("OrgAndRepoAgent", func(t *testing.T) {
		agent := &Agent{
			OrgID:  123,
			RepoID: 456,
		}
		repoSameOrg := &Repo{OrgID: 123}
		assert.True(t, agent.CanAccessRepo(repoSameOrg))

		repoSameRepo := &Repo{OrgID: 456}
		assert.True(t, agent.CanAccessRepo(repoSameRepo))

		repoOther := &Repo{OrgID: 789}
		assert.False(t, agent.CanAccessRepo(repoOther))
	})
}

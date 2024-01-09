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

package datastore

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestAgentFindByToken(t *testing.T) {
	store, closer := newTestStore(t, new(model.Agent))
	defer closer()

	agent := &model.Agent{
		ID:    int64(1),
		Name:  "test",
		Token: "secret-token",
	}
	err := store.AgentCreate(agent)
	assert.NoError(t, err)

	_agent, err := store.AgentFindByToken(agent.Token)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, int64(1), _agent.ID)

	_agent, err = store.AgentFindByToken("")
	assert.ErrorIs(t, err, ErrNoTokenProvided)
	assert.Nil(t, _agent)
}

func TestAgentFindByID(t *testing.T) {
	store, closer := newTestStore(t, new(model.Agent))
	defer closer()

	agent := &model.Agent{
		ID:    int64(1),
		Name:  "test",
		Token: "secret-token",
	}
	err := store.AgentCreate(agent)
	assert.NoError(t, err)

	_agent, err := store.AgentFind(agent.ID)
	assert.NoError(t, err)
	assert.Equal(t, "secret-token", _agent.Token)
}

func TestAgentList(t *testing.T) {
	store, closer := newTestStore(t, new(model.Agent))
	defer closer()

	agent1 := &model.Agent{
		ID:    int64(1),
		Name:  "test-1",
		Token: "secret-token-1",
	}
	agent2 := &model.Agent{
		ID:    int64(2),
		Name:  "test-2",
		Token: "secret-token-2",
	}
	err := store.AgentCreate(agent1)
	assert.NoError(t, err)
	err = store.AgentCreate(agent2)
	assert.NoError(t, err)

	agents, err := store.AgentList(&model.ListOptions{All: true})
	assert.NoError(t, err)
	assert.Equal(t, 2, len(agents))

	agents, err = store.AgentList(&model.ListOptions{Page: 1, PerPage: 1})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(agents))
}

// Copyright 2022 Woodpecker Authors
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

package grpc

import (
	"context"
	"testing"
	"time"

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
)

func TestRegisterAgent(t *testing.T) {
	g := goblin.Goblin(t)
	g.Describe("When existing agent Name is empty", func() {
		g.It("Should update Name with hostname from metadata", func() {
			store := mocks_store.NewStore(t)
			storeAgent := new(model.Agent)
			storeAgent.ID = 1337
			updatedAgent := model.Agent{
				ID:          1337,
				Created:     0,
				Updated:     0,
				Name:        "hostname",
				OwnerID:     0,
				Token:       "",
				LastContact: 0,
				Platform:    "platform",
				Backend:     "backend",
				Capacity:    2,
				Version:     "version",
				NoSchedule:  false,
			}

			store.On("AgentFind", int64(1337)).Once().Return(storeAgent, nil)
			store.On("AgentUpdate", &updatedAgent).Once().Return(nil)
			grpc := RPC{
				store: store,
			}
			ctx := metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs("hostname", "hostname", "agent_id", "1337"),
			)
			agentID, err := grpc.RegisterAgent(ctx, rpc.AgentInfo{
				Version:  "version",
				Platform: "platform",
				Backend:  "backend",
				Capacity: 2,
			})
			if !assert.NoError(t, err) {
				return
			}

			assert.EqualValues(t, 1337, agentID)
		})
	})

	g.Describe("When existing agent hostname is present", func() {
		g.It("Should not update the hostname", func() {
			store := mocks_store.NewStore(t)
			storeAgent := new(model.Agent)
			storeAgent.ID = 1337
			storeAgent.Name = "originalHostname"
			updatedAgent := model.Agent{
				ID:          1337,
				Created:     0,
				Updated:     0,
				Name:        "originalHostname",
				OwnerID:     0,
				Token:       "",
				LastContact: 0,
				Platform:    "platform",
				Backend:     "backend",
				Capacity:    2,
				Version:     "version",
				NoSchedule:  false,
			}

			store.On("AgentFind", int64(1337)).Once().Return(storeAgent, nil)
			store.On("AgentUpdate", &updatedAgent).Once().Return(nil)
			grpc := RPC{
				store: store,
			}
			ctx := metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs("hostname", "newHostname", "agent_id", "1337"),
			)
			agentID, err := grpc.RegisterAgent(ctx, rpc.AgentInfo{
				Version:  "version",
				Platform: "platform",
				Backend:  "backend",
				Capacity: 2,
			})
			if !assert.NoError(t, err) {
				return
			}

			assert.EqualValues(t, 1337, agentID)
		})
	})
}

func TestUpdateAgentLastWork(t *testing.T) {
	t.Run("When last work was never updated it should update last work timestamp", func(t *testing.T) {
		agent := model.Agent{
			LastWork: 0,
		}
		store := mocks_store.NewStore(t)
		rpc := RPC{
			store: store,
		}
		store.On("AgentUpdate", mock.Anything).Once().Return(nil)

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.NotZero(t, agent.LastWork)
	})

	t.Run("When last work was updated over a minute ago it should update last work timestamp", func(t *testing.T) {
		lastWork := time.Now().Add(-time.Hour).Unix()
		agent := model.Agent{
			LastWork: lastWork,
		}
		store := mocks_store.NewStore(t)
		rpc := RPC{
			store: store,
		}
		store.On("AgentUpdate", mock.Anything).Once().Return(nil)

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.NotEqual(t, lastWork, agent.LastWork)
	})

	t.Run("When last work was updated in the last minute it should not update last work timestamp again", func(t *testing.T) {
		lastWork := time.Now().Add(-time.Second * 30).Unix()
		agent := model.Agent{
			LastWork: lastWork,
		}
		rpc := RPC{}

		err := rpc.updateAgentLastWork(&agent)
		assert.NoError(t, err)

		assert.Equal(t, lastWork, agent.LastWork)
	})
}

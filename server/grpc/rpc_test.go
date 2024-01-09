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

	"github.com/franela/goblin"
	"github.com/stretchr/testify/assert"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	mocks_store "go.woodpecker-ci.org/woodpecker/v2/server/store/mocks"
	"google.golang.org/grpc/metadata"
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
			rpc := RPC{
				store: store,
			}
			ctx := metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs("hostname", "hostname", "agent_id", "1337"),
			)
			capacity := int32(2)
			agentID, err := rpc.RegisterAgent(ctx, "platform", "backend", "version", capacity)
			if !assert.NoError(t, err) {
				t.Fail()
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
			rpc := RPC{
				store: store,
			}
			ctx := metadata.NewIncomingContext(
				context.Background(),
				metadata.Pairs("hostname", "newHostname", "agent_id", "1337"),
			)
			capacity := int32(2)
			agentID, err := rpc.RegisterAgent(ctx, "platform", "backend", "version", capacity)
			if !assert.NoError(t, err) {
				t.Fail()
			}

			assert.EqualValues(t, 1337, agentID)
		})
	})
}

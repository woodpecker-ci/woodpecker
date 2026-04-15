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

package rpc

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/rpc/proto"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// newAuthServer is a test helper that wires up a WoodpeckerAuthServer with the
// given master token and a mock store, then returns both so tests can set
// expectations before calling Auth / getAgent.
func newAuthServer(t *testing.T, masterToken string, store *store_mocks.MockStore) *WoodpeckerAuthServer {
	t.Helper()
	jwtManager := NewJWTManager("test-secret")
	return NewWoodpeckerAuthServer(jwtManager, masterToken, store)
}

func TestAuth(t *testing.T) {
	t.Parallel()

	t.Run("master token with agentID=-1 creates new system agent and returns access token", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentCreate", &model.Agent{
			OwnerID:  model.IDNotSet,
			OrgID:    model.IDNotSet,
			Token:    "master-secret",
			Capacity: -1,
		}).Return(nil).Once()

		srv := newAuthServer(t, "master-secret", store)
		resp, err := srv.Auth(t.Context(), &proto.AuthRequest{
			AgentId:    -1,
			AgentToken: "master-secret",
		})

		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.NotEmpty(t, resp.AccessToken)
		// The newly created agent has ID 0 (zero-value) because AgentCreate
		// doesn't set it in the mock – verify the token at least round-trips.
		claims, verifyErr := NewJWTManager("test-secret").Verify(resp.AccessToken)
		require.NoError(t, verifyErr)
		assert.Equal(t, resp.AgentId, claims.AgentID)
	})

	t.Run("master token with existing agentID returns access token for that agent", func(t *testing.T) {
		t.Parallel()

		existingAgent := &model.Agent{
			ID:      42,
			OrgID:   model.IDNotSet, // system agent
			OwnerID: model.IDNotSet,
		}

		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(42)).Return(existingAgent, nil).Once()

		srv := newAuthServer(t, "master-secret", store)
		resp, err := srv.Auth(t.Context(), &proto.AuthRequest{
			AgentId:    42,
			AgentToken: "master-secret",
		})

		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.EqualValues(t, 42, resp.AgentId)
		assert.NotEmpty(t, resp.AccessToken)
	})

	t.Run("individual agent token authenticates successfully", func(t *testing.T) {
		t.Parallel()

		agent := &model.Agent{ID: 7, Token: "individual-token"}

		store := store_mocks.NewMockStore(t)
		store.On("AgentFindByToken", "individual-token").Return(agent, nil).Once()

		// no master token configured
		srv := newAuthServer(t, "", store)
		resp, err := srv.Auth(t.Context(), &proto.AuthRequest{
			AgentId:    0,
			AgentToken: "individual-token",
		})

		require.NoError(t, err)
		assert.Equal(t, "ok", resp.Status)
		assert.EqualValues(t, 7, resp.AgentId)
	})

	t.Run("bad token returns error", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentFindByToken", "wrong-token").
			Return(nil, types.ErrRecordNotExist).Once()

		srv := newAuthServer(t, "", store)
		_, err := srv.Auth(t.Context(), &proto.AuthRequest{
			AgentToken: "wrong-token",
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "agent could not auth")
	})
}

func TestGetAgent(t *testing.T) {
	t.Parallel()

	t.Run("master token + agentID=-1 creates and returns a new system agent", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentCreate", &model.Agent{
			OwnerID:  model.IDNotSet,
			OrgID:    model.IDNotSet,
			Token:    "master",
			Capacity: -1,
		}).Return(nil).Once()

		srv := newAuthServer(t, "master", store)
		agent, err := srv.getAgent(-1, "master")

		require.NoError(t, err)
		require.NotNil(t, agent)
		assert.Equal(t, "master", agent.Token)
		assert.EqualValues(t, model.IDNotSet, agent.OrgID)
	})

	t.Run("master token + agentID=-1 propagates AgentCreate error", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentCreate", &model.Agent{
			OwnerID:  model.IDNotSet,
			OrgID:    model.IDNotSet,
			Token:    "master",
			Capacity: -1,
		}).Return(errors.New("db error")).Once()

		srv := newAuthServer(t, "master", store)
		_, err := srv.getAgent(-1, "master")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "db error")
	})

	t.Run("master token + existing agentID returns the stored agent", func(t *testing.T) {
		t.Parallel()

		systemAgent := &model.Agent{ID: 99, OrgID: model.IDNotSet, OwnerID: model.IDNotSet}

		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(99)).Return(systemAgent, nil).Once()

		srv := newAuthServer(t, "master", store)
		agent, err := srv.getAgent(99, "master")

		require.NoError(t, err)
		assert.Equal(t, int64(99), agent.ID)
	})

	t.Run("master token + agentID not found in database returns error", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(404)).Return(nil, types.ErrRecordNotExist).Once()

		srv := newAuthServer(t, "master", store)
		_, err := srv.getAgent(404, "master")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "AgentID not found in database")
	})

	t.Run("master token + agentID store returns unexpected error is propagated", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(1)).Return(nil, errors.New("connection reset")).Once()

		srv := newAuthServer(t, "master", store)
		_, err := srv.getAgent(1, "master")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "connection reset")
	})

	t.Run("master token + agentID that is not a system agent returns error", func(t *testing.T) {
		t.Parallel()

		// An agent with a non-IDNotSet OrgID is not a system agent.
		orgAgent := &model.Agent{ID: 5, OrgID: 100, OwnerID: model.IDNotSet}

		store := store_mocks.NewMockStore(t)
		store.On("AgentFind", int64(5)).Return(orgAgent, nil).Once()

		srv := newAuthServer(t, "master", store)
		_, err := srv.getAgent(5, "master")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a system agent")
	})

	t.Run("individual token auth succeeds when token is found", func(t *testing.T) {
		t.Parallel()

		agent := &model.Agent{ID: 3, Token: "ind-token"}
		store := store_mocks.NewMockStore(t)
		store.On("AgentFindByToken", "ind-token").Return(agent, nil).Once()

		// No master token set – falls straight to individual auth.
		srv := newAuthServer(t, "", store)
		got, err := srv.getAgent(0, "ind-token")

		require.NoError(t, err)
		assert.Equal(t, int64(3), got.ID)
	})

	t.Run("individual token not found returns wrapped error", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentFindByToken", "bad-token").
			Return(nil, types.ErrRecordNotExist).Once()

		srv := newAuthServer(t, "", store)
		_, err := srv.getAgent(0, "bad-token")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "individual agent not found by token")
	})

	t.Run("individual token store returns unexpected error is propagated", func(t *testing.T) {
		t.Parallel()

		store := store_mocks.NewMockStore(t)
		store.On("AgentFindByToken", "token").
			Return(nil, errors.New("timeout")).Once()

		srv := newAuthServer(t, "", store)
		_, err := srv.getAgent(0, "token")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("master token configured but wrong token falls through to individual auth", func(t *testing.T) {
		t.Parallel()

		agent := &model.Agent{ID: 8, Token: "ind-token"}
		store := store_mocks.NewMockStore(t)
		// master token is "master" but caller sends "ind-token" → individual path
		store.On("AgentFindByToken", "ind-token").Return(agent, nil).Once()

		srv := newAuthServer(t, "master", store)
		got, err := srv.getAgent(0, "ind-token")

		require.NoError(t, err)
		assert.Equal(t, int64(8), got.ID)
	})
}

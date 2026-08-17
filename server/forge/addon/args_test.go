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

package addon

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestModelUserRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("nil user maps to nil both ways", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, modelUserFromModel(nil))
		var mu *modelUser
		assert.Nil(t, mu.asModel())
	})

	t.Run("sensitive fields survive json round-trip", func(t *testing.T) {
		t.Parallel()
		orig := &model.User{
			ID:            7,
			Login:         "octocat",
			ForgeRemoteID: model.ForgeRemoteID("123"),
			AccessToken:   "access-tok",
			RefreshToken:  "refresh-tok",
			Expiry:        4242,
			Hash:          "secret-hash",
		}

		// the wrapper exists precisely because these fields are json:"-" on
		// model.User and would otherwise be lost across the addon boundary
		wrapped := modelUserFromModel(orig)
		data, err := json.Marshal(wrapped)
		require.NoError(t, err)

		var decoded modelUser
		require.NoError(t, json.Unmarshal(data, &decoded))

		got := decoded.asModel()
		require.NotNil(t, got)
		assert.Equal(t, model.ForgeRemoteID("123"), got.ForgeRemoteID)
		assert.Equal(t, "access-tok", got.AccessToken)
		assert.Equal(t, "refresh-tok", got.RefreshToken)
		assert.Equal(t, int64(4242), got.Expiry)
		assert.Equal(t, "secret-hash", got.Hash)
		assert.Equal(t, "octocat", got.Login)
	})
}

func TestModelRepoRoundTrip(t *testing.T) {
	t.Parallel()

	t.Run("nil repo maps to nil both ways", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, modelRepoFromModel(nil))
		var mr *modelRepo
		assert.Nil(t, mr.asModel())
	})

	t.Run("sensitive fields survive json round-trip", func(t *testing.T) {
		t.Parallel()
		perm := &model.Perm{Pull: true, Push: true, Admin: true}
		orig := &model.Repo{
			ID:       9,
			FullName: "octocat/hello",
			UserID:   42,
			Hash:     "repo-hash",
			Perm:     perm,
		}

		wrapped := modelRepoFromModel(orig)
		data, err := json.Marshal(wrapped)
		require.NoError(t, err)

		var decoded modelRepo
		require.NoError(t, json.Unmarshal(data, &decoded))

		got := decoded.asModel()
		require.NotNil(t, got)
		assert.Equal(t, int64(42), got.UserID)
		assert.Equal(t, "repo-hash", got.Hash)
		require.NotNil(t, got.Perm)
		assert.True(t, got.Perm.Admin)
		assert.Equal(t, "octocat/hello", got.FullName)
	})
}

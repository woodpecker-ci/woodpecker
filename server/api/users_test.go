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

//go:build test

package api

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestPostUser(t *testing.T) {
	s := newTestStore(t)

	t.Run("missing forge_id falls back to the default forge", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRequest(http.MethodPost, &model.User{Login: "carol"})(tc)

		PostUser(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())

		created := new(model.User)
		tc.decodeJSON(t, created)
		assert.EqualValues(t, defaultForgeID, created.ForgeID, "user must never be created with forge id 0")

		// the user's org must be forge-scoped as well
		org, err := s.OrgGet(created.OrgID)
		require.NoError(t, err)
		assert.EqualValues(t, defaultForgeID, org.ForgeID, "org must never be created with forge id 0")
	})

	t.Run("explicit forge_id is kept", func(t *testing.T) {
		tc := newTestContext(t, s)
		withRequest(http.MethodPost, &model.User{Login: "dave", ForgeID: 2})(tc)

		PostUser(tc.Ctx)

		require.Equal(t, http.StatusOK, tc.Recorder.Code, tc.Recorder.Body.String())

		created := new(model.User)
		tc.decodeJSON(t, created)
		assert.EqualValues(t, 2, created.ForgeID)
	})
}

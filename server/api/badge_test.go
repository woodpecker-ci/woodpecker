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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

// badgeFixture seeds a repo with the given visibility, one successful pipeline
// and a badge token.
func badgeFixture(t *testing.T, s store.Store, name string, visibility model.RepoVisibility) (*model.Repo, *model.Token) {
	t.Helper()
	repo := &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(name),
		Owner:         "owner",
		Name:          name,
		FullName:      "owner/" + name,
		Branch:        "main",
		Visibility:    visibility,
		IsActive:      true,
	}
	require.NoError(t, s.CreateRepo(repo))

	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Branch: "main",
		Event:  model.EventPush,
		Status: model.StatusSuccess,
	}
	require.NoError(t, s.CreatePipeline(pipeline))

	token := model.NewToken(repo.ID, model.TokenTypeBadge)
	require.NoError(t, s.TokenCreate(token))

	return repo, token
}

// getBadge calls GetBadge for the given repo id path param and query string.
func getBadge(t *testing.T, s store.Store, repoID, query string) *httptest.ResponseRecorder {
	t.Helper()
	tc := newTestContext(t, s)
	withParam("repo_id_or_owner", repoID)(tc)
	tc.Ctx.Request = httptest.NewRequest(http.MethodGet, "/?"+query, nil)
	GetBadge(tc.Ctx)
	return tc.Recorder
}

// getCC calls GetCC for the given repo id path param and query string.
func getCC(t *testing.T, s store.Store, repoID, query string) *httptest.ResponseRecorder {
	t.Helper()
	tc := newTestContext(t, s)
	withParam("repo_id_or_owner", repoID)(tc)
	tc.Ctx.Request = httptest.NewRequest(http.MethodGet, "/?"+query, nil)
	GetCC(tc.Ctx)
	return tc.Recorder
}

func TestGetBadgePublicRepo(t *testing.T) {
	s := newTestStore(t)
	repo, _ := badgeFixture(t, s, "public", model.VisibilityPublic)

	t.Run("without token", func(t *testing.T) {
		rec := getBadge(t, s, strItoa(repo.ID), "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "success")
	})

	// a token is not rejected either, it is simply not needed here
	t.Run("with wrong token", func(t *testing.T) {
		plain := getBadge(t, s, strItoa(repo.ID), "")
		rec := getBadge(t, s, strItoa(repo.ID), "token=nope")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, plain.Body.String(), rec.Body.String())
	})
}

// A token row whose value ended up empty must not turn into a badge that any
// caller can read by sending no token at all.
func TestGetBadgeRejectsEmptyStoredToken(t *testing.T) {
	repo := &model.Repo{ID: 42, Visibility: model.VisibilityPrivate}
	s := store_mocks.NewMockStore(t)
	s.EXPECT().GetRepo(repo.ID).Return(repo, nil)
	s.EXPECT().TokenFind(repo, model.TokenTypeBadge).Return(&model.Token{
		RepoID: repo.ID,
		Type:   model.TokenTypeBadge,
	}, nil)

	rec := getBadge(t, s, strItoa(repo.ID), "")

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), badgeUnavailableLabel)
}

func TestGetBadgeNonPublicRepoRequiresToken(t *testing.T) {
	s := newTestStore(t)
	private, privateToken := badgeFixture(t, s, "private", model.VisibilityPrivate)
	internal, _ := badgeFixture(t, s, "internal", model.VisibilityInternal)

	t.Run("valid token serves status", func(t *testing.T) {
		rec := getBadge(t, s, strItoa(private.ID), "token="+privateToken.Value)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "success")
	})

	t.Run("missing token hides status", func(t *testing.T) {
		rec := getBadge(t, s, strItoa(private.ID), "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), badgeUnavailableLabel)
		assert.NotContains(t, rec.Body.String(), "success")
	})

	t.Run("wrong token hides status", func(t *testing.T) {
		rec := getBadge(t, s, strItoa(private.ID), "token=nope")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), badgeUnavailableLabel)
	})

	t.Run("internal repos need a token too", func(t *testing.T) {
		rec := getBadge(t, s, strItoa(internal.ID), "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), badgeUnavailableLabel)
	})
}

// The response for a repo that may not be seen must be byte identical to the
// one for a repo that does not exist, otherwise the endpoint tells anonymous
// callers which repo ids and names exist.
func TestGetBadgeDoesNotLeakExistence(t *testing.T) {
	s := newTestStore(t)
	private, _ := badgeFixture(t, s, "private", model.VisibilityPrivate)

	missingByID := getBadge(t, s, "999999", "")
	missingByName := func() *httptest.ResponseRecorder {
		tc := newTestContext(t, s)
		withParam("repo_id_or_owner", "owner")(tc)
		withParam("repo_name", "does-not-exist")(tc)
		GetBadge(tc.Ctx)
		return tc.Recorder
	}()
	forbidden := getBadge(t, s, strItoa(private.ID), "")
	malformed := getBadge(t, s, "not-a-number", "")

	for name, rec := range map[string]*httptest.ResponseRecorder{
		"missing by id":   missingByID,
		"missing by name": missingByName,
		"malformed id":    malformed,
	} {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, forbidden.Code, rec.Code)
			assert.Equal(t, forbidden.Body.String(), rec.Body.String())
			assert.Equal(t, "image/svg+xml", rec.Header().Get("Content-Type"))
		})
	}
}

// Query parameters must not change the denied response either, they would
// otherwise expose branch, workflow or step names of a hidden repo.
func TestGetBadgeDeniedResponseIgnoresQuery(t *testing.T) {
	s := newTestStore(t)
	private, _ := badgeFixture(t, s, "private", model.VisibilityPrivate)

	plain := getBadge(t, s, strItoa(private.ID), "")
	withQuery := getBadge(t, s, strItoa(private.ID), "branch=main&events=push&workflow=build&step=test")

	assert.Equal(t, plain.Code, withQuery.Code)
	assert.Equal(t, plain.Body.String(), withQuery.Body.String())
}

func TestGetCCTokenGate(t *testing.T) {
	s := newTestStore(t)
	public, _ := badgeFixture(t, s, "public", model.VisibilityPublic)
	private, privateToken := badgeFixture(t, s, "private", model.VisibilityPrivate)

	t.Run("public repo is served", func(t *testing.T) {
		rec := getCC(t, s, strItoa(public.ID), "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "owner/public")
	})

	t.Run("valid token serves private repo", func(t *testing.T) {
		rec := getCC(t, s, strItoa(private.ID), "token="+privateToken.Value)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "owner/private")
	})

	// this is the reported leak: the cctray endpoint handed out the full name
	// of every repo, including private ones
	t.Run("missing token hides full name", func(t *testing.T) {
		rec := getCC(t, s, strItoa(private.ID), "")
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotContains(t, rec.Body.String(), "owner/private")
		assert.Contains(t, rec.Body.String(), badgeUnavailableLabel)
	})

	t.Run("hidden and missing repos look the same", func(t *testing.T) {
		hidden := getCC(t, s, strItoa(private.ID), "")
		missing := getCC(t, s, "999999", "")
		assert.Equal(t, hidden.Code, missing.Code)
		assert.Equal(t, hidden.Body.String(), missing.Body.String())
		assert.True(t, strings.HasPrefix(missing.Header().Get("Content-Type"), "application/xml"))
	})
}

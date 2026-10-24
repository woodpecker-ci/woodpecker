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
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
)

// strItoa formats an int64 id as a string for use as a URL path param.
func strItoa(i int64) string { return strconv.FormatInt(i, 10) }

// testContext bundles a gin test context, its response recorder and the
// backing in-memory store so individual API handlers can be exercised in
// isolation against a real (sqlite) database.
type testContext struct {
	Ctx      *gin.Context
	Recorder *httptest.ResponseRecorder
	Store    store.Store
	t        *testing.T
}

// newTestStore returns a fully-migrated in-memory sqlite store. Create one per
// top-level test and share it across subtests (each subtest seeds its own repo)
// so the schema migration only runs once.
func newTestStore(t *testing.T) store.Store {
	t.Helper()
	return datastore.NewTestStore(t)
}

// newTestContext builds a gin test context wired to the given store. Configure
// it by applying the option helpers (withUser, withRepo, withParam,
// withRequest) to the returned value.
//
// The store is reachable both via the returned struct and via
// store.FromContext(ctx) inside the handler under test.
func newTestContext(t *testing.T, s store.Store) *testContext {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Set("store", s)
	// Default request so handlers that read query/body never nil-panic.
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	return &testContext{Ctx: c, Recorder: rec, Store: s, t: t}
}

type ctxOption func(*testContext)

// withUser sets the authenticated user in the gin session.
func withUser(u *model.User) ctxOption {
	return func(tc *testContext) { tc.Ctx.Set("user", u) }
}

// withRepo sets the repo and its permissions in the gin session.
// Session.Repo reads "repo" and attaches Perm from the "perm" key, so both
// are set here.
func withRepo(r *model.Repo, perm *model.Perm) ctxOption {
	return func(tc *testContext) {
		tc.Ctx.Set("repo", r)
		if perm != nil {
			tc.Ctx.Set("perm", perm)
		}
	}
}

// withParam sets a gin URL path parameter (e.g. "cron", "repo_id", "secret").
// The key is intentionally a parameter so the same helper serves every
// endpoint group; it currently varies across the api test suite.
//
//nolint:unparam // key varies across endpoint test files sharing this helper
func withParam(key, value string) ctxOption {
	return func(tc *testContext) {
		tc.Ctx.Params = append(tc.Ctx.Params, gin.Param{Key: key, Value: value})
	}
}

// withRequest replaces the request with the given method and optional JSON
// body. A non-nil body is marshaled to JSON and Content-Type set.
func withRequest(method string, body any) ctxOption {
	return func(tc *testContext) {
		var reader io.Reader
		if body != nil {
			b, err := json.Marshal(body)
			require.NoError(tc.t, err, "marshal request body")
			reader = bytes.NewReader(b)
		}
		req := httptest.NewRequest(method, "/", reader)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		tc.Ctx.Request = req
	}
}

// withRawBody replaces the request body with raw bytes (e.g. malformed JSON)
// to exercise bind-error paths.
func withRawBody(method, contentType string, body []byte) ctxOption {
	return func(tc *testContext) {
		req := httptest.NewRequest(method, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", contentType)
		tc.Ctx.Request = req
	}
}

// decodeJSON unmarshals the recorder body into out.
func (tc *testContext) decodeJSON(t *testing.T, out any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(tc.Recorder.Body.Bytes(), out))
}

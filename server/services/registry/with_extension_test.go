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

package registry

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yaronf/httpsign"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestWithExtensionRegistryListPipeline(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name          string
		repoName      string
		dbRegs        []*model.Registry
		expected      []*model.Registry
		expectedError bool
	}{
		{
			name:     "Extension overrides base registry by name",
			repoName: "override-test",
			dbRegs: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "docker.io", Username: "shared", Password: "db-value"},
				{ID: 2, RepoID: 1, Address: "quay.io", Username: "db-only", Password: "only-in-db"},
			},
			expected: []*model.Registry{
				{ID: 2, RepoID: 1, Address: "quay.io", Username: "db-only", Password: "only-in-db"},
				{Address: "docker.io", Username: "shared", Password: "external-value"},
				{Address: "codeberg.org", Username: "ext-only", Password: "only-in-ext"},
			},
			expectedError: false,
		},
		{
			name:     "Extension returns 204 no registries",
			repoName: "no-content",
			dbRegs: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "quay.io", Username: "db-secret", Password: "db-value"},
			},
			expected: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "quay.io", Username: "db-secret", Password: "db-value"},
			},
			expectedError: false,
		},
		{
			name:     "Extension error falls back to base registries",
			repoName: "server-error",
			dbRegs: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "quay.io", Username: "db-secret", Password: "db-value"},
			},
			expected: []*model.Registry{
				{ID: 1, RepoID: 1, Address: "quay.io", Username: "db-secret", Password: "db-value"},
			},
			expectedError: false,
		},
	}

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err, "can't generate ed25519 keypair")

	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		// check signature
		pubKeyID := "woodpecker-ci-extensions"

		verifier, err := httpsign.NewEd25519Verifier(pubEd25519Key,
			httpsign.NewVerifyConfig(),
			httpsign.Headers("@request-target", "content-digest"))
		if err != nil {
			http.Error(w, "can't create verifier", http.StatusInternalServerError)
			return
		}

		err = httpsign.VerifyRequest(pubKeyID, *verifier, r)
		if err != nil {
			http.Error(w, "Invalid signature", http.StatusBadRequest)
			return
		}

		type incoming struct {
			Repo     *model.Repo     `json:"repo"`
			Pipeline *model.Pipeline `json:"pipeline"`
			Netrc    *model.Netrc    `json:"netrc"`
		}

		var req incoming
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "can't read body", http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, "Failed to parse JSON"+err.Error(), http.StatusBadRequest)
			return
		}

		switch req.Repo.Name {
		case "no-content":
			w.WriteHeader(http.StatusNoContent)
			return
		case "server-error":
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"registries": []*model.Registry{
				{Address: "docker.io", Username: "shared", Password: "external-value"},
				{Address: "codeberg.org", Username: "ext-only", Password: "only-in-ext"},
			},
		}))
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)

	httpExtension := NewHTTP(ts.URL, client, true)

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store_mocks.NewMockStore(t)
			mockStore.On("RegistryList", mock.Anything, true, mock.Anything).Return(tt.dbRegs, nil)

			combined := NewWithExtension(NewDB(mockStore), httpExtension)

			registries, err := combined.RegistryListPipeline(
				t.Context(),
				&model.Repo{ID: 1, Name: tt.repoName},
				&model.Pipeline{},
				nil,
			)
			if tt.expectedError {
				require.Error(t, err, "expected an error")
			} else {
				require.NoError(t, err, "error fetching registries")
			}

			assert.ElementsMatch(t, tt.expected, registries, "expected some other registries")
		})
	}
}

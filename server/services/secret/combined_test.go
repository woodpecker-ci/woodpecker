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

package secret_test

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
	"go.woodpecker-ci.org/woodpecker/v3/server/services/secret"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func TestCombinedSecretListPipeline(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name          string
		repoName      string
		dbSecrets     []*model.Secret
		expected      []*model.Secret
		expectedError bool
	}{
		{
			name:     "Extension overrides base secret by name",
			repoName: "override-test",
			dbSecrets: []*model.Secret{
				{ID: 1, RepoID: 1, Name: "shared", Value: "db-value"},
				{ID: 2, RepoID: 1, Name: "db-only", Value: "only-in-db"},
			},
			expected: []*model.Secret{
				{Name: "shared", Value: "external-value"},
				{Name: "ext-only", Value: "only-in-ext"},
				{ID: 2, RepoID: 1, Name: "db-only", Value: "only-in-db"},
			},
			expectedError: false,
		},
		{
			name:     "Extension returns 204 no secrets",
			repoName: "no-content",
			dbSecrets: []*model.Secret{
				{ID: 1, RepoID: 1, Name: "db-secret", Value: "db-value"},
			},
			expected: []*model.Secret{
				{ID: 1, RepoID: 1, Name: "db-secret", Value: "db-value"},
			},
			expectedError: false,
		},
		{
			name:     "Extension error falls back to base secrets",
			repoName: "server-error",
			dbSecrets: []*model.Secret{
				{ID: 1, RepoID: 1, Name: "db-secret", Value: "db-value"},
			},
			expected: []*model.Secret{
				{ID: 1, RepoID: 1, Name: "db-secret", Value: "db-value"},
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
			"secrets": []*model.Secret{
				{Name: "shared", Value: "external-value"},
				{Name: "ext-only", Value: "only-in-ext"},
			},
		}))
	}

	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)

	httpExtension := secret.NewHTTP(ts.URL, client, true)

	for _, tt := range testTable {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := store_mocks.NewMockStore(t)
			mockStore.On("SecretList", mock.Anything, true, mock.Anything).Return(tt.dbSecrets, nil)

			combined := secret.NewCombined(secret.NewDB(mockStore), httpExtension)

			secrets, err := combined.SecretListPipeline(
				t.Context(),
				&model.Repo{ID: 1, Name: tt.repoName},
				&model.Pipeline{},
				nil,
			)
			if tt.expectedError {
				require.Error(t, err, "expected an error")
			} else {
				require.NoError(t, err, "error fetching secrets")
			}

			assert.ElementsMatch(t, tt.expected, secrets, "expected some other secrets")
		})
	}
}

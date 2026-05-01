// Copyright 2023 Woodpecker Authors
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

package utils_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaronf/httpsign"

	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils"
)

func TestSignClient(t *testing.T) {
	pubKeyID := "woodpecker-ci-extensions"

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	body := []byte("{\"foo\":\"bar\"}")

	verifyHandler := func(w http.ResponseWriter, r *http.Request) {
		verifier, err := httpsign.NewEd25519Verifier(pubEd25519Key,
			httpsign.NewVerifyConfig(),
			httpsign.Headers("@request-target", "content-digest")) // The Content-Digest header will be auto-generated
		assert.NoError(t, err)

		err = httpsign.VerifyRequest(pubKeyID, *verifier, r)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
	}

	server := httptest.NewServer(http.HandlerFunc(verifyHandler))

	req, err := http.NewRequest("GET", server.URL+"/", bytes.NewBuffer(body))
	require.NoError(t, err)

	req.Header.Set("Date", time.Now().Format(time.RFC3339))
	req.Header.Set("Content-Type", "application/json")

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)

	rr, err := client.Do(req)
	assert.NoError(t, err)
	defer rr.Body.Close()

	assert.Equal(t, http.StatusOK, rr.StatusCode)
}

func TestRetry(t *testing.T) {
	_, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	numRetry := 0
	body := []byte("{\"foo\":\"bar\"}")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRetry++
		if numRetry >= 6 {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	client, err := utils.NewHTTPClient(privEd25519Key, "loopback")
	require.NoError(t, err)

	// first time: retry fails all the times
	_, err = client.Send(t.Context(), http.MethodGet, server.URL+"/", bytes.NewBuffer(body), nil)
	assert.Error(t, err)
	assert.Equal(t, 3, numRetry)

	// second time: retry succeeds after two failed times
	rr, err := client.Send(t.Context(), http.MethodGet, server.URL+"/", bytes.NewBuffer(body), nil)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, rr)
	assert.Equal(t, 6, numRetry)
}

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

package utils

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yaronf/httpsign"
)

func TestSignClient(t *testing.T) {
	pubKeyID := "woodpecker-ci-extensions"

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if !assert.NoError(t, err) {
		return
	}

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
	if !assert.NoError(t, err) {
		return
	}

	req.Header.Set("Date", time.Now().Format(time.RFC3339))
	req.Header.Set("Content-Type", "application/json")

	client, err := signClient(privEd25519Key)
	if !assert.NoError(t, err) {
		return
	}

	rr, err := client.Do(req)
	assert.NoError(t, err)
	defer rr.Body.Close()

	assert.Equal(t, http.StatusOK, rr.StatusCode)
}

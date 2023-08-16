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

	"github.com/go-fed/httpsig"
	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

func TestSign(t *testing.T) {
	pubKeyID := "woodpecker-ci-plugins"

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	body := []byte("{\"foo\":\"bar\"}")

	req, err := http.NewRequest("GET", "http://example.com", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	err = utils.SignHTTPRequest(privEd25519Key, req)
	if err != nil {
		t.Fatal(err)
	}

	VerifyHandler := func(w http.ResponseWriter, r *http.Request) {
		verifier, err := httpsig.NewVerifier(r)
		if err != nil {
			t.Fatal(err)
		}

		if verifier.KeyId() != pubKeyID {
			t.Fatalf("expected key ID %q, got %q", pubKeyID, verifier.KeyId())
		}

		if err := verifier.Verify(pubEd25519Key, httpsig.ED25519); err != nil {
			t.Fatal(err)
		}

		w.WriteHeader(http.StatusOK)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(VerifyHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

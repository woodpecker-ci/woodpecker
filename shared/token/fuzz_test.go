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

package token

import (
	"testing"
)

// FuzzParse feeds untrusted raw JWT strings into the token parser. The
// property checked is that parsing never panics and that a returned token
// always has one of the allowed types.
func FuzzParse(f *testing.F) {
	const secret = "fuzz-secret"

	allowedTypes := []Type{UserToken, SessToken, HookToken, CsrfToken, AgentToken, OAuthStateToken}

	// seed with a validly signed token for each type so the fuzzer can reach
	// the code paths behind signature verification
	for _, tokenType := range allowedTypes {
		signed, err := New(tokenType).Sign(secret)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(signed)
	}
	f.Add("eyJhbGciOiJIUzI1NiJ9.e30.")
	f.Add("not.a.jwt")

	f.Fuzz(func(t *testing.T, raw string) {
		parsed, err := Parse(allowedTypes, raw, func(*Token) (string, error) {
			return secret, nil
		})
		if err != nil {
			return
		}
		found := false
		for _, allowed := range allowedTypes {
			if parsed.Type == allowed {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("parsed token has disallowed type %q", parsed.Type)
		}
	})
}

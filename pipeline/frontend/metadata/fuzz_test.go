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

package metadata_test

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
)

// FuzzEnvVarSubst exercises the envsubst evaluation of untrusted yaml
// (substitution expressions are attacker controlled) and checks that it
// never panics. Values containing newlines take the quoting code path.
func FuzzEnvVarSubst(f *testing.F) {
	f.Add("image: golang:${GO_VERSION}", "1.26")
	f.Add("cmd: ${CI_COMMIT_MESSAGE}", "line1\nline2")
	f.Add("x: ${VAR=default}", "")
	f.Add("y: ${VAR/./-}", "a.b.c")
	f.Add("z: ${VAR:0:3}", "abcdef")

	f.Fuzz(func(_ *testing.T, yaml, value string) {
		environ := map[string]string{
			"GO_VERSION":        value,
			"CI_COMMIT_MESSAGE": value,
			"VAR":               value,
		}
		_, _ = metadata.EnvVarSubst(yaml, environ)
	})
}

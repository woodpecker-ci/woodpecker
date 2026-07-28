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

package model

import (
	"testing"
)

// FuzzParseRepo feeds untrusted repository full names into the owner/name
// splitter. The property checked is that parsing never panics and that a
// successful parse returns non-empty owner and name.
func FuzzParseRepo(f *testing.F) {
	f.Add("octocat/hello-world")
	f.Add("owner/group/repo")
	f.Add("/missing-owner")

	f.Fuzz(func(t *testing.T, str string) {
		user, repo, err := ParseRepo(str)
		if err != nil {
			return
		}
		if user == "" || repo == "" {
			t.Fatalf("ParseRepo(%q) returned empty owner (%q) or name (%q) without error", str, user, repo)
		}
	})
}

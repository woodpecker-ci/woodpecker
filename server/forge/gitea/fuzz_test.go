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

package gitea

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// FuzzParseHooks feeds untrusted webhook payloads into every payload level
// hook parser (push, created, pull request and release). The property checked
// is that parsing never panics, no matter how malformed the payload is.
func FuzzParseHooks(f *testing.F) {
	fixtures, err := filepath.Glob(filepath.Join("fixtures", "*.json"))
	if err != nil {
		f.Fatal(err)
	}
	for _, fixture := range fixtures {
		data, err := os.ReadFile(fixture)
		if err != nil {
			f.Fatal(err)
		}
		f.Add(data)
	}

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _, _ = parsePushHook(bytes.NewReader(data))
		_, _, _ = parseCreatedHook(bytes.NewReader(data))
		_, _, _ = parsePullRequestHook(bytes.NewReader(data))
		_, _, _ = parseReleaseHook(bytes.NewReader(data))
	})
}

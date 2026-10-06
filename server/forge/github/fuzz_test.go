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

package github

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// FuzzParseHookPayload feeds untrusted webhook payloads of arbitrary event
// types into the payload parser. The property checked is that parsing never
// panics, no matter how malformed the payload or event type is.
func FuzzParseHookPayload(f *testing.F) {
	fixtures, err := filepath.Glob(filepath.Join("fixtures", "*.json"))
	if err != nil {
		f.Fatal(err)
	}
	for _, fixture := range fixtures {
		data, err := os.ReadFile(fixture)
		if err != nil {
			f.Fatal(err)
		}
		// derive the event type from the fixture name (HookPush.json -> push)
		webhookType := "push"
		name := filepath.Base(fixture)
		switch {
		case strings.HasPrefix(name, "HookPullRequest"):
			webhookType = "pull_request"
		case strings.HasPrefix(name, "HookDeploy"):
			webhookType = "deployment"
		case strings.HasPrefix(name, "HookRelease"):
			webhookType = "release"
		}
		f.Add(webhookType, data, true)
	}

	f.Fuzz(func(_ *testing.T, webhookType string, raw []byte, merge bool) {
		_, _, _, _, _, _ = parseHookPayload(webhookType, raw, merge)
	})
}

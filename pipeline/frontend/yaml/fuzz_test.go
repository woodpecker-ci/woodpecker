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

package yaml

import (
	"testing"
)

// FuzzParseBytes exercises the whole workflow yaml parsing including all
// custom UnmarshalYAML implementations (constraints, container lists,
// string-or-slice types, ...) with untrusted input. The property checked is
// that parsing never panics.
func FuzzParseBytes(f *testing.F) {
	f.Add([]byte(sampleYaml))
	f.Add([]byte(simpleYamlAnchors))
	f.Add([]byte("steps: { a: { image: alpine, commands: [ls] } }"))
	f.Add([]byte("when:\n  - event: push\n    branch: [main, 'feat/**']"))
	f.Add([]byte("matrix:\n  GO: [1, 2]\nsteps:\n  a:\n    image: golang:${GO}"))
	f.Add([]byte("steps:\n  a:\n    image: alpine\n    settings:\n      s:\n        from_secret: token"))

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = ParseBytes(data)
	})
}

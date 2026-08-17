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

package schema_test

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/linter/schema"
)

// FuzzLintString feeds untrusted yaml through the json-schema based linter
// (yaml -> json conversion + gojsonschema validation) and checks that it
// never panics.
func FuzzLintString(f *testing.F) {
	f.Add("steps: { a: { image: alpine, commands: [ls] } }")
	f.Add("when:\n  event: push\nsteps:\n  a:\n    image: alpine")
	f.Add("skip_clone: true\nsteps: []")
	f.Add("{}")
	f.Add("- 1\n- 2")

	f.Fuzz(func(_ *testing.T, data string) {
		_, _ = schema.LintString(data)
	})
}

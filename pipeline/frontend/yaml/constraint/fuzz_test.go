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

package constraint

import (
	"testing"

	"go.yaml.in/yaml/v4"
)

// FuzzListMatch exercises doublestar glob matching with untrusted patterns
// and values from pipeline `when` constraints and checks that it never
// panics.
func FuzzListMatch(f *testing.F) {
	f.Add("feat/**", "feat/a/b", "release/*")
	f.Add("{main,dev}", "main", "")
	f.Add("[a-z]*", "abc", "**")
	f.Add(`\{esc`, "{esc", "*")

	f.Fuzz(func(_ *testing.T, include, value, exclude string) {
		c := List{
			Include: []string{include},
			Exclude: []string{exclude},
		}
		_ = c.Match(value)
	})
}

// FuzzWhenUnmarshal exercises the custom yaml unmarshalers of the `when`
// constraint tree with untrusted yaml and checks that they never panic.
func FuzzWhenUnmarshal(f *testing.F) {
	f.Add("event: push")
	f.Add("- event: [push, tag]\n  branch: main")
	f.Add("evaluate: 'CI_COMMIT_MESSAGE contains \"x\"'")
	f.Add("path:\n  include: ['src/**']\n  on_empty: true")

	f.Fuzz(func(_ *testing.T, data string) {
		when := When{}
		_ = yaml.Unmarshal([]byte(data), &when)
	})
}

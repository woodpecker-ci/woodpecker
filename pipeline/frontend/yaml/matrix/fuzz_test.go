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

package matrix

import (
	"testing"
)

// FuzzParse exercises matrix axis parsing and permutation calculation with
// untrusted yaml and checks that it never panics.
func FuzzParse(f *testing.F) {
	f.Add([]byte("matrix:\n  GO: [1, 2]\n  OS: [linux, windows]"))
	f.Add([]byte("matrix:\n  include:\n    - GO: 1\n      OS: linux"))
	f.Add([]byte("matrix: {}"))
	f.Add([]byte("matrix:\n  A: [x]"))

	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = Parse(data)
	})
}

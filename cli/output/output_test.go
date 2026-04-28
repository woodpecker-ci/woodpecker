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

package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseOutputOptions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		in   string
		out  string
		opts []string
	}{
		{
			in:  "output",
			out: "output",
		},
		{
			in:   "output=a",
			out:  "output",
			opts: []string{"a"},
		},
		{
			in:  "output=",
			out: "output",
		},
		{
			in:   "output=a,b",
			out:  "output",
			opts: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		out, opts := ParseOutputOptions(tc.in)
		assert.Equal(t, tc.out, out)
		assert.Equal(t, tc.opts, opts)
	}
}

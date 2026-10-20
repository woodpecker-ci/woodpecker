// Copyright 2023 Woodpecker Authors
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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatrix(t *testing.T) {
	axis, _ := ParseString(fakeMatrix)
	assert.Len(t, axis, 24)

	set := map[string]bool{}
	for _, perm := range axis {
		set[perm.String()] = true
	}
	assert.Len(t, set, 24)
}

func TestMatrixEmpty(t *testing.T) {
	axis, err := ParseString("")
	assert.NoError(t, err)
	assert.Empty(t, axis)
}

func TestMatrixEmptyAxis(t *testing.T) {
	axis, err := ParseString("matrix:\n  A: [a1, a2]\n  EMPTY:")
	assert.NoError(t, err)
	assert.Len(t, axis, 2)
	assert.Equal(t, "a1", axis[0]["A"])
	assert.Equal(t, "a2", axis[1]["A"])
}

func TestMatrixIncluded(t *testing.T) {
	axis, err := ParseString(fakeMatrixInclude)
	assert.NoError(t, err)
	assert.Len(t, axis, 2)
	assert.Equal(t, "1.5", axis[0]["go_version"])
	assert.Equal(t, "1.6", axis[1]["go_version"])
	assert.Equal(t, "3.4", axis[0]["python_version"])
	assert.Equal(t, "3.4", axis[1]["python_version"])
}

var fakeMatrix = `
matrix:
  go_version:
    - go1
    - go1.2
  python_version:
    - 3.2
    - 3.3
  django_version:
    - 1.7
    - 1.7.1
    - 1.7.2
  redis_version:
    - 2.6
    - 2.8
`

var fakeMatrixInclude = `
matrix:
  include:
    - go_version: 1.5
      python_version: 3.4
    - go_version: 1.6
      python_version: 3.4
`

func TestMatrixDeterministicOrder(t *testing.T) {
	// the axis permutation order must be stable across parses: workflow
	// PIDs derived from it have to match when a pipeline is compiled more
	// than once (e.g. again at agent fetch time).
	first, err := ParseString(fakeMatrix)
	assert.NoError(t, err)

	for range 20 {
		again, err := ParseString(fakeMatrix)
		assert.NoError(t, err)
		assert.EqualValues(t, first, again)
	}
}

func TestMatrixMaxAxes(t *testing.T) {
	tests := []struct {
		name   string
		matrix string
		axes   int
		err    error
	}{
		{
			name:   "below limit",
			matrix: "matrix:\n  A: [1, 2, 3, 4]\n  B: [1, 2, 3, 4]\n  C: [1, 2, 3]",
			axes:   48,
		},
		{
			name:   "at limit",
			matrix: "matrix:\n  A: [1, 2]\n  B: [1, 2]\n  C: [1, 2]\n  D: [1, 2]\n  E: [1, 2]\n  F: [1, 2]",
			axes:   64,
		},
		{
			name:   "one above limit",
			matrix: "matrix:\n  A: [1, 2, 3, 4, 5]\n  B: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13]",
			err:    ErrorMaxAxesReached,
		},
		{
			name:   "above limit",
			matrix: "matrix:\n  A: [1, 2, 3]\n  B: [1, 2, 3]\n  C: [1, 2, 3]\n  D: [1, 2, 3]",
			err:    ErrorMaxAxesReached,
		},
		{
			name:   "include below limit",
			matrix: includeMatrixHelper(5),
			axes:   5,
		},
		{
			name:   "include at limit",
			matrix: includeMatrixHelper(64),
			axes:   64,
		},
		{
			name:   "include above limit",
			matrix: includeMatrixHelper(65),
			err:    ErrorMaxAxesReached,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			axes, err := ParseString(tt.matrix)

			assert.Len(t, axes, tt.axes)

			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
				return
			}

			assert.NoError(t, err)

			// every generated workflow gets its own combination
			distinct := map[string]bool{}
			for _, axis := range axes {
				distinct[fmt.Sprint(map[string]string(axis))] = true
			}
			assert.Len(t, distinct, tt.axes)
		})
	}
}

func includeMatrixHelper(entries int) string {
	var b strings.Builder
	b.WriteString("matrix:\n  include:\n")
	for i := range entries {
		fmt.Fprintf(&b, "    - A: %d\n", i)
	}
	return b.String()
}

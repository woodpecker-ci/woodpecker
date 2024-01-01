// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeSlices(t *testing.T) {
	resultSS := MergeSlices([]string{}, []string{"a", "b"}, []string{"c"}, nil)
	assert.EqualValues(t, []string{"a", "b", "c"}, resultSS)

	resultIS := MergeSlices([]int{}, []int{1, 2}, []int{4}, nil)
	assert.EqualValues(t, []int{1, 2, 4}, resultIS)
}

func TestEqualSliceValues(t *testing.T) {
	tests := []struct {
		in1 []string
		in2 []string
		out bool
	}{{
		in1: []string{"", "ab", "12", "ab"},
		in2: []string{"12", "ab"},
		out: false,
	}, {
		in1: nil,
		in2: nil,
		out: true,
	}, {
		in1: []string{"AA", "AA", "2", " "},
		in2: []string{"2", "AA", " ", "AA"},
		out: true,
	}, {
		in1: []string{"AA", "AA", "2", " "},
		in2: []string{"2", "2", " ", "AA"},
		out: false,
	}}

	for _, tc := range tests {
		assert.EqualValues(t, tc.out, EqualSliceValues(tc.in1, tc.in2), "could not correctly process input: '%#v', %#v", tc.in1, tc.in2)
	}

	assert.True(t, EqualSliceValues([]bool{true, false, false}, []bool{false, false, true}))
	assert.False(t, EqualSliceValues([]bool{true, false, false}, []bool{true, false, true}))
}

func TestSliceToBoolMap(t *testing.T) {
	assert.Equal(t, map[string]bool{
		"a": true,
		"b": true,
		"c": true,
	}, SliceToBoolMap([]string{"a", "b", "c"}))
	assert.Equal(t, map[string]bool{}, SliceToBoolMap([]string{}))
	assert.Equal(t, map[string]bool{}, SliceToBoolMap([]string{""}))
}

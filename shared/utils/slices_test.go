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

func TestUniqSlice(t *testing.T) {
	// String tests
	t.Run("StringTests", func(t *testing.T) {
		stringTests := []struct {
			name     string
			input    []string
			expected []string
		}{
			{"Basic duplicates", []string{"apple", "banana", "apple", "orange"}, []string{"apple", "banana", "orange"}},
			{"Empty slice", []string{}, []string{}},
			{"Single item", []string{"apple"}, []string{"apple"}},
			{"All duplicates", []string{"apple", "apple", "apple"}, []string{"apple"}},
			{"Multiple items", []string{"a", "b", "c", "a", "b", "c"}, []string{"a", "b", "c"}},
		}

		for _, test := range stringTests {
			t.Run(test.name, func(t *testing.T) {
				result := UniqSlice(test.input)
				assert.ElementsMatch(t, test.expected, result, "The unique slices do not match")
			})
		}
	})

	// Integer tests
	t.Run("IntTests", func(t *testing.T) {
		intTests := []struct {
			name     string
			input    []int
			expected []int
		}{
			{"Basic duplicates", []int{1, 2, 2, 3}, []int{1, 2, 3}},
			{"Empty slice", []int{}, []int{}},
			{"Single item", []int{1}, []int{1}},
			{"All duplicates", []int{1, 1, 1}, []int{1}},
			{"Multiple items", []int{1, 2, 3, 1, 2}, []int{1, 2, 3}},
		}

		for _, test := range intTests {
			t.Run(test.name, func(t *testing.T) {
				result := UniqSlice(test.input)
				assert.ElementsMatch(t, test.expected, result, "The unique slices do not match")
			})
		}
	})
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

func TestStringSliceDeleteEmpty(t *testing.T) {
	tests := []struct {
		in  []string
		out []string
	}{{
		in:  []string{"", "ab", "ab"},
		out: []string{"ab", "ab"},
	}, {
		in:  []string{"", "ab", ""},
		out: []string{"ab"},
	}, {
		in:  []string{""},
		out: []string{},
	}}

	for _, tc := range tests {
		exp := StringSliceDeleteEmpty(tc.in)
		assert.EqualValues(t, tc.out, exp, "got '%#v', expects %#v", exp, tc.out)
	}
}

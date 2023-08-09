// Copyright 2022 Woodpecker Authors
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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDedupStrings(t *testing.T) {
	tests := []struct {
		in  []string
		out []string
	}{{
		in:  []string{"", "ab", "12", "ab"},
		out: []string{"12", "ab"},
	}, {
		in:  nil,
		out: nil,
	}, {
		in:  []string{""},
		out: nil,
	}}

	for _, tc := range tests {
		result := DedupStrings(tc.in)
		sort.Strings(result)
		if len(tc.out) == 0 {
			assert.Len(t, result, 0)
		} else {
			assert.EqualValues(t, tc.out, result, "could not correctly process input '%#v'", tc.in)
		}
	}
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
}

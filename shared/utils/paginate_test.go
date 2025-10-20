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

func TestPaginate(t *testing.T) {
	// Generic mock generator that can handle all cases
	createMock := func(pages [][]int) func(page int) []int {
		return func(page int) []int {
			if page <= 0 {
				page = 0
			} else {
				page--
			}

			if page >= len(pages) {
				return []int{}
			}

			return pages[page]
		}
	}

	tests := []struct {
		name     string
		limit    int
		pages    [][]int
		expected []int
		apiCalls int
	}{
		{
			name:     "multiple pages",
			limit:    -1,
			pages:    [][]int{{11, 12, 13}, {21, 22, 23}, {31, 32}},
			expected: []int{11, 12, 13, 21, 22, 23, 31, 32},
			apiCalls: 3,
		},
		{
			name:     "zero limit",
			limit:    0,
			pages:    [][]int{{1, 2, 3}, {1, 2, 3}, {1, 2, 3}},
			expected: []int{1, 2, 3, 1, 2, 3, 1, 2, 3},
			apiCalls: 4,
		},
		{
			name:     "empty result",
			limit:    5,
			pages:    [][]int{{}},
			expected: []int{},
			apiCalls: 1,
		},
		{
			name:     "limit less than batch",
			limit:    2,
			pages:    [][]int{{1, 2, 3, 4, 5}},
			expected: []int{1, 2},
			apiCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiExec := 0
			mock := createMock(tt.pages)

			result, _ := Paginate(func(page int) ([]int, error) {
				apiExec++
				return mock(page), nil
			}, tt.limit)

			assert.EqualValues(t, tt.apiCalls, apiExec)
			assert.EqualValues(t, tt.expected, result)
		})
	}
}

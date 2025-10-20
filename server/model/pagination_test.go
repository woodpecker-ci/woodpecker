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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestApplyPagination(t *testing.T) {
	example := []int{
		0, 1, 2,
	}

	assert.Equal(t, ApplyPagination(&ListOptions{All: true}, example), example)
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 1, PerPage: 1}, example), []int{0})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 2, PerPage: 2}, example), []int{2})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 3, PerPage: 1}, example), []int{2})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 4, PerPage: 1}, example), []int{})
	assert.Equal(t, ApplyPagination(&ListOptions{Page: 5, PerPage: 1}, example), []int{})
}

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

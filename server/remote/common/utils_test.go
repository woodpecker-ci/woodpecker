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

package common_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/server/remote/common"
)

func Test_Netrc(t *testing.T) {
	host, err := common.ExtractHostFromCloneURL("https://git.example.com/foo/bar.git")
	if err != nil {
		t.Fatal(err)
	}

	if host != "git.example.com" {
		t.Errorf("Expected host to be git.example.com, got %s", host)
	}
}

func TestPaginate(t *testing.T) {
	apiExec := 0
	apiMock := func(page int) []int {
		apiExec++
		switch page {
		case 0, 1:
			return []int{11, 12, 13}
		case 2:
			return []int{21, 22, 23}
		case 3:
			return []int{31, 32}
		default:
			return []int{}
		}
	}

	result, _ := common.Paginate(func(page int) ([]int, error) {
		return apiMock(page), nil
	})

	assert.EqualValues(t, 3, apiExec)
	if assert.Len(t, result, 8) {
		assert.EqualValues(t, []int{11, 12, 13, 21, 22, 23, 31, 32}, result)
	}
}

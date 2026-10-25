// Copyright 2026 Woodpecker Authors
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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

func TestPaginate(t *testing.T) {
	t.Run("lists every page until a short page", func(t *testing.T) {
		var requested []int
		items, err := common.Paginate("numbers", func(page int) ([]int, error) {
			requested = append(requested, page)
			if page < 3 {
				return []int{page, page}, nil
			}
			return []int{page}, nil
		})
		require.NoError(t, err)
		assert.Equal(t, []int{1, 1, 2, 2, 3}, items)
		assert.Equal(t, []int{1, 2, 3}, requested)
	})

	t.Run("stops after MaxPages full pages", func(t *testing.T) {
		requests := 0
		_, err := common.Paginate("numbers", func(page int) ([]int, error) {
			requests++
			return []int{page, page}, nil
		})
		assert.ErrorIs(t, err, &types.ErrTooManyPages{})
		assert.EqualError(t, err, "too many pages: stopped listing numbers after 100 pages")
		assert.Equal(t, common.MaxPages, requests)
	})

	t.Run("forge errors are returned as is", func(t *testing.T) {
		forgeErr := errors.New("forge down")
		_, err := common.Paginate("numbers", func(int) ([]int, error) {
			return nil, forgeErr
		})
		assert.ErrorIs(t, err, forgeErr)
	})
}

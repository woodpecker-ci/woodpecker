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

package common

import (
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	shared_utils "go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// MaxPages caps how many pages a forge driver requests for one listing, so a
// forge that keeps reporting more pages can not keep it busy forever.
// A listing whose last allowed page is still full fails with ErrTooManyPages.
const MaxPages = 100

// Paginate is shared_utils.Paginate for forge drivers: it requests at most
// MaxPages pages and names the listing in the error if there are more.
func Paginate[T any](listing string, get func(page int) ([]T, error)) ([]T, error) {
	return shared_utils.Paginate(func(page int) ([]T, error) {
		if page > MaxPages {
			return nil, &types.ErrTooManyPages{Listing: listing, Pages: MaxPages}
		}
		return get(page)
	}, -1)
}

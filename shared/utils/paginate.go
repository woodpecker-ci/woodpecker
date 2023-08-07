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

package utils

// Paginate iterates over a func call until it does not return new items and return it as list
func Paginate[T any](get func(page int) ([]T, error)) ([]T, error) {
	items := make([]T, 0, 10)
	page := 1
	lenFirstBatch := -1

	for {
		batch, err := get(page)
		if err != nil {
			return nil, err
		}
		items = append(items, batch...)

		if page == 1 {
			lenFirstBatch = len(batch)
		} else if len(batch) < lenFirstBatch || len(batch) == 0 {
			break
		}

		page++
	}

	return items, nil
}

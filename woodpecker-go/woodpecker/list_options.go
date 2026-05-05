// Copyright 2024 Woodpecker Authors
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

package woodpecker

import (
	"fmt"
	"net/url"
)

// ListOptions represents the options for the Woodpecker API pagination.
type ListOptions struct {
	Page    int
	PerPage int
}

// getURLQuery returns the query string for the ListOptions.
func (o ListOptions) getURLQuery() url.Values {
	query := make(url.Values)
	if o.Page > 0 {
		query.Add("page", fmt.Sprintf("%d", o.Page))
	}
	if o.PerPage > 0 {
		query.Add("perPage", fmt.Sprintf("%d", o.PerPage))
	}

	return query
}

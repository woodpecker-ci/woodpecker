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
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListOptions_getURLQuery(t *testing.T) {
	tests := []struct {
		name     string
		opts     ListOptions
		expected url.Values
	}{
		{
			name:     "no options",
			opts:     ListOptions{},
			expected: url.Values{},
		},
		{
			name:     "with page",
			opts:     ListOptions{Page: 2},
			expected: url.Values{"page": {"2"}},
		},
		{
			name:     "with per page",
			opts:     ListOptions{PerPage: 10},
			expected: url.Values{"perPage": {"10"}},
		},
		{
			name:     "with page and per page",
			opts:     ListOptions{Page: 3, PerPage: 20},
			expected: url.Values{"page": {"3"}, "perPage": {"20"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := tt.opts.getURLQuery()
			assert.Equal(t, tt.expected, actual)
		})
	}
}

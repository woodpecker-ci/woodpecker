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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringSliceAddToMap(t *testing.T) {
	tests := []struct {
		name     string
		sl       []string
		m        map[string]string
		expected map[string]string
		err      bool
	}{
		{
			name: "add values to map",
			sl:   []string{"foo=bar", "baz=qux=nux"},
			m:    make(map[string]string),
			expected: map[string]string{
				"foo": "bar",
				"baz": "qux=nux",
			},
			err: false,
		},
		{
			name:     "empty slice",
			sl:       []string{},
			m:        make(map[string]string),
			expected: map[string]string{},
			err:      false,
		},
		{
			name:     "missing value",
			sl:       []string{"foo", "baz=qux"},
			m:        make(map[string]string),
			expected: map[string]string{},
			err:      true,
		},
		{
			name:     "empty string in slice",
			sl:       []string{"foo=bar", "", "baz=qux"},
			m:        make(map[string]string),
			expected: map[string]string{},
			err:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := stringSliceAddToMap(tt.sl, tt.m)

			if tt.err {
				assert.Error(t, err)
			} else {
				assert.EqualValues(t, tt.expected, tt.m)
			}
		})
	}
}

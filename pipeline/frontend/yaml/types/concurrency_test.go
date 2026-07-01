// Copyright 2026 Woodpecker Authors
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

package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

func TestUnmarshalConcurrency(t *testing.T) {
	tests := []struct {
		name     string
		yaml     string
		expected Concurrency
	}{
		{
			name:     "shorthand integer",
			yaml:     `concurrency: 3`,
			expected: Concurrency{Limit: 3},
		},
		{
			name:     "full form with limit and group",
			yaml:     "concurrency:\n  limit: 2\n  group: deploy",
			expected: Concurrency{Limit: 2, Group: "deploy"},
		},
		{
			name:     "full form with limit only",
			yaml:     "concurrency:\n  limit: 1",
			expected: Concurrency{Limit: 1},
		},
		{
			name:     "full form with group only",
			yaml:     "concurrency:\n  group: deploy",
			expected: Concurrency{Group: "deploy"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var parsed struct {
				Concurrency Concurrency `yaml:"concurrency"`
			}
			err := yaml.Unmarshal([]byte(tc.yaml), &parsed)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, parsed.Concurrency)
		})
	}
}

func TestUnmarshalConcurrencyError(t *testing.T) {
	var parsed struct {
		Concurrency Concurrency `yaml:"concurrency"`
	}
	// a sequence is neither a valid shorthand int nor a valid object.
	err := yaml.Unmarshal([]byte("concurrency:\n  - invalid"), &parsed)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal concurrency")
}

func TestConcurrencyIsZero(t *testing.T) {
	tests := []struct {
		name        string
		concurrency Concurrency
		expected    bool
	}{
		{name: "empty", concurrency: Concurrency{}, expected: true},
		{name: "disabled limit", concurrency: Concurrency{Limit: 0}, expected: true},
		{name: "negative limit", concurrency: Concurrency{Limit: -1}, expected: true},
		{name: "with limit", concurrency: Concurrency{Limit: 1}, expected: false},
		{name: "with group only", concurrency: Concurrency{Group: "deploy"}, expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.concurrency.IsZero())
		})
	}
}

func TestMarshalConcurrencyOmitsZero(t *testing.T) {
	parsed := struct {
		Concurrency Concurrency `yaml:"concurrency,omitempty"`
	}{}
	out, err := yaml.Marshal(parsed)
	assert.NoError(t, err)
	// IsZero makes an unset concurrency omitted from the output.
	assert.NotContains(t, string(out), "concurrency")
}

func TestMarshalConcurrency(t *testing.T) {
	tests := []struct {
		name        string
		concurrency Concurrency
		expected    string
	}{
		{
			name:        "shorthand integer when no group",
			concurrency: Concurrency{Limit: 3},
			expected:    "concurrency: 3\n",
		},
		{
			name:        "full form when group is set",
			concurrency: Concurrency{Limit: 2, Group: "deploy"},
			expected:    "concurrency:\n    limit: 2\n    group: deploy\n",
		},
		{
			name:        "group only omits the zero limit",
			concurrency: Concurrency{Group: "deploy"},
			expected:    "concurrency:\n    group: deploy\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := struct {
				Concurrency Concurrency `yaml:"concurrency"`
			}{tc.concurrency}
			out, err := yaml.Marshal(wrapped)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, string(out))
		})
	}
}

func TestConcurrencyRoundTrip(t *testing.T) {
	tests := []struct {
		name        string
		concurrency Concurrency
		// wantShorthand asserts whether the marshaled form is the bare
		// integer (true) or the expanded object (false).
		wantShorthand bool
	}{
		{name: "limit only", concurrency: Concurrency{Limit: 3}, wantShorthand: true},
		{name: "limit and group", concurrency: Concurrency{Limit: 2, Group: "deploy"}, wantShorthand: false},
		{name: "group only", concurrency: Concurrency{Group: "deploy"}, wantShorthand: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			type wrapper struct {
				Concurrency Concurrency `yaml:"concurrency"`
			}

			out, err := yaml.Marshal(wrapper{tc.concurrency})
			assert.NoError(t, err)

			if tc.wantShorthand {
				assert.NotContains(t, string(out), "limit:", "expected shorthand, got object form")
				assert.NotContains(t, string(out), "group:")
			} else {
				assert.Contains(t, string(out), "group:")
			}

			var back wrapper
			err = yaml.Unmarshal(out, &back)
			assert.NoError(t, err)
			assert.Equal(t, tc.concurrency, back.Concurrency, "value changed across marshal/unmarshal")
		})
	}
}

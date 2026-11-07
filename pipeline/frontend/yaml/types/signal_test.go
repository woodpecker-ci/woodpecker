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
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
)

func TestSignalConfigUnmarshal(t *testing.T) {
	five := 5
	tests := []struct {
		name string
		in   string
		want map[string]SignalConfig
	}{
		{"single code", "restart: 75", map[string]SignalConfig{"restart": {ExitCodes: base.IntOrSlice{75}}}},
		{"list of codes", "skip: [78, 79]", map[string]SignalConfig{"skip": {ExitCodes: base.IntOrSlice{78, 79}}}},
		{
			"long form",
			"restart:\n  exit_codes: [75, 76]\n  max_restarts: 5",
			map[string]SignalConfig{"restart": {ExitCodes: base.IntOrSlice{75, 76}, MaxRestarts: &five}},
		},
		{
			"long form single code",
			"restart:\n  exit_codes: 75",
			map[string]SignalConfig{"restart": {ExitCodes: base.IntOrSlice{75}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := make(map[string]SignalConfig)
			require.NoError(t, yaml.Unmarshal([]byte(tt.in), &got))
			assert.Equal(t, tt.want, got)
		})
	}

	t.Run("invalid", func(t *testing.T) {
		got := make(map[string]SignalConfig)
		assert.Error(t, yaml.Unmarshal([]byte(`skip: abc`), &got))
	})
}

func TestSignalConfigMarshal(t *testing.T) {
	out, err := yaml.Marshal(map[string]SignalConfig{"skip": {ExitCodes: base.IntOrSlice{78}}})
	require.NoError(t, err)
	assert.Equal(t, "skip: 78\n", string(out))

	three := 3
	out, err = yaml.Marshal(map[string]SignalConfig{"restart": {ExitCodes: base.IntOrSlice{75}, MaxRestarts: &three}})
	require.NoError(t, err)
	assert.Equal(t, "restart:\n    exit_codes: 75\n    max_restarts: 3\n", string(out))
}

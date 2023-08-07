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

package compiler

import (
	"testing"

	"github.com/stretchr/testify/assert"

	yaml_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
	yaml_base_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types/base"
)

func TestSecretAvailable(t *testing.T) {
	secret := Secret{
		Match:      []string{"golang"},
		PluginOnly: false,
	}
	assert.True(t, secret.Available(&yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
	// secret only available for "golang" plugin
	secret = Secret{
		Match:      []string{"golang"},
		PluginOnly: true,
	}
	assert.True(t, secret.Available(&yaml_types.Container{
		Image:    "golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{},
	}))
	assert.False(t, secret.Available(&yaml_types.Container{
		Image:    "not-golang",
		Commands: yaml_base_types.StringOrSlice{"echo 'this is not a plugin'"},
	}))
}

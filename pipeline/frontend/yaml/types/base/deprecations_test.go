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

// TODO: delete file after v3.0.0 release

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructMap struct {
	Foos EnvironmentMap `yaml:"foos,omitempty"`
}

func TestEnvironmentMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`
	s := StructMap{}
	err := yaml.Unmarshal([]byte(str), &s)
	if assert.Error(t, err) {
		assert.EqualValues(t, "list syntax for 'environment' has been removed, use map syntax instead (https://woodpecker-ci.org/docs/usage/environment)", err.Error())
	}

	s.Foos = EnvironmentMap{"bar": "baz", "far": "faz"}
	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)
	str = `foos:
    bar: baz
    far: faz
`
	assert.EqualValues(t, str, string(d))

	s2 := StructMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, EnvironmentMap{"bar": "baz", "far": "faz"}, s2.Foos)
}

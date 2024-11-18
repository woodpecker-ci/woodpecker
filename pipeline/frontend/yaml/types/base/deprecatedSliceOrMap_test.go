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

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructDeprecatedSliceOrMap struct {
	Foos DeprecatedSliceOrMap `yaml:"foos,omitempty"`
	Bars []string             `yaml:"bars,omitempty"`
}

func TestDeprecatedSliceOrMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`

	s := StructDeprecatedSliceOrMap{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

	assert.Equal(t, DeprecatedSliceOrMap{Map: map[string]any{"bar": "baz", "far": "faz"}, WasSlice: true}, s.Foos)

	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)
	str = `foos:
    bar: baz
    far: faz
`
	assert.EqualValues(t, str, string(d))

	s2 := StructDeprecatedSliceOrMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, DeprecatedSliceOrMap{Map: map[string]any{"bar": "baz", "far": "faz"}, WasSlice: false}, s2.Foos)
}

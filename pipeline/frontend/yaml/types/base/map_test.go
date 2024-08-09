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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructSliceOrMap struct {
	Foos SliceOrMap `yaml:"foos,omitempty"`
	Bars []string   `yaml:"bars"`
}

func TestSliceOrMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`

	s := StructSliceOrMap{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

	assert.Equal(t, SliceOrMap{"bar": "baz", "far": "faz"}, s.Foos)

	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)

	s2 := StructSliceOrMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, SliceOrMap{"bar": "baz", "far": "faz"}, s2.Foos)
}

func TestStr2SliceOrMapPtrMap(t *testing.T) {
	s := map[string]*StructSliceOrMap{"udav": {
		Foos: SliceOrMap{"io.rancher.os.bar": "baz", "io.rancher.os.far": "true"},
		Bars: []string{},
	}}
	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)

	s2 := map[string]*StructSliceOrMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, s, s2)
}

var sampleStructSliceOrMap = `
foos:
  io.rancher.os.bar: baz
  io.rancher.os.far: true
bars: []
`

func TestUnmarshalSliceOrMap(t *testing.T) {
	s := StructSliceOrMap{}
	err := yaml.Unmarshal([]byte(sampleStructSliceOrMap), &s)
	assert.Equal(t, fmt.Errorf("Cannot unmarshal 'true' of type bool into a string value"), err)
}

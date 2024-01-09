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

package base

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

type StructStringorInt struct {
	Foo StringOrInt
}

func TestStringorIntYaml(t *testing.T) {
	for _, str := range []string{`{foo: 10}`, `{foo: "10"}`} {
		s := StructStringorInt{}
		assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

		assert.Equal(t, StringOrInt(10), s.Foo)

		d, err := yaml.Marshal(&s)
		assert.NoError(t, err)

		s2 := StructStringorInt{}
		assert.NoError(t, yaml.Unmarshal(d, &s2))

		assert.Equal(t, StringOrInt(10), s2.Foo)
	}
}

type StructStringOrSlice struct {
	Foo StringOrSlice
}

func TestStringOrSliceYaml(t *testing.T) {
	str := `{foo: [bar, "baz"]}`
	s := StructStringOrSlice{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))
	assert.Equal(t, StringOrSlice{"bar", "baz"}, s.Foo)

	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)

	s = StructStringOrSlice{}
	assert.NoError(t, yaml.Unmarshal(d, &s))
	assert.Equal(t, StringOrSlice{"bar", "baz"}, s.Foo)

	str = `{foo: []}`
	s = StructStringOrSlice{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))
	assert.Equal(t, StringOrSlice{}, s.Foo)

	str = `{}`
	s = StructStringOrSlice{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))
	assert.Nil(t, s.Foo)
}

type StructSliceorMap struct {
	Foos SliceOrMap `yaml:"foos,omitempty"`
	Bars []string   `yaml:"bars"`
}

func TestSliceOrMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`

	s := StructSliceorMap{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

	assert.Equal(t, SliceOrMap{"bar": "baz", "far": "faz"}, s.Foos)

	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)

	s2 := StructSliceorMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, SliceOrMap{"bar": "baz", "far": "faz"}, s2.Foos)
}

var sampleStructSliceorMap = `
foos:
  io.rancher.os.bar: baz
  io.rancher.os.far: true
bars: []
`

func TestUnmarshalSliceOrMap(t *testing.T) {
	s := StructSliceorMap{}
	err := yaml.Unmarshal([]byte(sampleStructSliceorMap), &s)
	assert.Equal(t, fmt.Errorf("Cannot unmarshal 'true' of type bool into a string value"), err)
}

func TestStr2SliceOrMapPtrMap(t *testing.T) {
	s := map[string]*StructSliceorMap{"udav": {
		Foos: SliceOrMap{"io.rancher.os.bar": "baz", "io.rancher.os.far": "true"},
		Bars: []string{},
	}}
	d, err := yaml.Marshal(&s)
	assert.NoError(t, err)

	s2 := map[string]*StructSliceorMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, s, s2)
}

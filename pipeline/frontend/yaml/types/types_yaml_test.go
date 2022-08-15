package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructStringorInt struct {
	Foo StringorInt
}

func TestStringorIntYaml(t *testing.T) {
	for _, str := range []string{`{foo: 10}`, `{foo: "10"}`} {
		s := StructStringorInt{}
		assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

		assert.Equal(t, StringorInt(10), s.Foo)

		d, err := yaml.Marshal(&s)
		assert.Nil(t, err)

		s2 := StructStringorInt{}
		assert.NoError(t, yaml.Unmarshal(d, &s2))

		assert.Equal(t, StringorInt(10), s2.Foo)
	}
}

type StructStringorslice struct {
	Foo Stringorslice
}

func TestStringorsliceYaml(t *testing.T) {
	str := `{foo: [bar, baz]}`

	s := StructStringorslice{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

	assert.Equal(t, Stringorslice{"bar", "baz"}, s.Foo)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructStringorslice{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, Stringorslice{"bar", "baz"}, s2.Foo)
}

type StructSliceorMap struct {
	Foos SliceorMap `yaml:"foos,omitempty"`
	Bars []string   `yaml:"bars"`
}

func TestSliceOrMapYaml(t *testing.T) {
	str := `{foos: [bar=baz, far=faz]}`

	s := StructSliceorMap{}
	assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

	assert.Equal(t, SliceorMap{"bar": "baz", "far": "faz"}, s.Foos)

	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := StructSliceorMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, SliceorMap{"bar": "baz", "far": "faz"}, s2.Foos)
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
	assert.Equal(t, fmt.Errorf("cannot unmarshal 'true' of type bool into a string value"), err)
}

func TestStr2SliceOrMapPtrMap(t *testing.T) {
	s := map[string]*StructSliceorMap{"udav": {
		Foos: SliceorMap{"io.rancher.os.bar": "baz", "io.rancher.os.far": "true"},
		Bars: []string{},
	}}
	d, err := yaml.Marshal(&s)
	assert.Nil(t, err)

	s2 := map[string]*StructSliceorMap{}
	assert.NoError(t, yaml.Unmarshal(d, &s2))

	assert.Equal(t, s, s2)
}

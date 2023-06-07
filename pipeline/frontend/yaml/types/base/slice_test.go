package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructStringOrSlice struct {
	Foo StringOrSlice
}

func TestStringOrSliceYaml(t *testing.T) {
	t.Run("unmarshal", func(t *testing.T) {
		str := `{foo: [bar, baz]}`

		s := StructStringOrSlice{}
		assert.NoError(t, yaml.Unmarshal([]byte(str), &s))

		assert.Equal(t, StringOrSlice{"bar", "baz"}, s.Foo)

		d, err := yaml.Marshal(&s)
		assert.Nil(t, err)

		s2 := StructStringOrSlice{}
		assert.NoError(t, yaml.Unmarshal(d, &s2))

		assert.Equal(t, StringOrSlice{"bar", "baz"}, s2.Foo)
	})

	t.Run("marshal", func(t *testing.T) {
		str := StructStringOrSlice{}
		out, err := yaml.Marshal(str)
		assert.NoError(t, err)
		assert.EqualValues(t, "foo: \"\"\n", string(out))

		str = StructStringOrSlice{Foo: []string{"a\""}}
		out, err = yaml.Marshal(str)
		assert.NoError(t, err)
		assert.EqualValues(t, "foo: \"\"\n", string(out))

		str = StructStringOrSlice{Foo: []string{"a", "b", "c"}}
		out, err = yaml.Marshal(str)
		assert.NoError(t, err)
		assert.EqualValues(t, "foo: \"\"\n", string(out))
	})
}

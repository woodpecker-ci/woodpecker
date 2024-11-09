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

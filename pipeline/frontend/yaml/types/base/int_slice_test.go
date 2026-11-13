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

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.yaml.in/yaml/v4"
)

type StructIntOrSlice struct {
	Foo IntOrSlice `yaml:"foo"`
}

func TestIntOrSliceYaml(t *testing.T) {
	t.Run("unmarshal single", func(t *testing.T) {
		s := StructIntOrSlice{}
		assert.NoError(t, yaml.Unmarshal([]byte(`foo: 78`), &s))
		assert.Equal(t, IntOrSlice{78}, s.Foo)
	})

	t.Run("unmarshal list", func(t *testing.T) {
		s := StructIntOrSlice{}
		assert.NoError(t, yaml.Unmarshal([]byte(`foo: [78, 79]`), &s))
		assert.Equal(t, IntOrSlice{78, 79}, s.Foo)
	})

	t.Run("unmarshal invalid", func(t *testing.T) {
		s := StructIntOrSlice{}
		assert.Error(t, yaml.Unmarshal([]byte(`foo: abc`), &s))
		assert.Error(t, yaml.Unmarshal([]byte(`foo: [78, abc]`), &s))
	})

	t.Run("marshal", func(t *testing.T) {
		out, err := yaml.Marshal(StructIntOrSlice{Foo: IntOrSlice{78}})
		assert.NoError(t, err)
		assert.Equal(t, "foo: 78\n", string(out))

		out, err = yaml.Marshal(StructIntOrSlice{Foo: IntOrSlice{78, 79}})
		assert.NoError(t, err)
		assert.Equal(t, "foo:\n    - 78\n    - 79\n", string(out))
	})
}

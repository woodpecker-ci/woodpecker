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
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

type StructDependsOn struct {
	DependsOn DependsOn `yaml:"depends_on,omitempty"`
}

func TestDependsOnYaml(t *testing.T) {
	t.Run("unmarshal string", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`{depends_on: lint}`), &s))
		assert.Equal(t, DependsOn{{Name: "lint"}}, s.DependsOn)
	})

	t.Run("unmarshal string array", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`{depends_on: [lint, test]}`), &s))
		assert.Equal(t, DependsOn{{Name: "lint"}, {Name: "test"}}, s.DependsOn)
	})

	t.Run("unmarshal object array", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`
depends_on:
  - name: lint
    optional: true
  - name: test
`), &s))
		assert.Equal(t, DependsOn{
			{Name: "lint", Optional: true},
			{Name: "test"},
		}, s.DependsOn)
	})

	t.Run("unmarshal mixed array", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`
depends_on:
  - lint
  - name: test
    optional: true
`), &s))
		assert.Equal(t, DependsOn{
			{Name: "lint"},
			{Name: "test", Optional: true},
		}, s.DependsOn)
	})

	t.Run("unmarshal empty array", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`{depends_on: []}`), &s))
		assert.NotNil(t, s.DependsOn)
		assert.Empty(t, s.DependsOn)
	})

	t.Run("unmarshal absent", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`{}`), &s))
		assert.Nil(t, s.DependsOn)
	})

	t.Run("unmarshal object missing name", func(t *testing.T) {
		s := StructDependsOn{}
		err := yaml.Unmarshal([]byte(`{depends_on: [{optional: true}]}`), &s)
		assert.Error(t, err)
	})
}

func TestDependsOnMarshal(t *testing.T) {
	t.Run("all required marshals as strings", func(t *testing.T) {
		s := StructDependsOn{DependsOn: DependsOn{{Name: "lint"}, {Name: "test"}}}
		out, err := yaml.Marshal(s)
		assert.NoError(t, err)
		assert.Equal(t, "depends_on:\n    - lint\n    - test\n", string(out))
	})

	t.Run("single required marshals as string", func(t *testing.T) {
		s := StructDependsOn{DependsOn: DependsOn{{Name: "lint"}}}
		out, err := yaml.Marshal(s)
		assert.NoError(t, err)
		assert.Equal(t, "depends_on: lint\n", string(out))
	})

	t.Run("with optional marshals as objects", func(t *testing.T) {
		s := StructDependsOn{DependsOn: DependsOn{
			{Name: "lint"},
			{Name: "test", Optional: true},
		}}
		out, err := yaml.Marshal(s)
		assert.NoError(t, err)
		assert.Equal(t, "depends_on:\n    - name: lint\n    - name: test\n      optional: true\n", string(out))
	})

	t.Run("empty omitted", func(t *testing.T) {
		s := StructDependsOn{}
		out, err := yaml.Marshal(s)
		assert.NoError(t, err)
		assert.Equal(t, "{}\n", string(out))
	})
}

func TestDependsOnHelpers(t *testing.T) {
	d := DependsOn{
		{Name: "a"},
		{Name: "b", Optional: true},
		{Name: "c"},
		{Name: "d", Optional: true},
	}
	assert.Equal(t, []string{"a", "b", "c", "d"}, d.Names())
	assert.Equal(t, []string{"a", "c"}, d.RequiredNames())
	assert.Equal(t, []string{"b", "d"}, d.OptionalNames())

	var nilDeps DependsOn
	assert.Nil(t, nilDeps.Names())
	assert.Nil(t, nilDeps.RequiredNames())
	assert.Nil(t, nilDeps.OptionalNames())
}

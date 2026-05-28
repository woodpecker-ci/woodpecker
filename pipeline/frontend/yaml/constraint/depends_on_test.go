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

package constraint

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

	t.Run("empty array unmarshals non-nil (signals step DAG mode with no edges)", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`{depends_on: []}`), &s))
		assert.NotNil(t, s.DependsOn, "present-but-empty must stay non-nil; the step compiler treats nil as sequential and non-nil as DAG")
		assert.Empty(t, s.DependsOn)
	})

	t.Run("absent key unmarshals nil (signals sequential step execution)", func(t *testing.T) {
		s := StructDependsOn{}
		assert.NoError(t, yaml.Unmarshal([]byte(`{}`), &s))
		assert.Nil(t, s.DependsOn, "absent must stay nil; otherwise plain step lists would be forced into DAG mode")
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

	t.Run("nil omitted", func(t *testing.T) {
		s := StructDependsOn{}
		out, err := yaml.Marshal(s)
		assert.NoError(t, err)
		assert.Equal(t, "{}\n", string(out))
	})

	t.Run("non-nil empty marshals as empty array (preserves step DAG mode signal)", func(t *testing.T) {
		s := StructDependsOn{DependsOn: DependsOn{}}
		out, err := yaml.Marshal(s)
		assert.NoError(t, err)
		assert.Equal(t, "depends_on: []\n", string(out), "non-nil empty must serialize as []; omitting it would flip step execution from DAG to sequential on the next read")
	})
}

// TestDependsOnRoundTrip locks the marshal/unmarshal contract that the
// step compiler relies on: nil means sequential, non-nil (even empty)
// means DAG. The contract has to survive a serialize/deserialize cycle,
// otherwise a config that round-trips through any tooling would silently
// switch step execution mode.
func TestDependsOnRoundTrip(t *testing.T) {
	cases := []struct {
		name string
		in   DependsOn
	}{
		{"nil stays nil", nil},
		{"non-nil empty stays non-nil empty", DependsOn{}},
		{"single required", DependsOn{{Name: "lint"}}},
		{"multiple required", DependsOn{{Name: "lint"}, {Name: "test"}}},
		{"mixed required and optional", DependsOn{{Name: "lint"}, {Name: "test", Optional: true}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, err := yaml.Marshal(StructDependsOn{DependsOn: tc.in})
			assert.NoError(t, err)

			var back StructDependsOn
			assert.NoError(t, yaml.Unmarshal(out, &back))

			if tc.in == nil {
				assert.Nil(t, back.DependsOn)
				return
			}
			assert.NotNil(t, back.DependsOn, "non-nil input must round-trip non-nil")
			assert.Equal(t, tc.in, back.DependsOn)
		})
	}
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

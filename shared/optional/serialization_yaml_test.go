// Copyright 2025 Woodpecker Authors.
// Copyright 2024 "6543".
//
// Licensed under the MIT License.

package optional_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"go.woodpecker-ci.org/woodpecker/v3/shared/optional"
)

type testBoolStruct struct {
	OptBoolOmitEmpty1 optional.Option[bool] `json:"opt_bool_omit_empty_1,omitempty" yaml:"opt_bool_omit_empty_1,omitempty"`
	OptBoolOmitEmpty2 optional.Option[bool] `json:"opt_bool_omit_empty_2,omitempty" yaml:"opt_bool_omit_empty_2,omitempty"`
	OptBoolOmitEmpty3 optional.Option[bool] `json:"opt_bool_omit_empty_3,omitempty" yaml:"opt_bool_omit_empty_3,omitempty"`
	OptBool4          optional.Option[bool] `json:"opt_bool_4" yaml:"opt_bool_4"`
	OptBool5          optional.Option[bool] `json:"opt_bool_5" yaml:"opt_bool_5"`
	OptBool6          optional.Option[bool] `json:"opt_bool_6" yaml:"opt_bool_6"`
}

func TestOptionalBoolYaml(t *testing.T) {
	tYaml := `
opt_bool_omit_empty_1: false
opt_bool_omit_empty_2: true
opt_bool_4: false
opt_bool_5: true
`

	tObj := new(testBoolStruct)
	t.Run("Unmarshal", func(t *testing.T) {
		err := yaml.Unmarshal([]byte(tYaml), tObj)
		require.NoError(t, err)
		assert.EqualValues(t, &testBoolStruct{
			OptBoolOmitEmpty1: optional.Some(false),
			OptBoolOmitEmpty2: optional.Some(true),
			OptBoolOmitEmpty3: optional.None[bool](),
			OptBool4:          optional.Some(false),
			OptBool5:          optional.Some(true),
			OptBool6:          optional.None[bool](),
		}, tObj)
	})
	t.Run("Marshal", func(t *testing.T) {
		tBytes, err := yaml.Marshal(tObj)
		require.NoError(t, err)
		assert.EqualValues(t, `opt_bool_omit_empty_1: false
opt_bool_omit_empty_2: true
opt_bool_4: false
opt_bool_5: true
opt_bool_6: null
`, string(tBytes))
	})
}

func TestOptionalToYaml(t *testing.T) {
	tests := []struct {
		name string
		obj  *testSerializationStruct
		want string
	}{
		{
			name: "empty",
			obj:  new(testSerializationStruct),
			want: `normal_string: ""
normal_bool: false
optional_two_bool: null
optional_two_string: null
`,
		},
		{
			name: "some",
			obj: &testSerializationStruct{
				NormalString: "a string",
				NormalBool:   true,
				OptBool:      optional.Some(false),
				OptString:    optional.Some(""),
			},
			want: `normal_string: a string
normal_bool: true
optional_bool: false
optional_string: ""
optional_two_bool: null
optional_two_string: null
`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := yaml.Marshal(tc.obj)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, string(b), "yaml module returned unexpected")
		})
	}
}

func TestOptionalFromYaml(t *testing.T) {
	tests := []struct {
		name string
		data string
		want testSerializationStruct
	}{
		{
			name: "empty",
			data: ``,
			want: testSerializationStruct{},
		},
		{
			name: "empty but init",
			data: `normal_string: ""
normal_bool: false
optional_bool:
optional_two_bool:
optional_two_string:
`,
			want: testSerializationStruct{},
		},
		{
			name: "some",
			data: `
normal_string: a string
normal_bool: true
optional_bool: false
optional_string: ""
optional_two_bool: null
optional_twostring: null
`,
			want: testSerializationStruct{
				NormalString: "a string",
				NormalBool:   true,
				OptBool:      optional.Some(false),
				OptString:    optional.Some(""),
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var obj testSerializationStruct
			err := yaml.Unmarshal([]byte(tc.data), &obj)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, obj, "yaml module returned unexpected")
		})
	}
}

// Copyright 2025 Woodpecker Authors.
// Copyright 2024 "6543".
//
// Licensed under the MIT License.

package optional_test

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/shared/optional"
)

func TestOptionalToJson(t *testing.T) {
	tests := []struct {
		name string
		obj  *testSerializationStruct
		want string
	}{
		{
			name: "empty",
			obj:  new(testSerializationStruct),
			want: `{"normal_string":"","normal_bool":false,"optional_two_bool":null,"optional_twostring":null}`,
		},
		{
			name: "some",
			obj: &testSerializationStruct{
				NormalString: "a string",
				NormalBool:   true,
				OptBool:      optional.Some(false),
				OptString:    optional.Some(""),
			},
			want: `{"normal_string":"a string","normal_bool":true,"optional_bool":false,"optional_string":"","optional_two_bool":null,"optional_twostring":null}`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b, err := json.Marshal(tc.obj)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, string(b), "gitea json module returned unexpected")

			b, err = jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(tc.obj)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, string(b), "std json module returned unexpected")
		})
	}
}

func TestOptionalFromJson(t *testing.T) {
	tests := []struct {
		name string
		data string
		want testSerializationStruct
	}{
		{
			name: "empty",
			data: `{}`,
			want: testSerializationStruct{
				NormalString: "",
				OptBool:      optional.None[bool](),
			},
		},
		{
			name: "some",
			data: `{"normal_string":"a string","normal_bool":true,"optional_bool":false,"optional_string":"","optional_two_bool":null,"optional_twostring":null}`,
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
			var obj1 testSerializationStruct
			err := json.Unmarshal([]byte(tc.data), &obj1)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, obj1, "gitea json module returned unexpected")

			var obj2 testSerializationStruct
			err = jsoniter.ConfigCompatibleWithStandardLibrary.Unmarshal([]byte(tc.data), &obj2)
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, obj2, "std json module returned unexpected")
		})
	}
}

// Copyright 2025 Woodpecker Authors.
// Copyright 2024 "6543".
//
// Licensed under the MIT License.

package optional_test

import (
	"go.woodpecker-ci.org/woodpecker/v3/shared/optional"
)

type testSerializationStruct struct {
	NormalString string                  `json:"normal_string" yaml:"normal_string"`
	NormalBool   bool                    `json:"normal_bool" yaml:"normal_bool"`
	OptBool      optional.Option[bool]   `json:"optional_bool,omitempty" yaml:"optional_bool,omitempty"`
	OptString    optional.Option[string] `json:"optional_string,omitempty" yaml:"optional_string,omitempty"`
	OptTwoBool   optional.Option[bool]   `json:"optional_two_bool" yaml:"optional_two_bool"`
	OptTwoString optional.Option[string] `json:"optional_twostring" yaml:"optional_two_string"`
}

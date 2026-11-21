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

import "errors"

// IntOrSlice represents an integer or an array of integers.
// We need to override the yaml decoder to accept both options.
type IntOrSlice []int

// UnmarshalYAML implements the Unmarshaler interface.
func (s *IntOrSlice) UnmarshalYAML(unmarshal func(any) error) error {
	var intType int
	if err := unmarshal(&intType); err == nil {
		*s = []int{intType}
		return nil
	}

	var sliceType []int
	if err := unmarshal(&sliceType); err == nil {
		*s = sliceType
		return nil
	}

	return errors.New("failed to unmarshal IntOrSlice")
}

// MarshalYAML implements custom Yaml marshaling.
func (s IntOrSlice) MarshalYAML() (any, error) {
	if len(s) == 0 {
		return nil, nil
	} else if len(s) == 1 {
		return s[0], nil
	}
	return []int(s), nil
}

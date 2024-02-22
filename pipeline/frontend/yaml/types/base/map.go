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
	"errors"
	"fmt"
	"strings"
)

// SliceOrMap represents a map of strings, string slice are converted into a map
type SliceOrMap map[string]any

// UnmarshalYAML implements the Unmarshaler interface.
func (s *SliceOrMap) UnmarshalYAML(unmarshal func(any) error) error {
	var sliceType []any
	if err := unmarshal(&sliceType); err == nil {
		parts := map[string]any{}
		for _, s := range sliceType {
			if str, ok := s.(string); ok {
				str := strings.TrimSpace(str)
				keyValueSlice := strings.SplitN(str, "=", 2)

				key := keyValueSlice[0]
				val := ""
				if len(keyValueSlice) == 2 {
					val = keyValueSlice[1]
				}
				parts[key] = val
			} else {
				return fmt.Errorf("cannot unmarshal '%v' of type %T into a string value", s, s)
			}
		}
		*s = parts
		return nil
	}

	var mapType map[string]any
	if err := unmarshal(&mapType); err == nil {
		*s = mapType
		return nil
	}

	return errors.New("failed to unmarshal SliceOrMap")
}

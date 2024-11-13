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

// TODO: delete file after v3.0.0 release

package base

import (
	"fmt"
)

type EnvironmentMap map[string]any

// UnmarshalYAML implements the Unmarshaler interface.
func (s *EnvironmentMap) UnmarshalYAML(unmarshal func(any) error) error {
	var mapType map[string]any
	err := unmarshal(&mapType)
	if err == nil {
		*s = mapType
		return nil
	}

	var sliceType []any
	if err := unmarshal(&sliceType); err == nil {
		return fmt.Errorf("list syntax for 'environment' has been removed, use map syntax instead (https://woodpecker-ci.org/docs/usage/environment)")
	}

	return err
}

type SecretsSlice []string

// UnmarshalYAML implements the Unmarshaler interface.
func (s *SecretsSlice) UnmarshalYAML(unmarshal func(any) error) error {
	var stringSlice []string
	err := unmarshal(&stringSlice)
	if err == nil {
		*s = stringSlice
		return nil
	}

	var objectSlice []any
	if err := unmarshal(&objectSlice); err == nil {
		return fmt.Errorf("'secrets' property has been removed, use 'from_secret' instead (https://woodpecker-ci.org/docs/usage/secrets)")
	}

	return err
}

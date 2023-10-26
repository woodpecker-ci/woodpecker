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
	"strconv"

	"gopkg.in/yaml.v3"
)

// BoolTrue is a custom Yaml boolean type that defaults to true.
type BoolTrue struct {
	value bool
}

// UnmarshalYAML implements custom Yaml unmarshaling.
func (b *BoolTrue) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}

	v, err := strconv.ParseBool(s)
	if err == nil {
		b.value = !v
	}
	if s != "" && err != nil {
		return err
	}
	return nil
}

// Bool returns the bool value.
func (b BoolTrue) Bool() bool {
	return !b.value
}

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
	"strconv"

	"github.com/docker/go-units"
)

// StringOrInt represents a string or an integer.
type StringOrInt int64

// UnmarshalYAML implements the Unmarshaler interface.
func (s *StringOrInt) UnmarshalYAML(unmarshal func(any) error) error {
	var intType int64
	if err := unmarshal(&intType); err == nil {
		*s = StringOrInt(intType)
		return nil
	}

	var stringType string
	if err := unmarshal(&stringType); err == nil {
		intType, err := strconv.ParseInt(stringType, 10, 64)
		if err != nil {
			return err
		}
		*s = StringOrInt(intType)
		return nil
	}

	return errors.New("failed to unmarshal StringOrInt")
}

// MemStringOrInt represents a string or an integer
// the String supports notations like 10m for then Megabyte of memory
type MemStringOrInt int64

// UnmarshalYAML implements the Unmarshaler interface.
func (s *MemStringOrInt) UnmarshalYAML(unmarshal func(any) error) error {
	var intType int64
	if err := unmarshal(&intType); err == nil {
		*s = MemStringOrInt(intType)
		return nil
	}

	var stringType string
	if err := unmarshal(&stringType); err == nil {
		intType, err := units.RAMInBytes(stringType)
		if err != nil {
			return err
		}
		*s = MemStringOrInt(intType)
		return nil
	}

	return errors.New("failed to unmarshal MemStringOrInt")
}

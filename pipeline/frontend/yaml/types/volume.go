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

package types

import (
	"errors"
	"fmt"
	"strings"
)

// Volumes represents a list of service volumes in compose file.
// It has several representation, hence this specific struct.
type Volumes struct {
	Volumes []*Volume
}

// Volume represent a service volume
type Volume struct {
	Source      string `yaml:"-"`
	Destination string `yaml:"-"`
	AccessMode  string `yaml:"-"`
}

// String implements the Stringer interface.
func (v *Volume) String() string {
	var paths []string
	if v.Source != "" {
		paths = []string{v.Source, v.Destination}
	} else {
		paths = []string{v.Destination}
	}
	if v.AccessMode != "" {
		paths = append(paths, v.AccessMode)
	}
	return strings.Join(paths, ":")
}

// MarshalYAML implements the Marshaller interface.
func (v Volumes) MarshalYAML() (any, error) {
	vs := []string{}
	for _, volume := range v.Volumes {
		vs = append(vs, volume.String())
	}
	return vs, nil
}

// UnmarshalYAML implements the Unmarshaler interface.
func (v *Volumes) UnmarshalYAML(unmarshal func(any) error) error {
	var sliceType []any
	if err := unmarshal(&sliceType); err == nil {
		v.Volumes = []*Volume{}
		for _, volume := range sliceType {
			name, ok := volume.(string)
			if !ok {
				return fmt.Errorf("cannot unmarshal '%v' to type %T into a string value", name, name)
			}
			elts := strings.SplitN(name, ":", 3)
			var vol *Volume
			//nolint: gomnd
			switch {
			case len(elts) == 1:
				vol = &Volume{
					Destination: elts[0],
				}
			case len(elts) == 2:
				vol = &Volume{
					Source:      elts[0],
					Destination: elts[1],
				}
			case len(elts) == 3:
				vol = &Volume{
					Source:      elts[0],
					Destination: elts[1],
					AccessMode:  elts[2],
				}
			default:
				// FIXME
				return fmt.Errorf("")
			}
			v.Volumes = append(v.Volumes, vol)
		}
		return nil
	}

	return errors.New("failed to unmarshal Volumes")
}

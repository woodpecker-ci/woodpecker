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

package types

import (
	"errors"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types/base"
)

// SignalConfig configures one entry of the step `signals:` map. It accepts the
// short form (an exit code or a list of exit codes):
//
//	restart: 75
//	restart: [75, 76]
//
// and the long form with options:
//
//	restart:
//	  exit_codes: [75, 76]
//	  max_restarts: 5
type SignalConfig struct {
	ExitCodes base.IntOrSlice `yaml:"exit_codes"`
	// MaxRestarts is the number of restarts after the first run (restart signal only).
	MaxRestarts *int `yaml:"max_restarts,omitempty"`
}

// signalConfigLong is used to decode the long form without recursing into UnmarshalYAML.
type signalConfigLong SignalConfig

// UnmarshalYAML implements the Unmarshaler interface.
func (s *SignalConfig) UnmarshalYAML(unmarshal func(any) error) error {
	var codes base.IntOrSlice
	if err := unmarshal(&codes); err == nil {
		*s = SignalConfig{ExitCodes: codes}
		return nil
	}

	var long signalConfigLong
	if err := unmarshal(&long); err != nil {
		return errors.New("signal must be an exit code, a list of exit codes or an object with `exit_codes`")
	}
	*s = SignalConfig(long)
	return nil
}

// MarshalYAML implements custom Yaml marshaling.
func (s SignalConfig) MarshalYAML() (any, error) {
	if s.MaxRestarts == nil {
		return s.ExitCodes, nil
	}
	return signalConfigLong(s), nil
}

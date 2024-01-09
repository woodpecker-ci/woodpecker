// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
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

// AuthError represents forge authentication error.
type AuthError struct {
	Err         string
	Description string
	URI         string
}

// Error implements error interface.
func (ae *AuthError) Error() string {
	err := ae.Err
	if ae.Description != "" {
		err += " " + ae.Description
	}
	if ae.URI != "" {
		err += " " + ae.URI
	}
	return err
}

// check interface
var _ error = new(AuthError)

var ErrNotImplemented = errors.New("not implemented")

type ErrIgnoreEvent struct {
	Event string
}

func (err *ErrIgnoreEvent) Error() string {
	return fmt.Sprintf("explicit ignored event '%s'", err.Event)
}

func (*ErrIgnoreEvent) Is(target error) bool {
	_, ok := target.(*ErrIgnoreEvent) //nolint:errorlint
	return ok
}

type ErrConfigNotFound struct {
	Configs []string
}

func (m *ErrConfigNotFound) Error() string {
	return fmt.Sprintf("configs not found: %s", strings.Join(m.Configs, ", "))
}

func (*ErrConfigNotFound) Is(target error) bool {
	_, ok := target.(*ErrConfigNotFound) //nolint:errorlint
	return ok
}

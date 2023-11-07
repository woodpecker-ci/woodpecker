// Copyright 2022 Woodpecker Authors
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

package pipeline

import "errors"

type ErrConfigNotFound struct {
	Msg string
}

func (e ErrConfigNotFound) Error() string {
	return e.Msg
}

func (e ErrConfigNotFound) Is(target error) bool {
	_, ok := target.(ErrConfigNotFound) //nolint:errorlint
	if !ok {
		_, ok = target.(*ErrConfigNotFound) //nolint:errorlint
	}
	return ok
}

type ErrBadRequest struct {
	Msg string
}

func (e ErrBadRequest) Error() string {
	return e.Msg
}

func (e ErrBadRequest) Is(target error) bool {
	_, ok := target.(ErrBadRequest) //nolint:errorlint
	if !ok {
		_, ok = target.(*ErrBadRequest) //nolint:errorlint
	}
	return ok
}

var ErrFiltered = errors.New("ignoring hook: 'when' filters filtered out all steps")

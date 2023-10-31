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

import (
	"errors"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"go.uber.org/multierr"
)

type ErrNotFound struct {
	Msg string
}

func (e ErrNotFound) Error() string {
	return e.Msg
}

func (e ErrNotFound) Is(target error) bool {
	_, ok := target.(ErrNotFound) //nolint:errorlint
	if !ok {
		_, ok = target.(*ErrNotFound) //nolint:errorlint
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

func ErrorToPipelineErrors(err error) []*model.PipelineError {
	var pipelineErrors []*model.PipelineError
	for _, _err := range multierr.Errors(err) {
		var err *model.PipelineError
		if errors.As(_err, &err) {
			pipelineErrors = append(pipelineErrors, err)
		} else {
			pipelineErrors = append(pipelineErrors, &model.PipelineError{
				Message: err.Error(),
				Type:    model.PipelineErrorTypeGeneral,
			})
		}
	}

	return pipelineErrors
}

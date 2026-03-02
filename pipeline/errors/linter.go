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

package errors

import (
	"errors"

	"go.uber.org/multierr"
)

type LinterErrorData struct {
	File  string `json:"file"`
	Field string `json:"field"`
}

type DeprecationErrorData struct {
	File  string `json:"file"`
	Field string `json:"field"`
	Docs  string `json:"docs"`
}

type BadHabitErrorData struct {
	File  string `json:"file"`
	Field string `json:"field"`
	Docs  string `json:"docs"`
}

func GetLinterData(e *PipelineError) *LinterErrorData {
	if e.Type != PipelineErrorTypeLinter {
		return nil
	}

	if data, ok := e.Data.(*LinterErrorData); ok {
		return data
	}

	return nil
}

func GetPipelineErrors(err error) []*PipelineError {
	var pipelineErrors []*PipelineError
	for _, _err := range multierr.Errors(err) {
		var err *PipelineError
		if errors.As(_err, &err) {
			pipelineErrors = append(pipelineErrors, err)
		} else {
			pipelineErrors = append(pipelineErrors, &PipelineError{
				Message: _err.Error(),
				Type:    PipelineErrorTypeGeneric,
			})
		}
	}

	return pipelineErrors
}

func HasBlockingErrors(err error) bool {
	if err == nil {
		return false
	}

	errs := GetPipelineErrors(err)

	for _, err := range errs {
		if !err.IsWarning {
			return true
		}
	}

	return false
}

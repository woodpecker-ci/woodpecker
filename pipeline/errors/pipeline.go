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

package errors

import (
	"fmt"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

type PipelineErrorType string

const (
	PipelineErrorTypeLinter      PipelineErrorType = "linter"      // some error with the config syntax
	PipelineErrorTypeDeprecation PipelineErrorType = "deprecation" // using some deprecated feature
	PipelineErrorTypeCompiler    PipelineErrorType = "compiler"    // some error with the config semantics
	PipelineErrorTypeGeneric     PipelineErrorType = "generic"     // some generic error
	PipelineErrorTypeBadHabit    PipelineErrorType = "bad_habit"   // some bad-habit error
)

type PipelineError struct {
	Type      PipelineErrorType `json:"type"`
	Message   string            `json:"message"`
	IsWarning bool              `json:"is_warning"`
	Data      any               `json:"data"`
}

func (e *PipelineError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

type ErrInvalidWorkflowSetup struct {
	Err  error
	Step *backend.Step
}

func (e *ErrInvalidWorkflowSetup) Error() string {
	if e.Step != nil {
		return fmt.Sprintf("error in workflow setup step '%s': %v", e.Step.Name, e.Err)
	}
	return fmt.Sprintf("error in workflow setup: %v", e.Err)
}

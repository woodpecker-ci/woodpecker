package types

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

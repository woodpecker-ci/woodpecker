package types

import "fmt"

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

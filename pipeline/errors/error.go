package errors

import (
	"errors"
	"fmt"

	"go.uber.org/multierr"
)

type PipelineErrorType string

const (
	PipelineErrorTypeLinter      PipelineErrorType = "linter"      // some error with the config syntax
	PipelineErrorTypeDeprecation PipelineErrorType = "deprecation" // using some deprecated feature
	PipelineErrorTypeCompiler    PipelineErrorType = "compiler"    // some error with the config semantics
	PipelineErrorTypeGeneral     PipelineErrorType = "general"
)

type PipelineError struct {
	Type      PipelineErrorType `json:"type"`
	Message   string            `json:"message"`
	IsWarning bool              `json:"is_warning"`
	Data      interface{}       `json:"data"`
}

func (e *PipelineError) Error() string {
	return fmt.Sprintf("[%s]: %s", e.Type, e.Message)
}

func GetPipelineErrors(err error) []*PipelineError {
	var pipelineErrors []*PipelineError
	for _, _err := range multierr.Errors(err) {
		var err *PipelineError
		if errors.As(_err, &err) {
			pipelineErrors = append(pipelineErrors, err)
		} else {
			pipelineErrors = append(pipelineErrors, &PipelineError{
				Message: err.Error(),
				Type:    PipelineErrorTypeGeneral,
			})
		}
	}

	return pipelineErrors
}

func HasBlockingErrors(err error) bool {
	errs := GetPipelineErrors(err)

	for _, err := range errs {
		if !err.IsWarning {
			return true
		}
	}

	return false
}

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
	PipelineErrorTypeGeneric     PipelineErrorType = "generic"     // some generic error
)

type PipelineError struct {
	Type      PipelineErrorType `json:"type"`
	Message   string            `json:"message"`
	IsWarning bool              `json:"is_warning"`
	Data      interface{}       `json:"data"`
}

type LinterErrorData struct {
	File  string `json:"file"`
	Field string `json:"field"`
}

type DeprecationErrorData struct {
	File  string `json:"file"`
	Field string `json:"field"`
	Docs  string `json:"docs"`
}

func (e *PipelineError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

func (e *PipelineError) GetLinterData() *LinterErrorData {
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

var (
	// ErrSkip is used as a return value when container execution should be
	// skipped at runtime. It is not returned as an error by any function.
	ErrSkip = errors.New("Skipped")

	// ErrCancel is used as a return value when the container execution receives
	// a cancellation signal from the context.
	ErrCancel = errors.New("Canceled")

	// ErrFiltered is used as a when all steps  are filtered out or the pipeline is skipped by skip-ci in the commit message
	ErrFiltered = errors.New("Filtered as no steps matched or skipped by commit message")
)

// An ExitError reports an unsuccessful exit.
type ExitError struct {
	Name string
	Code int
}

// Error returns the error message in string format.
func (e *ExitError) Error() string {
	return fmt.Sprintf("%s : exit code %d", e.Name, e.Code)
}

// An OomError reports the process received an OOMKill from the kernel.
type OomError struct {
	Name string
	Code int
}

// Error returns the error message in string format.
func (e *OomError) Error() string {
	return fmt.Sprintf("%s : received oom kill", e.Name)
}

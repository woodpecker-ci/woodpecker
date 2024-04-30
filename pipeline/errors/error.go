package errors

import (
	"errors"

	"go.uber.org/multierr"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors/types"
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

func GetLinterData(e *types.PipelineError) *LinterErrorData {
	if e.Type != types.PipelineErrorTypeLinter {
		return nil
	}

	if data, ok := e.Data.(*LinterErrorData); ok {
		return data
	}

	return nil
}

func GetPipelineErrors(err error) []*types.PipelineError {
	var pipelineErrors []*types.PipelineError
	for _, _err := range multierr.Errors(err) {
		var err *types.PipelineError
		if errors.As(_err, &err) {
			pipelineErrors = append(pipelineErrors, err)
		} else {
			pipelineErrors = append(pipelineErrors, &types.PipelineError{
				Message: _err.Error(),
				Type:    types.PipelineErrorTypeGeneric,
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

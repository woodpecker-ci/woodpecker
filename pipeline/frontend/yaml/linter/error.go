package linter

import (
	"errors"
	"fmt"

	"go.uber.org/multierr"
)

type LinterError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Warning bool   `json:"warning"` // This error is a just warning and does not prevent the pipeline from running
}

func (e *LinterError) Error() string {
	errs := multierr.Errors(e)
	if len(errs) == 1 {
		return fmt.Sprintf("linter error in %s: %s", e.Field, e.Message)
	}

	errStr := "Got multiple linter errors:\n"
	for _, err := range errs {
		errStr += "- " + err.Error() + "\n"
	}
	return errStr
}

func (e *LinterError) IsBlocking() bool {
	errs := multierr.Errors(e)
	for _, err := range errs {
		var linterError *LinterError
		if errors.As(err, &linterError) {
			if !linterError.Warning {
				return true
			}
		} else {
			return true
		}
	}

	return false
}

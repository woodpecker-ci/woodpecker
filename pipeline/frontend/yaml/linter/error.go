package linter

import "fmt"

type LinterError struct {
	Message string         `json:"message"`
	Field   string         `json:"field"`
	Warning bool           `json:"warning"` // This error is a just warning and does not prevent the pipeline from running
	Errors  []*LinterError `json:"errors,omitempty"`
}

func (e *LinterError) Error() string {
	if e.Errors == nil || len(e.Errors) == 0 {
		return fmt.Sprintf("linter error in %s: %s", e.Field, e.Message)
	}

	errStr := "Got multiple linter errors:\n"
	for _, err := range e.Errors {
		errStr += "- " + err.Error() + "\n"
	}
	return errStr
}

func (e LinterError) Unwrap() []*LinterError {
	return e.Errors
}

func (e *LinterError) AddError(err *LinterError) {
	if e.Errors == nil {
		e.Errors = make([]*LinterError, 0)
	}

	if len(e.Errors) == 0 {
		e.Errors = append(e.Errors, &LinterError{
			Message: e.Message,
			Field:   e.Field,
			Warning: e.Warning,
		})
		e.Message = ""
		e.Field = ""
		e.Warning = false
	}

	if err.Errors != nil && len(err.Errors) > 0 {
		e.Errors = append(e.Errors, err.Errors...)
	} else {
		e.Errors = append(e.Errors, err)
	}
}

func (e *LinterError) IsBlocking() bool {
	for _, err := range e.Errors {
		if !err.Warning {
			return true
		}
	}

	return false
}

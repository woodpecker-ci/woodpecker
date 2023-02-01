package linter

import "fmt"

type LinterError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
	Warning bool   `json:"warning"` // This error is a just warning and does not prevent the pipeline from running
}

func (e *LinterError) Error() string {
	return fmt.Sprintf("linter error in %s: %s", e.Field, e.Message)
}

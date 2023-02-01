package linter

import "fmt"

type LinterError struct {
	Message string `json:"message"`
	Field   string `json:"field"`
}

func (e *LinterError) Error() string {
	return fmt.Sprintf("linter error in %s: %s", e.Field, e.Message)
}

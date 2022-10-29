package pipeline

import (
	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/multipart"
)

// Logger handles the process logging.
type Logger interface {
	Log(*backend.Step, multipart.Reader) error
}

// LogFunc type is an adapter to allow the use of an ordinary
// function for process logging.
type LogFunc func(*backend.Step, multipart.Reader) error

// Log calls f(step, r).
func (f LogFunc) Log(step *backend.Step, r multipart.Reader) error {
	return f(step, r)
}

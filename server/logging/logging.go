package logging

import (
	"context"
	"errors"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO(#742): write adapter for external pubsub provider

// ErrNotFound is returned when the log does not exist.
var ErrNotFound = errors.New("stream: not found")

// Handler defines a callback function for handling log entries.
type Handler func(...*model.LogEntry)

// Log defines a log multiplexer.
type Log interface {
	// Open opens the log.
	Open(c context.Context, stepID int64) error

	// Write writes the entry to the log.
	Write(c context.Context, stepID int64, entry *model.LogEntry) error

	// Tail tails the log.
	Tail(c context.Context, stepID int64, handler Handler) error

	// Close closes the log.
	Close(c context.Context, stepID int64) error
}

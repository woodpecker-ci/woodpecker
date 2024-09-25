// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logging

import (
	"context"
	"errors"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// ErrNotFound is returned when the log does not exist.
var ErrNotFound = errors.New("stream: not found")

// LogChan defines a channel type for receiving ordered batches of log entries.
type LogChan chan []*model.LogEntry

// Log defines a log multiplexer.
type Log interface {
	// Open opens the log.
	Open(c context.Context, stepID int64) error

	// Write writes the entry to the log.
	Write(c context.Context, stepID int64, entries []*model.LogEntry) error

	// Tail tails the log.
	Tail(c context.Context, stepID int64, handler LogChan) error

	// Close closes the log.
	Close(c context.Context, stepID int64) error
}

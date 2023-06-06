// Copyright 2022 Woodpecker Authors
// Copyright 2011 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline/shared"
)

// Identifies the type of line in the logs.
const (
	LogEntryStdout int = iota
	LogEntryStderr
	LogEntryExitCode
	LogEntryMetadata
	LogEntryProgress
)

// Line is a line of console output.
type LogEntry struct {
	StepUUID string `json:"step_uuid,omitempty"`
	Time     int64  `json:"time,omitempty"`
	Type     int    `json:"type,omitempty"`
	Line     int    `json:"line,omitempty"`
	Data     string `json:"data,omitempty"`
}

func (l *LogEntry) String() string {
	switch l.Type {
	case LogEntryExitCode:
		return fmt.Sprintf("[%s] exit code %s", l.StepUUID, l.Data)
	default:
		return fmt.Sprintf("[%s:L%v:%vs] %s", l.StepUUID, l.Line, l.Time, l.Data)
	}
}

// LineWriter sends logs to the client.
type LineWriter struct {
	peer     Peer
	stepUUID string
	num      int
	now      time.Time
	rep      *strings.Replacer
	lines    []*LogEntry
}

// NewLineWriter returns a new line reader.
func NewLineWriter(peer Peer, stepUUID string, secret ...string) *LineWriter {
	return &LineWriter{
		peer:     peer,
		stepUUID: stepUUID,
		now:      time.Now().UTC(),
		rep:      shared.NewSecretsReplacer(secret),
		lines:    nil,
	}
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if w.rep != nil {
		data = w.rep.Replace(data)
	}
	log.Trace().Str("step-uuid", w.stepUUID).Msgf("grpc write line: %s", data)

	line := &LogEntry{
		Data:     data,
		StepUUID: w.stepUUID,
		Time:     int64(time.Since(w.now).Seconds()),
		Type:     LogEntryStdout,
		Line:     w.num,
	}
	if err := w.peer.Log(context.Background(), line); err != nil {
		log.Error().Err(err).Str("step-uuid", w.stepUUID).Msg("fail to write pipeline log to peer")
	}
	w.num++

	w.lines = append(w.lines, line)
	return len(p), nil
}

// Lines returns the line history
func (w *LineWriter) Lines() []*LogEntry {
	return w.lines
}

// Clear clears the line history
func (w *LineWriter) Clear() {
	w.lines = w.lines[:0]
}

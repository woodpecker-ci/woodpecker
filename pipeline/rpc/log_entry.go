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
	StepID int64  `json:"step_id,omitempty"`
	Time   int64  `json:"time,omitempty"`
	Type   int    `json:"type,omitempty"`
	Line   int    `json:"line,omitempty"`
	Data   string `json:"data,omitempty"`
}

func (l *LogEntry) String() string {
	switch l.Type {
	case LogEntryExitCode:
		return fmt.Sprintf("[%d] exit code %s", l.StepID, l.Data)
	default:
		return fmt.Sprintf("[%d:L%v:%vs] %s", l.StepID, l.Line, l.Time, l.Data)
	}
}

// LineWriter sends logs to the client.
type LineWriter struct {
	peer   Peer
	stepID int64
	num    int
	now    time.Time
	rep    *strings.Replacer
	lines  []*LogEntry
}

// NewLineWriter returns a new line reader.
func NewLineWriter(peer Peer, stepID int64, secret ...string) *LineWriter {
	return &LineWriter{
		peer:   peer,
		stepID: stepID,
		now:    time.Now().UTC(),
		rep:    shared.NewSecretsReplacer(secret),
		lines:  nil,
	}
}

func (w *LineWriter) Write(p []byte) (n int, err error) {
	data := string(p)
	if w.rep != nil {
		data = w.rep.Replace(data)
	}
	log.Trace().Int64("step-id", w.stepID).Msgf("grpc write line: %s", data)

	line := &LogEntry{
		Data:   data,
		StepID: w.stepID,
		Time:   int64(time.Since(w.now).Seconds()),
		Type:   LogEntryStdout,
		// TODO: figure out a way to calc and set correct line number
		Line: w.num,
	}
	if err := w.peer.Log(context.Background(), line); err != nil {
		log.Error().Err(err).Msgf("fail to write pipeline log to peer '%d'", w.stepID)
	}
	w.num++

	// for _, part := range bytes.Split(p, []byte{'\n'}) {
	// 	line := &Line{
	// 		Out:  string(part),
	// 		Step: w.name,
	// 		Pos:  w.num,
	// 		Time: int64(time.Since(w.now).Seconds()),
	// 		Type: LineStdout,
	// 	}
	// 	w.peer.Log(context.Background(), w.id, line)
	// 	w.num++
	// }
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

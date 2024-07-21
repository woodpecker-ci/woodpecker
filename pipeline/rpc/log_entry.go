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
	"fmt"
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
	Data     []byte `json:"data,omitempty"`
}

func (l *LogEntry) String() string {
	switch l.Type {
	case LogEntryExitCode:
		return fmt.Sprintf("[%s] exit code %s", l.StepUUID, l.Data)
	default:
		return fmt.Sprintf("[%s:L%v:%vs] %s", l.StepUUID, l.Line, l.Time, l.Data)
	}
}

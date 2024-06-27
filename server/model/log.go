// Copyright 2021 Woodpecker Authors
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

package model

// LogEntryType identifies the type of line in the logs.
type LogEntryType int // @name	LogEntryType

const (
	LogEntryStdout LogEntryType = iota
	LogEntryStderr
	LogEntryExitCode
	LogEntryMetadata
	LogEntryProgress
)

type LogEntry struct {
	ID      int64        `json:"id"       xorm:"pk autoincr 'id'"`
	StepID  int64        `json:"step_id"  xorm:"INDEX 'step_id'"`
	Time    int64        `json:"time"     xorm:"'time'"`
	Line    int          `json:"line"     xorm:"'line'"`
	Data    []byte       `json:"data"     xorm:"LONGBLOB"`
	Created int64        `json:"-"        xorm:"created"`
	Type    LogEntryType `json:"type"     xorm:"'type'"`
} //	@name LogEntry

// TODO: store info what specific command the line belongs to (must be optional and impl. by backend)

func (LogEntry) TableName() string {
	return "log_entries"
}

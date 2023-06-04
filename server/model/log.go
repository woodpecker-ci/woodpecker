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

type LogEntry struct {
	ID      int64  `json:"id"       xorm:"pk autoincr 'log_id'"`
	StepID  int64  `json:"step_id"  xorm:"UNIQUE 'log_step_id'"`
	Time    int64  `json:"time"     xorm:"'log_time'"`
	Line    int    `json:"line"     xorm:"'log_line'"`
	Data    []byte `json:"data"     xorm:"LONGBLOB 'log_data'"`
	Created int64  `json:"-"        xorm:"created"`
	// TODO: should we add type from pipeline/rpc/line.go?
}

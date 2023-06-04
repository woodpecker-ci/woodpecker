// Copyright 2023 Woodpecker Authors
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

package migration

import (
	"encoding/json"

	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// old:
// - log_id
// - log_step_id
// - log_data

// new:
// - log_id
// - log_step_id
// - log_time
// - log_pos
// - log_data (raw text)
// - log_command (can be added later when we support executing command by command)

type oldLogs018 struct {
	ID     int64  `xorm:"pk autoincr 'log_id'"`
	StepID int64  `xorm:"UNIQUE 'log_step_id'"`
	Data   []byte `xorm:"LONGBLOB 'log_data'"`
}

func (oldLogs018) TableName() string {
	return "logs"
}

type oldLogEntry018 struct {
	Step string `json:"step,omitempty"`
	Time int64  `json:"time,omitempty"`
	Type int    `json:"type,omitempty"`
	Pos  int    `json:"pos,omitempty"`
	Out  string `json:"out,omitempty"`
}

var alterLogsTable = task{
	name: "alter-logs-table",
	fn: func(sess *xorm.Session) error {
		// make sure old logs table exists
		if err := sess.Sync(new(oldLogs018)); err != nil {
			return err
		}

		if err := renameTable(sess, "logs", "old_logs"); err != nil {
			return err
		}

		if err := sess.Sync(new(model.LogEntry)); err != nil {
			return err
		}

		// TODO: copy data over from old_logs to logs
		page := 0
		for {
			var logs []*oldLogs018
			err := sess.Limit(10, page*10).Find(&logs)
			if err != nil {
				return err
			}

			for _, l := range logs {

				logEntries := []*oldLogEntry018{}
				if err := json.Unmarshal(l.Data, &logEntries); err != nil {
					return err
				}

				time := int64(0)
				for _, logEntry := range logEntries {

					if logEntry.Time > time {
						time = logEntry.Time
					}

					log := &model.LogEntry{
						ID:     l.ID,
						StepID: l.StepID,
						Data:   []byte(logEntry.Out),
						Line:   logEntry.Pos,
						Time:   time,
					}

					if _, err := sess.Insert(log); err != nil {
						return err
					}
				}
			}

			if len(logs) < 10 {
				break
			}

			page++
		}

		return sess.DropTable("old_logs")
	},
}

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

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type oldLogs019 struct {
	ID     int64  `xorm:"pk autoincr 'log_id'"`
	StepID int64  `xorm:"UNIQUE 'log_step_id'"`
	Data   []byte `xorm:"LONGBLOB 'log_data'"`
}

func (oldLogs019) TableName() string {
	return "logs"
}

type oldLogEntry019 struct {
	Step string `json:"step,omitempty"`
	Time int64  `json:"time,omitempty"`
	Type int    `json:"type,omitempty"`
	Pos  int    `json:"pos,omitempty"`
	Out  string `json:"out,omitempty"`
}

type newLogEntry019 struct {
	ID      int64  `json:"id"       xorm:"pk autoincr 'id'"`
	StepID  int64  `json:"step_id"  xorm:"'step_id'"`
	Time    int64  `json:"time"`
	Line    int    `json:"line"`
	Data    []byte `json:"data"     xorm:"LONGBLOB"`
	Created int64  `json:"-"        xorm:"created"`
	Type    int    `json:"type"`
}

func (newLogEntry019) TableName() string {
	return "log_entries"
}

var initLogsEntriesTable = task{
	name:     "init-log_entries",
	required: true,
	fn: func(sess *xorm.Session) error {
		return sess.Sync(new(newLogEntry019))
	},
}

var migrateLogs2LogEntries = task{
	name:     "migrate-logs-to-log_entries",
	required: true,
	fn: func(sess *xorm.Session) error {
		// make sure old logs table exists
		if exist, err := sess.IsTableExist(new(oldLogs019)); !exist || err != nil {
			return err
		}

		if err := sess.Sync(new(oldLogs019)); err != nil {
			return err
		}

		log.Info().Msg("migrate-logs-to-log_entries: start migration of logs")

		page := 0
		for {
			var logs []*oldLogs019
			err := sess.Limit(10, page*10).Find(&logs)
			if err != nil {
				return err
			}

			log.Info().Msgf("migrate-logs-to-log_entries: start page %d", page)

			for _, l := range logs {

				logEntries := []*oldLogEntry019{}
				if err := json.Unmarshal(l.Data, &logEntries); err != nil {
					return err
				}

				time := int64(0)
				for _, logEntry := range logEntries {

					if logEntry.Time > time {
						time = logEntry.Time
					}

					log := &model.LogEntry{
						StepID: l.StepID,
						Data:   []byte(logEntry.Out),
						Line:   logEntry.Pos,
						Time:   time,
						Type:   model.LogEntryType(logEntry.Type),
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

		return sess.DropTable("logs")
	},
}

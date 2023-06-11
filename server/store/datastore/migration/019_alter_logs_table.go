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
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

// maxDefaultSqliteItems set the threshold at witch point the migration will fail by default
var maxDefaultSqliteItems019 = 5000

// perPage019 set the size of the slice to read per page
var perPage019 = 100

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
	required: false,
	engineFn: func(e *xorm.Engine) error {
		// make sure old logs table exists
		if exist, err := e.IsTableExist(new(oldLogs019)); !exist || err != nil {
			return err
		}

		// first we check if we have just 1000 entries to migrate
		toMigrate, err := e.Count(new(oldLogs019))
		if err != nil {
			return err
		}
		allowLongMigration, _ := strconv.ParseBool(os.Getenv("WOODPECKER_ALLOW_LONG_MIGRATION"))
		if toMigrate > int64(maxDefaultSqliteItems019) && !allowLongMigration {
			return fmt.Errorf("migrating logs to log_entries is skipped, as we have %d entries to convert. set 'WOODPECKER_ALLOW_LONG_MIGRATION' to 'true' to migrate anyway", toMigrate)
		}

		if err := e.Sync(new(oldLogs019)); err != nil {
			return err
		}

		page := 0
		logs := make([]*oldLogs019, 0, perPage019)
		logEntries := make([]*oldLogEntry019, 0, 50)
		sigterm := abool.New()
		_ = utils.WithContextSigtermCallback(context.Background(), func() {
			println("ctrl+c received, terminating process")
			sigterm.Set()
		})

		for {
			if sigterm.IsSet() {
				return fmt.Errorf("migration 'migrate-logs-to-log_entries' successfully aborted")
			}

			sess := e.NewSession().NoCache()
			defer sess.Close()
			if err := sess.Begin(); err != nil {
				return err
			}
			logs = logs[:0]

			err := sess.Limit(perPage019).Find(&logs)
			if err != nil {
				return err
			}

			log.Trace().Msgf("migrate-logs-to-log_entries: process page %d", page)

			for _, l := range logs {
				logEntries = logEntries[:0]
				if err := json.Unmarshal(l.Data, &logEntries); err != nil {
					return err
				}

				time := int64(0)
				for _, logEntry := range logEntries {

					if logEntry.Time > time {
						time = logEntry.Time
					}

					log := &newLogEntry019{
						StepID: l.StepID,
						Data:   []byte(logEntry.Out),
						Line:   logEntry.Pos,
						Time:   time,
						Type:   logEntry.Type,
					}

					if _, err := sess.Insert(log); err != nil {
						return err
					}
				}

				if _, err := sess.Delete(l); err != nil {
					return err
				}
			}

			if err := sess.Commit(); err != nil {
				return err
			}

			if len(logs) < perPage019 {
				break
			}

			runtime.GC()
			page++
		}

		return e.DropTables("logs")
	},
}

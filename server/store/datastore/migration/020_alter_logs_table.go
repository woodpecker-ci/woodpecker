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
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/tevino/abool/v2"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/shared/utils"
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
	ID      int64 `xorm:"pk autoincr 'id'"`
	StepID  int64 `xorm:"'step_id'"`
	Time    int64
	Line    int
	Data    []byte `xorm:"LONGBLOB"`
	Created int64  `xorm:"created"`
	Type    int
}

func (newLogEntry019) TableName() string {
	return "log_entries"
}

var initLogsEntriesTable = xormigrate.Migration{
	ID: "init-log_entries",
	MigrateSession: func(sess *xorm.Session) error {
		return sess.Sync(new(newLogEntry019))
	},
}

var migrateLogs2LogEntries = xormigrate.Migration{
	ID:   "migrate-logs-to-log_entries",
	Long: true,
	Migrate: func(e *xorm.Engine) error {
		// make sure old logs table exists
		if exist, err := e.IsTableExist(new(oldLogs019)); !exist || err != nil {
			return err
		}

		if err := e.Sync(new(oldLogs019)); err != nil {
			return err
		}

		hasJSONErrors := false

		page := 0
		offset := 0
		logs := make([]*oldLogs019, 0, perPage019)
		logEntries := make([]*oldLogEntry019, 0, 50)

		sigterm := abool.New()
		ctx, cancelCtx := context.WithCancelCause(context.Background())
		defer cancelCtx(nil)
		_ = utils.WithContextSigtermCallback(ctx, func() {
			log.Info().Msg("ctrl+c received, stopping current migration")
			sigterm.Set()
		})

		for {
			if sigterm.IsSet() {
				return fmt.Errorf("migration 'migrate-logs-to-log_entries' gracefully aborted")
			}

			sess := e.NewSession().NoCache()
			defer sess.Close()
			if err := sess.Begin(); err != nil {
				return err
			}
			logs = logs[:0]

			err := sess.Limit(perPage019, offset).Find(&logs)
			if err != nil {
				return err
			}

			log.Trace().Msgf("migrate-logs-to-log_entries: process page %d", page)

			for _, l := range logs {
				logEntries = logEntries[:0]
				if err := json.Unmarshal(l.Data, &logEntries); err != nil {
					hasJSONErrors = true
					offset++
					continue
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

		if hasJSONErrors {
			return fmt.Errorf("skipped some logs as json could not be deserialized for them")
		}

		return e.DropTables("logs")
	},
}

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

package datastore

import (
	"errors"

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// Maximum number of records to store in one PostgreSQL statement.
// Too large a value results in `pq: got XX parameters but PostgreSQL only supports 65535 parameters`.
const pgBatchSize = 1000

// LogFind returns the log entries of a step in the order the agent produced
// them, which is the order their line numbers carry.
//
// An agent that reconnects mid-step can resend a batch the server already
// stored, so a line number can occur more than once. Keeping one row per line
// number matches what a UNIQUE(step_id, line) index would enforce.
func (s storage) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	var logEntries []*model.LogEntry
	if err := s.engine.Asc("line").Where("step_id = ?", step.ID).Find(&logEntries); err != nil {
		return nil, err
	}

	return dedupLogEntries(logEntries), nil
}

// dedupLogEntries keeps the first entry of every line number. The input must
// be ordered by line, which puts the repeats next to each other.
func dedupLogEntries(logEntries []*model.LogEntry) []*model.LogEntry {
	deduped := make([]*model.LogEntry, 0, len(logEntries))

	for _, logEntry := range logEntries {
		if len(deduped) > 0 && logEntry.Line == deduped[len(deduped)-1].Line {
			continue
		}

		deduped = append(deduped, logEntry)
	}

	return deduped
}

func (s storage) LogAppend(_ *model.Step, logEntries []*model.LogEntry) error {
	var errs error

	// TODO: adapted from slices.Chunk(); switch to it in Go 1.23+
	for i := 0; i < len(logEntries); i += pgBatchSize {
		end := min(pgBatchSize, len(logEntries[i:]))
		chunk := logEntries[i : i+end]

		if err := wrapInsert(s.engine.Insert(chunk)); err != nil {
			log.Error().Err(err).Msg("could not store log entries to db")
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (s storage) LogDelete(step *model.Step) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	return logDelete(sess, step.ID)
}

func logDelete(sess *xorm.Session, stepID int64) error {
	_, err := sess.Where("step_id = ?", stepID).Delete(new(model.LogEntry))
	return err
}

func (s storage) StepFinished(_ *model.Step) {}

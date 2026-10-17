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
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// Maximum number of records to store in one PostgreSQL statement.
// Too large a value results in `pq: got XX parameters but PostgreSQL only supports 65535 parameters`.
const pgBatchSize = 1000

// LogFind returns the log entries of a step in the order the agent produced
// them, which is the order their line numbers carry.
func (s storage) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	var logEntries []*model.LogEntry
	return logEntries, s.engine.Asc("line").Where("step_id = ?", step.ID).Find(&logEntries)
}

// LogAppend stores the log entries. Entries whose line is already stored for
// the step are skipped, so an agent can resend a batch that was only partly
// stored. It returns types.ErrInsertDuplicateDetected if every entry was
// already stored.
func (s storage) LogAppend(_ *model.Step, logEntries []*model.LogEntry) error {
	var errs error
	stored := 0

	// TODO: adapted from slices.Chunk(); switch to it in Go 1.23+
	for i := 0; i < len(logEntries); i += pgBatchSize {
		end := min(pgBatchSize, len(logEntries[i:]))
		chunk := logEntries[i : i+end]

		err := wrapInsert(s.engine.Insert(chunk))
		if errors.Is(err, types.ErrInsertDuplicateDetected) {
			// the statement was rolled back, retry entry by entry and skip the stored ones
			n, err := s.logAppendSkipStored(chunk)
			stored += n
			errs = errors.Join(errs, err)
			continue
		}
		if err != nil {
			log.Error().Err(err).Msg("could not store log entries to db")
			errs = errors.Join(errs, err)
			continue
		}
		stored += len(chunk)
	}

	if errs == nil && stored == 0 && len(logEntries) > 0 {
		return types.ErrInsertDuplicateDetected
	}
	return errs
}

func (s storage) logAppendSkipStored(logEntries []*model.LogEntry) (stored int, errs error) {
	for _, logEntry := range logEntries {
		err := wrapInsert(s.engine.Insert(logEntry))
		if errors.Is(err, types.ErrInsertDuplicateDetected) {
			continue
		}
		if err != nil {
			log.Error().Err(err).Msg("could not store log entry to db")
			errs = errors.Join(errs, err)
			continue
		}
		stored++
	}
	return stored, errs
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

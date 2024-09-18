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
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// Maximum number of records to store in one PostgreSQL statement.
// Too large a value results in `pq: got XX parameters but PostgreSQL only supports 65535 parameters`.
const pgBatchSize = 1000

func (s storage) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	var logEntries []*model.LogEntry
	return logEntries, s.engine.Asc("id").Where("step_id = ?", step.ID).Find(&logEntries)
}

func (s storage) LogAppend(_ *model.Step, logEntries []*model.LogEntry) error {
	var err error

	// TODO: adapted from slices.Chunk(); switch to it in Go 1.23+
	for i := 0; i < len(logEntries); i += pgBatchSize {
		end := min(pgBatchSize, len(logEntries[i:]))
		chunk := logEntries[i : i+end]

		if _, err = s.engine.Insert(chunk); err != nil {
			log.Error().Err(err).Msg("could not store log entries to db")
		}
	}

	return err
}

func (s storage) LogDelete(step *model.Step) error {
	_, err := s.engine.Where("step_id = ?", step.ID).Delete(new(model.LogEntry))
	return err
}

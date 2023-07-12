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
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

func (s storage) LogFind(step *model.Step) ([]*model.LogEntry, error) {
	var logEntries []*model.LogEntry
	if err := s.engine.Asc("id").Where("step_id = ?", step.ID).Find(&logEntries); err != nil {
		return nil, err
	}
	deCompressLogs(logEntries...)
	return logEntries, nil
}

func compressLogs(logEntries ...*model.LogEntry) {
	for i := range logEntries {
		logEntries[i].Data = utils.ZStdCompress(logEntries[i].Data)
		logEntries[i].Compr = "zstd"
	}
}

func deCompressLogs(logEntries ...*model.LogEntry) {
	for i := range logEntries {
		if logEntries[i].Compr == "zstd" {
			logEntries[i].Data, _ = utils.ZStdDecompress(logEntries[i].Data)
		}
	}
}

func (s storage) LogSave(step *model.Step, logEntries []*model.LogEntry) error {
	compressLogs(logEntries...)
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	for _, logEntry := range logEntries {
		if logEntry.StepID != step.ID {
			return fmt.Errorf("got a log-entry with step id '%d' but expected '%d'", logEntry.StepID, step.ID)
		}
		if _, err := sess.Insert(logEntry); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) LogAppend(logEntry *model.LogEntry) error {
	compressLogs(logEntry)
	_, err := s.engine.Insert(logEntry)
	return err
}

func (s storage) LogDelete(step *model.Step) error {
	_, err := s.engine.Where("step_id = ?", step.ID).Delete(new(model.LogEntry))
	return err
}

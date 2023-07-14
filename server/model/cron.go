// Copyright 2022 Woodpecker Authors
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

import (
	"fmt"

	"github.com/robfig/cron"
)

type Cron struct {
	ID        int64  `json:"id"                  xorm:"pk autoincr"`
	Name      string `json:"name"                xorm:"UNIQUE(s) INDEX"`
	RepoID    int64  `json:"repo_id"             xorm:"repo_id UNIQUE(s) INDEX"`
	CreatorID int64  `json:"creator_id"          xorm:"creator_id INDEX"`
	NextExec  int64  `json:"next_exec"`
	Schedule  string `json:"schedule"            xorm:"NOT NULL"` //	@weekly,	3min, ...
	Created   int64  `json:"created_at"          xorm:"created NOT NULL DEFAULT 0"`
	Branch    string `json:"branch"`
} //	@name Cron

// TableName returns the database table name for xorm
func (Cron) TableName() string {
	return "crons"
}

// Validate a cron
func (c *Cron) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is required")
	}

	if c.Schedule == "" {
		return fmt.Errorf("schedule is required")
	}

	_, err := cron.Parse(c.Schedule)
	if err != nil {
		return fmt.Errorf("can't parse schedule: %w", err)
	}

	return nil
}

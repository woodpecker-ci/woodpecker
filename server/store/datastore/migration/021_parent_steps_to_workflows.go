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

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type oldStep021 struct {
	ID         int64             `xorm:"pk autoincr 'step_id'"`
	PipelineID int64             `xorm:"UNIQUE(s) INDEX 'step_pipeline_id'"`
	PID        int               `xorm:"UNIQUE(s) 'step_pid'"`
	PPID       int               `xorm:"step_ppid"`
	Name       string            `xorm:"step_name"`
	State      model.StatusValue `xorm:"step_state"`
	Error      string            `xorm:"TEXT 'step_error'"`
	Started    int64             `xorm:"step_started"`
	Stopped    int64             `xorm:"step_stopped"`
	AgentID    int64             `xorm:"step_agent_id"`
	Platform   string            `xorm:"step_platform"`
	Environ    map[string]string `xorm:"json 'step_environ'"`
}

func (oldStep021) TableName() string {
	return "steps"
}

var parentStepsToWorkflows = xormigrate.Migration{
	ID: "parent-steps-to-workflows",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(workflowV031)); err != nil {
			return err
		}
		// make sure the columns exist before removing them
		if err := sess.Sync(new(oldStep021)); err != nil {
			return err
		}

		var parentSteps []*oldStep021
		err := sess.Where("step_ppid = ?", 0).Find(&parentSteps)
		if err != nil {
			return err
		}

		for _, p := range parentSteps {
			asWorkflow := &workflowV031{
				PipelineID: p.PipelineID,
				PID:        p.PID,
				Name:       p.Name,
				State:      p.State,
				Error:      p.Error,
				Started:    p.Started,
				Stopped:    p.Stopped,
				AgentID:    p.AgentID,
				Platform:   p.Platform,
				Environ:    p.Environ,
			}

			_, err = sess.Insert(asWorkflow)
			if err != nil {
				return err
			}

			_, err = sess.Delete(&oldStep021{ID: p.ID})
			if err != nil {
				return err
			}
		}

		return dropTableColumns(sess, "steps", "step_agent_id", "step_platform", "step_environ")
	},
}

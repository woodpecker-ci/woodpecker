// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var addCompilePhase = xormigrate.Migration{
	ID: "add-compile-phase",
	MigrateSession: func(sess *xorm.Session) error {
		type workflows struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			// new phase field
			Phase string `xorm:"phase"`

			// new compile result field, holding what a compile workflow emitted
			// until the last one has finished and the merge can run
			CompileResult []byte `xorm:"json 'compile_result'"`
		}

		type pipelines struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			// new compile state field, which makes the merge exactly-once
			CompileState string `xorm:"compile_state"`
		}

		type pipelineConfigs struct {
			ConfigID   int64 `xorm:"UNIQUE(s) NOT NULL 'config_id'"`
			PipelineID int64 `xorm:"UNIQUE(s) NOT NULL 'pipeline_id'"`

			// new source and effective fields, distinguishing what a pipeline
			// was built from and what it actually ran
			Source    bool `xorm:"NOT NULL DEFAULT false 'source'"`
			Effective bool `xorm:"NOT NULL DEFAULT false 'effective'"`
		}

		if err := sess.Sync(new(workflows), new(pipelines), new(pipelineConfigs)); err != nil {
			return err
		}

		// Every workflow that exists today is an ordinary one.
		if _, err := sess.Exec("UPDATE workflows SET phase = ?", model.WorkflowPhaseRun); err != nil {
			return err
		}

		// No pipeline that exists today has a compile phase, so what it was
		// built from is also what it ran.
		_, err := sess.Exec("UPDATE pipeline_configs SET source = ?, effective = ?", true, true)
		return err
	},
}

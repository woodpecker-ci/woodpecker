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
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) StepLoad(id int64) (*model.Step, error) {
	step := new(model.Step)
	return step, wrapGet(s.engine.ID(id).Get(step))
}

func (s storage) StepFind(pipeline *model.Pipeline, pid int) (*model.Step, error) {
	step := &model.Step{
		PipelineID: pipeline.ID,
		PID:        pid,
	}
	return step, wrapGet(s.engine.Get(step))
}

func (s storage) StepChild(pipeline *model.Pipeline, ppid int, child string) (*model.Step, error) {
	step := &model.Step{
		PipelineID: pipeline.ID,
		PPID:       ppid,
		Name:       child,
	}
	return step, wrapGet(s.engine.Get(step))
}

func (s storage) StepList(pipeline *model.Pipeline) ([]*model.Step, error) {
	stepList := make([]*model.Step, 0)
	return stepList, s.engine.
		Where("step_pipeline_id = ?", pipeline.ID).
		OrderBy("step_pid").
		Find(&stepList)
}

func (s storage) StepListWorkflow(workflow *model.Workflow) ([]*model.Step, error) {
	return s.stepListWorkflow(s.engine.NewSession(), workflow)
}

func (s storage) stepListWorkflow(sess *xorm.Session, workflow *model.Workflow) ([]*model.Step, error) {
	stepList := make([]*model.Step, 0)
	return stepList, sess.
		Where("step_workflow_id = ?", workflow.ID).
		OrderBy("step_pid").
		Find(&stepList)
}

func (s storage) stepCreate(sess *xorm.Session, steps []*model.Step) error {
	if err := sess.Begin(); err != nil {
		return err
	}

	for i := range steps {
		// only Insert on single object ref set auto created ID back to object
		if _, err := sess.Insert(steps[i]); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) StepUpdate(step *model.Step) error {
	_, err := s.engine.ID(step.ID).AllCols().Update(step)
	return err
}

func (s storage) StepClear(pipeline *model.Pipeline) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Where("step_pipeline_id = ?", pipeline.ID).Delete(new(model.Step)); err != nil {
		return err
	}

	if _, err := sess.Where("workflow_pipeline_id = ?", pipeline.ID).Delete(new(model.Workflow)); err != nil {
		return err
	}

	return sess.Commit()
}

func deleteStep(sess *xorm.Session, stepID int64) error {
	if _, err := sess.Where("log_step_id = ?", stepID).Delete(new(model.Logs)); err != nil {
		return err
	}
	_, err := sess.ID(stepID).Delete(new(model.Step))
	return err
}

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
	"xorm.io/builder"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func (s storage) StepLoad(id int64) (*model.Step, error) {
	step := new(model.Step)
	return step, wrapGet(s.engine.ID(id).Get(step))
}

func (s storage) StepFind(pipeline *model.Pipeline, pid int) (*model.Step, error) {
	step := new(model.Step)
	return step, wrapGet(s.engine.Where(
		builder.Eq{"pipeline_id": pipeline.ID, "pid": pid},
	).Get(step))
}

func (s storage) StepByUUID(uuid string) (*model.Step, error) {
	step := new(model.Step)
	return step, wrapGet(s.engine.Where(
		builder.Eq{"uuid": uuid},
	).Get(step))
}

func (s storage) StepChild(pipeline *model.Pipeline, ppid int, child string) (*model.Step, error) {
	step := new(model.Step)
	return step, wrapGet(s.engine.Where(
		builder.Eq{"pipeline_id": pipeline.ID, "ppid": ppid, "name": child},
	).Get(step))
}

func (s storage) StepList(pipeline *model.Pipeline) ([]*model.Step, error) {
	stepList := make([]*model.Step, 0)
	return stepList, s.engine.
		Where("pipeline_id = ?", pipeline.ID).
		OrderBy("pid").
		Find(&stepList)
}

func (s storage) StepListFromWorkflowFind(workflow *model.Workflow) ([]*model.Step, error) {
	return s.stepListWorkflow(s.engine.NewSession(), workflow)
}

func (s storage) stepListWorkflow(sess *xorm.Session, workflow *model.Workflow) ([]*model.Step, error) {
	stepList := make([]*model.Step, 0)
	return stepList, sess.
		Where("pipeline_id = ?", workflow.PipelineID).
		Where("ppid = ?", workflow.PID).
		OrderBy("pid").
		Find(&stepList)
}

func (s storage) stepCreate(sess *xorm.Session, steps []*model.Step) error {
	for i := range steps {
		// only Insert on single object ref set auto created ID back to object
		if _, err := sess.Insert(steps[i]); err != nil {
			return err
		}
	}
	return nil
}

func (s storage) StepUpdate(step *model.Step) error {
	_, err := s.engine.ID(step.ID).AllCols().Update(step)
	return err
}

func deleteStep(sess *xorm.Session, stepID int64) error {
	if _, err := sess.Where("id = ?", stepID).Delete(new(model.LogEntry)); err != nil {
		return err
	}
	return wrapDelete(sess.ID(stepID).Delete(new(model.Step)))
}

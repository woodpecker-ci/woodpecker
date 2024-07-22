// Copyright 2023 Woodpecker Authors
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

package datastore

import (
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func (s storage) WorkflowGetTree(pipeline *model.Pipeline) ([]*model.Workflow, error) {
	sess := s.engine.NewSession()
	wfList, err := s.workflowList(sess, pipeline)
	if err != nil {
		return nil, err
	}

	for _, wf := range wfList {
		wf.Children, err = s.stepListWorkflow(sess, wf)
		if err != nil {
			return nil, err
		}
	}

	return wfList, sess.Commit()
}

func (s storage) WorkflowsCreate(workflows []*model.Workflow) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := s.workflowsCreate(sess, workflows); err != nil {
		return err
	}

	return sess.Commit()
}

func (s storage) workflowsCreate(sess *xorm.Session, workflows []*model.Workflow) error {
	for i := range workflows {
		// only Insert on single object ref set auto created ID back to object
		if err := s.stepCreate(sess, workflows[i].Children); err != nil {
			return err
		}
		if _, err := sess.Insert(workflows[i]); err != nil {
			return err
		}
	}
	return nil
}

// WorkflowsReplace performs an atomic replacement of workflows and associated steps by deleting all existing workflows and steps and inserting the new ones.
func (s storage) WorkflowsReplace(pipeline *model.Pipeline, workflows []*model.Workflow) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := s.workflowsDelete(sess, pipeline.ID); err != nil {
		return err
	}

	if err := s.workflowsCreate(sess, workflows); err != nil {
		return err
	}

	return sess.Commit()
}

func (s storage) workflowsDelete(sess *xorm.Session, pipelineID int64) error {
	// delete related steps
	for startSteps := 0; ; startSteps += perPage {
		stepIDs := make([]int64, 0, perPage)
		if err := sess.Limit(perPage, startSteps).Table("steps").Cols("id").Where("pipeline_id = ?", pipelineID).Find(&stepIDs); err != nil {
			return err
		}
		if len(stepIDs) == 0 {
			break
		}

		for i := range stepIDs {
			if err := deleteStep(sess, stepIDs[i]); err != nil {
				return err
			}
		}
	}

	_, err := sess.Where("pipeline_id = ?", pipelineID).Delete(new(model.Workflow))
	return err
}

func (s storage) WorkflowList(pipeline *model.Pipeline) ([]*model.Workflow, error) {
	return s.workflowList(s.engine.NewSession(), pipeline)
}

// workflowList lists workflows without child steps.
func (s storage) workflowList(sess *xorm.Session, pipeline *model.Pipeline) ([]*model.Workflow, error) {
	var wfList []*model.Workflow
	err := sess.Where("pipeline_id = ?", pipeline.ID).
		OrderBy("pid").
		Find(&wfList)
	if err != nil {
		return nil, err
	}

	return wfList, nil
}

func (s storage) WorkflowLoad(id int64) (*model.Workflow, error) {
	workflow := new(model.Workflow)
	return workflow, wrapGet(s.engine.ID(id).Get(workflow))
}

func (s storage) WorkflowUpdate(workflow *model.Workflow) error {
	_, err := s.engine.ID(workflow.ID).AllCols().Update(workflow)
	return err
}

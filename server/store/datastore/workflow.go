package datastore

import (
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
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

	for i := range workflows {
		// only Insert on single object ref set auto created ID back to object
		if err := s.stepCreate(sess, workflows[i].Children); err != nil {
			return err
		}
		if _, err := sess.Insert(workflows[i]); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) WorkflowList(pipeline *model.Pipeline) ([]*model.Workflow, error) {
	return s.workflowList(s.engine.NewSession(), pipeline)
}

// workflowList lists workflows without child steps
func (s storage) workflowList(sess *xorm.Session, pipeline *model.Pipeline) ([]*model.Workflow, error) {
	var wfList []*model.Workflow
	err := sess.Where("workflow_pipeline_id = ?", pipeline.ID).
		OrderBy("workflow_pid").
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

func (s storage) WorkflowFind(pipeline *model.Pipeline, pid int) (*model.Workflow, error) {
	wf := &model.Workflow{
		PipelineID: pipeline.ID,
		PID:        pid,
	}
	return wf, wrapGet(s.engine.Get(wf))
}

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
	"time"

	"xorm.io/builder"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func (s storage) GetPipeline(id int64) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{}
	return pipeline, wrapGet(s.engine.ID(id).Get(pipeline))
}

func (s storage) GetPipelineNumber(repo *model.Repo, num int64) (*model.Pipeline, error) {
	pipeline := new(model.Pipeline)
	return pipeline, wrapGet(s.engine.Where(
		builder.Eq{"repo_id": repo.ID, "number": num},
	).Get(pipeline))
}

func (s storage) GetPipelineLast(repo *model.Repo, branch string) (*model.Pipeline, error) {
	pipeline := new(model.Pipeline)
	return pipeline, wrapGet(s.engine.
		Desc("number").
		Where(builder.Eq{"repo_id": repo.ID, "branch": branch, "event": model.EventPush}).
		Get(pipeline))
}

func (s storage) GetPipelineLastBefore(repo *model.Repo, branch string, num int64) (*model.Pipeline, error) {
	pipeline := new(model.Pipeline)
	return pipeline, wrapGet(s.engine.
		Desc("number").
		Where(builder.Lt{"id": num}.
			And(builder.Eq{"repo_id": repo.ID, "branch": branch})).
		Get(pipeline))
}

func (s storage) GetPipelineList(repo *model.Repo, p *model.ListOptions, f *model.PipelineFilter) ([]*model.Pipeline, error) {
	pipelines := make([]*model.Pipeline, 0, 16)

	cond := builder.NewCond().And(builder.Eq{"repo_id": repo.ID})

	if f != nil {
		if f.After != 0 {
			cond = cond.And(builder.Gt{"created": f.After})
		}

		if f.Before != 0 {
			cond = cond.And(builder.Lt{"created": f.Before})
		}
	}

	return pipelines, s.paginate(p).Where(cond).
		Desc("number").
		Find(&pipelines)
}

// GetActivePipelineList get all pipelines that are pending, running or blocked.
func (s storage) GetActivePipelineList(repo *model.Repo) ([]*model.Pipeline, error) {
	pipelines := make([]*model.Pipeline, 0)
	query := s.engine.
		Where("repo_id = ?", repo.ID).
		In("status", model.StatusPending, model.StatusRunning, model.StatusBlocked).
		Desc("number")
	return pipelines, query.Find(&pipelines)
}

func (s storage) GetPipelineCount() (int64, error) {
	return s.engine.Count(new(model.Pipeline))
}

func (s storage) CreatePipeline(pipeline *model.Pipeline, stepList ...*model.Step) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	repoExist, err := sess.Where("id = ?", pipeline.RepoID).Exist(&model.Repo{})
	if err != nil {
		return err
	}

	if !repoExist {
		return ErrorRepoNotExist{RepoID: pipeline.RepoID}
	}

	// calc pipeline number
	var number int64
	if _, err := sess.Select("MAX(number)").
		Table(new(model.Pipeline)).
		Where("repo_id = ?", pipeline.RepoID).
		Get(&number); err != nil {
		return err
	}
	pipeline.Number = number + 1

	pipeline.Created = time.Now().UTC().Unix()
	// only Insert set auto created ID back to object
	if _, err := sess.Insert(pipeline); err != nil {
		return err
	}

	for i := range stepList {
		stepList[i].PipelineID = pipeline.ID
		// only Insert set auto created ID back to object
		if _, err := sess.Insert(stepList[i]); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) UpdatePipeline(pipeline *model.Pipeline) error {
	_, err := s.engine.ID(pipeline.ID).AllCols().Update(pipeline)
	return err
}

func (s storage) DeletePipeline(pipeline *model.Pipeline) error {
	return s.deletePipeline(s.engine.NewSession(), pipeline.ID)
}

func (s storage) deletePipeline(sess *xorm.Session, pipelineID int64) error {
	if err := s.workflowsDelete(sess, pipelineID); err != nil {
		return err
	}

	var confIDs []int64
	if err := sess.Table(new(model.PipelineConfig)).Select("config_id").Where("pipeline_id = ?", pipelineID).Find(&confIDs); err != nil {
		return err
	}
	for _, confID := range confIDs {
		exist, err := sess.Where(builder.Eq{"config_id": confID}.And(builder.Neq{"pipeline_id": pipelineID})).Exist(new(model.PipelineConfig))
		if err != nil {
			return err
		}
		if !exist {
			// this config is only used for this pipeline. so delete it
			if _, err := sess.Where(builder.Eq{"id": confID}).Delete(new(model.Config)); err != nil {
				return err
			}
		}
	}

	if _, err := sess.Where("pipeline_id = ?", pipelineID).Delete(new(model.PipelineConfig)); err != nil {
		return err
	}
	return wrapDelete(sess.ID(pipelineID).Delete(new(model.Pipeline)))
}

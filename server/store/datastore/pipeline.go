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

	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetPipeline(id int64) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{}
	return pipeline, wrapGet(s.engine.ID(id).Get(pipeline))
}

func (s storage) GetPipelineNumber(repo *model.Repo, num int64) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Number: num,
	}
	return pipeline, wrapGet(s.engine.Get(pipeline))
}

func (s storage) GetPipelineRef(repo *model.Repo, ref string) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Ref:    ref,
	}
	return pipeline, wrapGet(s.engine.Get(pipeline))
}

func (s storage) GetPipelineCommit(repo *model.Repo, sha, branch string) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Branch: branch,
		Commit: sha,
	}
	return pipeline, wrapGet(s.engine.Get(pipeline))
}

func (s storage) GetPipelineLast(repo *model.Repo, branch string) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Branch: branch,
		Event:  model.EventPush,
	}
	return pipeline, wrapGet(s.engine.Desc("pipeline_number").Get(pipeline))
}

func (s storage) GetPipelineLastBefore(repo *model.Repo, branch string, num int64) (*model.Pipeline, error) {
	pipeline := &model.Pipeline{
		RepoID: repo.ID,
		Branch: branch,
	}
	return pipeline, wrapGet(s.engine.
		Desc("pipeline_number").
		Where("pipeline_id < ?", num).
		Get(pipeline))
}

func (s storage) GetPipelineList(repo *model.Repo, p *model.PaginationData) ([]*model.Pipeline, error) {
	pipelines := make([]*model.Pipeline, 0, p.PerPage)
	return pipelines, s.engine.Where("pipeline_repo_id = ?", repo.ID).
		Desc("pipeline_number").
		Limit(p.PerPage, p.PerPage*(p.Page-1)).
		Find(&pipelines)
}

// GetActivePipelineList get all pipelines that are pending, running or blocked
func (s storage) GetActivePipelineList(repo *model.Repo, page int) ([]*model.Pipeline, error) {
	pipelines := make([]*model.Pipeline, 0, perPage)
	query := s.engine.
		Where("pipeline_repo_id = ?", repo.ID).
		In("pipeline_status", model.StatusPending, model.StatusRunning, model.StatusBlocked).
		Desc("pipeline_number")
	if page > 0 {
		query = query.Limit(perPage, perPage*(page-1))
	}
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

	repoExist, err := sess.Where("repo_id = ?", pipeline.RepoID).Exist(&model.Repo{})
	if err != nil {
		return err
	}

	if !repoExist {
		return ErrorRepoNotExist{RepoID: pipeline.RepoID}
	}

	// calc pipeline number
	var number int64
	if _, err := sess.SQL("SELECT MAX(pipeline_number) FROM `pipelines` WHERE pipeline_repo_id = ?", pipeline.RepoID).Get(&number); err != nil {
		return err
	}
	pipeline.Number = number + 1

	pipeline.Created = time.Now().UTC().Unix()
	pipeline.Enqueued = pipeline.Created
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

func deletePipeline(sess *xorm.Session, pipelineID int64) error {
	// delete related steps
	for startSteps := 0; ; startSteps += perPage {
		stepIDs := make([]int64, 0, perPage)
		if err := sess.Limit(perPage, startSteps).Table("steps").Cols("step_id").Where("step_pipeline_id = ?", pipelineID).Find(&stepIDs); err != nil {
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
	if _, err := sess.Where("pipeline_id = ?", pipelineID).Delete(new(model.PipelineConfig)); err != nil {
		return err
	}
	_, err := sess.ID(pipelineID).Delete(new(model.Pipeline))
	return err
}

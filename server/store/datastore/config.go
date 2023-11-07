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

	"go.woodpecker-ci.org/woodpecker/server/model"
)

func (s storage) ConfigsForPipeline(pipelineID int64) ([]*model.Config, error) {
	configs := make([]*model.Config, 0, perPage)
	return configs, s.engine.
		Table("config").
		Join("LEFT", "pipeline_config", "config.config_id = pipeline_config.config_id").
		Where("pipeline_config.pipeline_id = ?", pipelineID).
		Find(&configs)
}

func (s storage) ConfigFindIdentical(repoID int64, hash string) (*model.Config, error) {
	conf := new(model.Config)
	if err := wrapGet(s.engine.Where(
		builder.Eq{"config_repo_id": repoID, "config_hash": hash},
	).Get(conf)); err != nil {
		return nil, err
	}
	return conf, nil
}

func (s storage) ConfigFindApproved(config *model.Config) (bool, error) {
	return s.engine.Table("pipelines").Select("pipelines.pipeline_id").
		Join("INNER", "pipeline_config", "pipelines.pipeline_id = pipeline_config.pipeline_id").
		Where(builder.Eq{"pipelines.pipeline_repo_id": config.RepoID, "pipeline_config.config_id": config.ID}.
			And(builder.NotIn("pipelines.pipeline_status", model.StatusBlocked, model.StatusPending))).
		Exist()
}

func (s storage) ConfigCreate(config *model.Config) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(config)
	return err
}

func (s storage) PipelineConfigCreate(config *model.PipelineConfig) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(config)
	return err
}

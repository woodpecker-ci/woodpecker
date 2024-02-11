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
	"crypto/sha256"
	"errors"
	"fmt"

	"xorm.io/builder"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func (s storage) ConfigsForPipeline(pipelineID int64) ([]*model.Config, error) {
	configs := make([]*model.Config, 0, perPage)
	return configs, s.engine.
		Table("config").
		Join("LEFT", "pipeline_config", "config.config_id = pipeline_config.config_id").
		Where("pipeline_config.pipeline_id = ?", pipelineID).
		Find(&configs)
}

func (s storage) configFindIdentical(sess *xorm.Session, repoID int64, hash, name string) (*model.Config, error) {
	conf := new(model.Config)
	if err := wrapGet(sess.Where(
		builder.Eq{"config_repo_id": repoID, "config_hash": hash, "config_name": name},
	).Get(conf)); err != nil {
		return nil, err
	}
	return conf, nil
}

func (s storage) ConfigPersist(conf *model.Config) (*model.Config, error) {
	conf.Hash = fmt.Sprintf("%x", sha256.Sum256(conf.Data))

	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return nil, err
	}

	existingConfig, err := s.configFindIdentical(sess, conf.RepoID, conf.Hash, conf.Name)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		return nil, err
	}
	if existingConfig != nil {
		return existingConfig, nil
	}

	if err := s.configCreate(sess, conf); err != nil {
		return nil, err
	}

	return conf, sess.Commit()
}

func (s storage) ConfigFindApproved(config *model.Config) (bool, error) {
	return s.engine.Table("pipelines").Select("pipelines.pipeline_id").
		Join("INNER", "pipeline_config", "pipelines.pipeline_id = pipeline_config.pipeline_id").
		Where(builder.Eq{"pipelines.pipeline_repo_id": config.RepoID, "pipeline_config.config_id": config.ID}.
			And(builder.NotIn("pipelines.pipeline_status", model.StatusBlocked, model.StatusPending))).
		Exist()
}

func (s storage) ConfigCreate(config *model.Config) error {
	return s.configCreate(s.engine.NewSession(), config)
}

func (s storage) configCreate(sess *xorm.Session, config *model.Config) error {
	// should never happen but just in case
	if config.Name == "" {
		return fmt.Errorf("insert config to store failed: 'Name' has to be set")
	}
	if config.Hash == "" {
		return fmt.Errorf("insert config to store failed: 'Hash' has to be set")
	}

	// only Insert set auto created ID back to object
	_, err := sess.Insert(config)
	return err
}

func (s storage) PipelineConfigCreate(config *model.PipelineConfig) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(config)
	return err
}

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
		Table("configs").
		Join("LEFT", "pipeline_configs", "configs.id = pipeline_configs.config_id").
		Where("pipeline_configs.pipeline_id = ?", pipelineID).
		Find(&configs)
}

func (s storage) configFindIdentical(sess *xorm.Session, repoID int64, hash, name string) (*model.Config, error) {
	conf := new(model.Config)
	if err := wrapGet(sess.Where(
		builder.Eq{"repo_id": repoID, "hash": hash, "name": name},
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

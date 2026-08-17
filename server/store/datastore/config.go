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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// ConfigsForPipeline returns what the pipeline actually ran: the source
// configs, with anything a compile workflow rewrote already applied.
func (s storage) ConfigsForPipeline(pipelineID int64) ([]*model.Config, error) {
	return s.configsForPipeline(pipelineID, builder.Eq{"pipeline_configs.effective": true})
}

// SourceConfigsForPipeline returns what the pipeline was built from, before any
// compile workflow rewrote it.
//
// What a compile workflow emits is an output artifact and is never fed back in,
// so approving or restarting a pipeline compiles again rather than replaying
// the previous run's output as input.
func (s storage) SourceConfigsForPipeline(pipelineID int64) ([]*model.Config, error) {
	return s.configsForPipeline(pipelineID, builder.Eq{"pipeline_configs.source": true})
}

func (s storage) configsForPipeline(pipelineID int64, cond builder.Cond) ([]*model.Config, error) {
	configs := make([]*model.Config, 0, perPage)
	return configs, s.engine.
		Table("configs").
		Join("LEFT", "pipeline_configs", "configs.id = pipeline_configs.config_id").
		Where(builder.Eq{"pipeline_configs.pipeline_id": pipelineID}).
		And(cond).
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
	if err != nil && !errors.Is(err, types.ErrRecordNotExist) {
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
	return wrapInsert(sess.Insert(config))
}

func (s storage) PipelineConfigCreate(config *model.PipelineConfig) error {
	// only Insert set auto created ID back to object
	return wrapInsert(s.engine.Insert(config))
}

// PipelineConfigsSetEffective records what the compile phase produced as the
// configs the pipeline actually runs.
//
// The source links are left alone: a compile workflow rewrites what runs, never
// what the pipeline was built from.
func (s storage) PipelineConfigsSetEffective(pipelineID int64, configs []*model.Config) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.
		Where(builder.Eq{"pipeline_id": pipelineID}).
		Cols("effective").
		Update(&model.PipelineConfig{Effective: false}); err != nil {
		return err
	}

	for _, config := range configs {
		link := &model.PipelineConfig{PipelineID: pipelineID, ConfigID: config.ID}

		exists, err := sess.Get(link)
		if err != nil {
			return err
		}

		if !exists {
			link.Effective = true
			if err := wrapInsert(sess.Insert(link)); err != nil {
				return err
			}
			continue
		}

		if _, err := sess.
			Where(builder.Eq{"pipeline_id": pipelineID, "config_id": config.ID}).
			Cols("effective").
			Update(&model.PipelineConfig{Effective: true}); err != nil {
			return err
		}
	}

	return sess.Commit()
}

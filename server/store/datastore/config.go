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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) ConfigsForBuild(buildID int64) ([]*model.Config, error) {
	configs := make([]*model.Config, 0, perPage)
	return configs, s.engine.
		Table("config").
		Join("LEFT", "build_config", "config.config_id = build_config.config_id").
		Where("build_config.build_id = ?", buildID).
		Find(&configs)
}

func (s storage) ConfigFindIdentical(repoID int64, hash string) (*model.Config, error) {
	conf := &model.Config{
		RepoID: repoID,
		Hash:   hash,
	}
	if err := wrapGet(s.engine.Get(conf)); err != nil {
		return nil, err
	}
	return conf, nil
}

func (s storage) ConfigFindApproved(config *model.Config) (bool, error) {
	return s.engine.Where("build_repo_id = ?", config.RepoID).
		And(builder.In("build_id", builder.Expr( // TODO: use JOIN
			`SELECT build_id
  FROM build_config
  WHERE build_config.config_id = ?`, config.ID))).
		And(builder.In("build_status", "blocked", "pending")).
		Exist(new(model.Build))
}

func (s storage) ConfigCreate(config *model.Config) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(config)
	return err
}

func (s storage) BuildConfigCreate(config *model.BuildConfig) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(config)
	return err
}

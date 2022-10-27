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
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetPipelineQueue() ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	status := []string{"pending", "running"}
	err := s.engine.Table("pipelines p").
		Select(`
	r.repo_owner,
	r.repo_name,
	r.repo_full_name,
	p.pipeline_number,
	p.pipeline_event,
	p.pipeline_status,
	p.pipeline_created,
	p.pipeline_started,
	p.pipeline_finished,
	p.pipeline_commit,
	p.pipeline_branch,
	p.pipeline_ref,
	p.pipeline_refspec,
	p.pipeline_remote,
	p.pipeline_title,
	p.pipeline_message,
	p.pipeline_author,
	p.pipeline_email,
	p.pipeline_avatar`).
		Join("INNER", "repos r", "p.pipeline_repo_id = r.repo_id").
		In("p.pipeline_status", status).
		Find(&feed)
	return feed, err
}

func (s storage) UserFeed(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	err := s.engine.Table("repos r").
		Select(`
	r.repo_owner,
	r.repo_name,
	r.repo_full_name,
	pl.pipeline_number,
	pl.pipeline_event,
	pl.pipeline_status,
	pl.pipeline_created,
	pl.pipeline_started,
	pl.pipeline_finished,
	pl.pipeline_commit,
	pl.pipeline_branch,
	pl.pipeline_ref,
	pl.pipeline_refspec,
	pl.pipeline_remote,
	pl.pipeline_title,
	pl.pipeline_message,
	pl.pipeline_author,
	pl.pipeline_email,
	pl.pipeline_avatar`).
		Join("INNER", "perms pe", "r.repo_id = pe.perm_repo_id").
		Join("INNER", "pipelines pl", "r.repo_id = pl.pipeline_repo_id").
		Where("pe.perm_user_id = ? AND (pe.perm_push = ? OR pe.perm_admin = ?)", user.ID, true, true).
		Desc("pl.pipeline_id").
		Limit(50).
		Find(&feed)

	return feed, err
}

func (s storage) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)

	err := s.engine.Table("repos r").
		Select(`
	r.repo_owner,
	r.repo_name,
	r.repo_full_name,
	pl.pipeline_number,
	pl.pipeline_event,
	pl.pipeline_status,
	pl.pipeline_created,
	pl.pipeline_started,
	pl.pipeline_finished,
	pl.pipeline_commit,
	pl.pipeline_branch,
	pl.pipeline_ref,
	pl.pipeline_refspec,
	pl.pipeline_remote,
	pl.pipeline_title,
	pl.pipeline_message,
	pl.pipeline_author,
	pl.pipeline_email,
	pl.pipeline_avatar`).
		Join("LEFT", "pipelines pl", "pl.pipeline_id = (SELECT pipeline_id FROM pipelines WHERE pipelines.pipeline_repo_id = r.repo_id ORDER BY pipeline_id DESC LIMIT 1)").
		Join("INNER", "perms pe", "r.repo_id = pe.perm_repo_id").
		Where("pe.perm_user_id = ? AND (pe.perm_push = ? OR pe.perm_admin = ?) AND r.repo_active = ?", user.ID, true, true, true).
		Asc("r.repo_full_name").
		Find(&feed)

	return feed, err
}

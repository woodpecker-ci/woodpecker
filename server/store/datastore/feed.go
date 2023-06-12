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

var feedItemSelect = `repos.repo_id as feed_repo_id,
pipelines.pipeline_id as feed_pipeline_id,
pipelines.pipeline_number as feed_pipeline_number,
pipelines.pipeline_event as feed_pipeline_event,
pipelines.pipeline_status as feed_pipeline_status,
pipelines.pipeline_created as feed_pipeline_created,
pipelines.pipeline_started as feed_pipeline_started,
pipelines.pipeline_finished as feed_pipeline_finished,
pipelines.pipeline_commit as feed_pipeline_commit,
pipelines.pipeline_branch as feed_pipeline_branch,
pipelines.pipeline_ref as feed_pipeline_ref,
pipelines.pipeline_refspec as feed_pipeline_refspec,
pipelines.pipeline_clone_url as feed_pipeline_clone_url,
pipelines.pipeline_title as feed_pipeline_title,
pipelines.pipeline_message as feed_pipeline_message,
pipelines.pipeline_author as feed_pipeline_author,
pipelines.pipeline_email as feed_pipeline_email,
pipelines.pipeline_avatar as feed_pipeline_avatar`

func (s storage) GetPipelineQueue() ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	err := s.engine.Table("pipelines").
		Select(feedItemSelect).
		Join("INNER", "repos", "pipelines.pipeline_repo_id = repos.repo_id").
		In("pipelines.pipeline_status", model.StatusPending, model.StatusRunning).
		Find(&feed)
	return feed, err
}

func (s storage) UserFeed(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	err := s.engine.Table("repos").
		Select(feedItemSelect).
		Join("INNER", "perms", "repos.repo_id = perms.perm_repo_id").
		Join("INNER", "pipelines", "repos.repo_id = pipelines.pipeline_repo_id").
		Where(userPushOrAdminCondition(user.ID)).
		Desc("pipelines.pipeline_id").
		Limit(perPage).
		Find(&feed)

	return feed, err
}

func (s storage) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)

	err := s.engine.Table("repos").
		Select(feedItemSelect).
		Join("INNER", "perms", "repos.repo_id = perms.perm_repo_id").
		Join("LEFT", "pipelines", "pipelines.pipeline_id = "+`(
			SELECT pipelines.pipeline_id FROM pipelines
			WHERE pipelines.pipeline_repo_id = repos.repo_id
			ORDER BY pipelines.pipeline_id DESC
			LIMIT 1
			)`).
		Where(userPushOrAdminCondition(user.ID)).
		And(builder.Eq{"repos.repo_active": true}).
		Asc("repos.repo_full_name").
		Find(&feed)

	return feed, err
}

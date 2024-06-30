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

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var feedItemSelect = `repos.id as repo_id,
pipelines.id as pipeline_id,
pipelines.number as pipeline_number,
pipelines.event as pipeline_event,
pipelines.status as pipeline_status,
pipelines.created as pipeline_created,
pipelines.started as pipeline_started,
pipelines.finished as pipeline_finished,
'pipelines.commit' as pipeline_commit,
pipelines.branch as pipeline_branch,
pipelines.ref as pipeline_ref,
pipelines.refspec as pipeline_refspec,
pipelines.title as pipeline_title,
pipelines.message as pipeline_message,
pipelines.author as pipeline_author,
pipelines.email as pipeline_email,
pipelines.avatar as pipeline_avatar`

func (s storage) GetPipelineQueue() ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	err := s.engine.Table("pipelines").
		Select(feedItemSelect).
		Join("INNER", "repos", "pipelines.repo_id = repos.id").
		In("pipelines.status", model.StatusPending, model.StatusRunning).
		Find(&feed)
	return feed, err
}

func (s storage) UserFeed(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	err := s.engine.Table("repos").
		Select(feedItemSelect).
		Join("INNER", "perms", "repos.id = perms.repo_id").
		Join("INNER", "pipelines", "repos.id = pipelines.repo_id").
		Where(userPushOrAdminCondition(user.ID)).
		Desc("pipelines.id").
		Limit(perPage).
		Find(&feed)

	return feed, err
}

func (s storage) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)

	err := s.engine.Table("repos").
		Select(feedItemSelect).
		Join("INNER", "perms", "repos.id = perms.repo_id").
		Join("LEFT", "pipelines", "pipelines.id = "+`(
			SELECT pipelines.id FROM pipelines
			WHERE pipelines.repo_id = repos.id
			ORDER BY pipelines.id DESC
			LIMIT 1
			)`).
		Where(userPushOrAdminCondition(user.ID)).
		And(builder.Eq{"repos.active": true}).
		Asc("repos.full_name").
		Find(&feed)

	return feed, err
}

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

package datastore_xorm

import (
	"xorm.io/builder"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetBuildQueue() ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	err := s.engine.SQL(`
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM
 builds b
,repos r
WHERE b.build_repo_id = r.repo_id
  AND b.build_status IN ('pending','running')
`).Find(&feed)
	return feed, err
}

func (s storage) UserFeed(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	return feed, s.engine.SQL(`
SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM repos
INNER JOIN perms  ON perms.perm_repo_id   = repos.repo_id
INNER JOIN builds ON builds.build_repo_id = repos.repo_id
`).Where("perms.perm_user_id = ?", user.ID).
		And(builder.Eq{"perms.perm_push": true}.Or(builder.Eq{"perms.perm_admin": true})).
		Desc("build_created").
		Limit(perPage).
		Find(&feed)
}

func (s storage) RepoListLatest(user *model.User) ([]*model.Feed, error) {
	feed := make([]*model.Feed, 0, perPage)
	return feed, s.engine.SQL(`SELECT
 repo_owner
,repo_name
,repo_full_name
,build_number
,build_event
,build_status
,build_created
,build_started
,build_finished
,build_commit
,build_branch
,build_ref
,build_refspec
,build_remote
,build_title
,build_message
,build_author
,build_email
,build_avatar
FROM repos LEFT OUTER JOIN builds ON build_id = (
	SELECT build_id FROM builds
	WHERE builds.build_repo_id = repos.repo_id
	LIMIT 1
)
INNER JOIN perms ON perms.perm_repo_id = repos.repo_id`).
		Where("perms.perm_user_id = ?", user.ID).
		And(builder.Eq{"perms.perm_push": true}.Or(builder.Eq{"perms.perm_admin": true})).
		And(builder.Eq{"repos.repo_active": true}).
		Asc("repo_full_name").
		Desc("build_created").
		Find(&feed)
}

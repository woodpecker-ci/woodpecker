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
	"github.com/rs/zerolog/log"
	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetRepo(id int64) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(s.engine.ID(id).Get(repo))
}

func (s storage) GetRepoName(fullName string) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	return s.getRepoName(sess, fullName)
}

func (s storage) getRepoName(e *xorm.Session, fullName string) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(e.Where("repo_full_name = ?", fullName).Get(repo))
}

func (s storage) GetRepoCount() (int64, error) {
	return s.engine.Count(&model.Repo{IsActive: true})
}

func (s storage) CreateRepo(repo *model.Repo) error {
	_, err := s.engine.InsertOne(repo)
	return err
}

func (s storage) UpdateRepo(repo *model.Repo) error {
	_, err := s.engine.ID(repo.ID).AllCols().Update(repo)
	return err
}

func (s storage) DeleteRepo(repo *model.Repo) error {
	_, err := s.engine.ID(repo.ID).Delete(repo)
	// TODO: delete related within a session
	return err
}

// RepoList list all repos where permissions fo specific user are stored
// TODO: paginate
func (s storage) RepoList(user *model.User, owned bool) ([]*model.Repo, error) {
	repos := make([]*model.Repo, 0, perPage)
	sess := s.engine.Table("repos").
		Join("INNER", "perms", "perms.perm_repo_id = repos.repo_id").
		Where("perms.perm_user_id = ?", user.ID)
	if owned {
		sess = sess.And(builder.Eq{"perms.perm_push": true}.Or(builder.Eq{"perms.perm_admin": true}))
	}
	return repos, sess.
		Asc("repo_full_name").
		Find(&repos)
}

func (s storage) RepoBatch(repos []*model.Repo) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	for _, repo := range repos {
		if repo.UserID == 0 || len(repo.Owner) == 0 || len(repo.Name) == 0 || len(repo.FullName) == 0 {
			log.Debug().Msgf("skip insert/update repo: %v", repo)
			continue
		}
		exist, err := sess.Exist(&repo)
		if err != nil {
			return err
		}
		if exist {
			if _, err := sess.Update(&repo); err != nil {
				return err
			}
		} else {
			if _, err := sess.InsertOne(&repo); err != nil {
				return err
			}
		}
	}

	return sess.Commit()
}

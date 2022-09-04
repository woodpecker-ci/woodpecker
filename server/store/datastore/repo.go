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
	"github.com/rs/zerolog/log"
	"xorm.io/builder"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetRepo(id int64) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(s.engine.ID(id).Get(repo))
}

func (s storage) GetRepoRemoteID(id string) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	return s.getRepoRemoteID(sess, id)
}

func (s storage) getRepoRemoteID(e *xorm.Session, id string) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(e.Where("remote_id = ?", id).Get(repo))
}

func (s storage) GetRepoNameFallback(remoteID, fullName string) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	return s.getRepoNameFallback(sess, remoteID, fullName)
}

func (s storage) getRepoNameFallback(e *xorm.Session, remoteID, fullName string) (*model.Repo, error) {
	repo, err := s.getRepoRemoteID(e, remoteID)
	if err == RecordNotExist {
		return s.getRepoName(e, fullName)
	}
	return repo, err
}

func (s storage) GetRepoName(fullName string) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	repo, err := s.getRepoName(sess, fullName)
	if err == RecordNotExist {
		// the repository does not exist, so look for a redirection
		redirect, err := s.getRedirection(sess, fullName)
		if err != nil {
			return nil, err
		}
		return s.GetRepo(redirect.RepoID)
	}
	return repo, err
}

func (s storage) getRepoName(e *xorm.Session, fullName string) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(e.Where("repo_full_name = ?", fullName).Get(repo))
}

func (s storage) GetRepoCount() (int64, error) {
	return s.engine.Where(builder.Eq{"repo_active": true}).Count(new(model.Repo))
}

func (s storage) CreateRepo(repo *model.Repo) error {
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(repo)
	return err
}

func (s storage) UpdateRepo(repo *model.Repo) error {
	_, err := s.engine.ID(repo.ID).AllCols().Update(repo)
	return err
}

func (s storage) DeleteRepo(repo *model.Repo) error {
	const batchSize = perPage
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if _, err := sess.Where("config_repo_id = ?", repo.ID).Delete(new(model.Config)); err != nil {
		return err
	}
	if _, err := sess.Where("perm_repo_id = ?", repo.ID).Delete(new(model.Perm)); err != nil {
		return err
	}
	if _, err := sess.Where("registry_repo_id = ?", repo.ID).Delete(new(model.Registry)); err != nil {
		return err
	}
	if _, err := sess.Where("secret_repo_id = ?", repo.ID).Delete(new(model.Secret)); err != nil {
		return err
	}
	if _, err := sess.Where("repo_id = ?", repo.ID).Delete(new(model.Redirection)); err != nil {
		return err
	}

	// delete related builds
	for startBuilds := 0; ; startBuilds += batchSize {
		buildIDs := make([]int64, 0, batchSize)
		if err := sess.Limit(batchSize, startBuilds).Table("builds").Cols("build_id").Where("build_repo_id = ?", repo.ID).Find(&buildIDs); err != nil {
			return err
		}
		if len(buildIDs) == 0 {
			break
		}

		for i := range buildIDs {
			if err := deleteBuild(sess, buildIDs[i]); err != nil {
				return err
			}
		}
	}

	if _, err := sess.ID(repo.ID).Delete(new(model.Repo)); err != nil {
		return err
	}

	return sess.Commit()
}

// RepoList list all repos where permissions for specific user are stored
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

// RepoBatch Sync batch of repos from SCM (with permissions) to store (create if not exist else update)
// TODO: only store activated repos ...
func (s storage) RepoBatch(repos []*model.Repo) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	for i := range repos {
		if len(repos[i].Owner) == 0 || len(repos[i].Name) == 0 || len(repos[i].FullName) == 0 {
			log.Debug().Msgf("skip insert/update repo: %#v", repos[i])
			continue
		}

		repo, err := s.getRepoNameFallback(sess, repos[i].RemoteID, repos[i].FullName)
		if err != nil && err != RecordNotExist {
			return err
		}
		exist := err == nil // if there's an error, it must be a RecordNotExist

		if exist {
			if repos[i].FullName != repo.FullName {
				// create redirection
				err := s.createRedirection(sess, &model.Redirection{RepoID: repo.ID, FullName: repo.FullName})
				if err != nil {
					return err
				}
			}
			if repos[i].RemoteID != "" && repos[i].RemoteID != "0" {
				if _, err := sess.
					Where("remote_id = ?", repos[i].RemoteID).
					Cols("repo_owner", "repo_name", "repo_full_name", "repo_scm", "repo_avatar", "repo_link", "repo_private", "repo_clone", "repo_branch", "remote_id").
					Update(repos[i]); err != nil {
					return err
				}
			} else {
				if _, err := sess.
					Where("repo_owner = ?", repos[i].Owner).
					And(" repo_name = ?", repos[i].Name).
					Cols("repo_owner", "repo_name", "repo_full_name", "repo_scm", "repo_avatar", "repo_link", "repo_private", "repo_clone", "repo_branch", "remote_id").
					Update(repos[i]); err != nil {
					return err
				}
			}

			_, err := sess.
				Where("remote_id = ?", repos[i].RemoteID).
				Get(repos[i])
			if err != nil {
				return err
			}
		} else {
			// only Insert on single object ref set auto created ID back to object
			if _, err := sess.Insert(repos[i]); err != nil {
				return err
			}
		}

		if repos[i].Perm != nil {
			repos[i].Perm.RepoID = repos[i].ID
			repos[i].Perm.Repo = repos[i]
			if err := s.permUpsert(sess, repos[i].Perm); err != nil {
				return err
			}
		}
	}

	return sess.Commit()
}

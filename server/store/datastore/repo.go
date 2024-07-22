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
	"errors"
	"fmt"
	"strings"

	"xorm.io/builder"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func (s storage) GetRepo(id int64) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(s.engine.ID(id).Get(repo))
}

func (s storage) GetRepoForgeID(remoteID model.ForgeRemoteID) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	return s.getRepoForgeID(sess, remoteID)
}

func (s storage) getRepoForgeID(e *xorm.Session, remoteID model.ForgeRemoteID) (*model.Repo, error) {
	repo := new(model.Repo)
	return repo, wrapGet(e.Where("forge_remote_id = ?", remoteID).Get(repo))
}

func (s storage) GetRepoNameFallback(remoteID model.ForgeRemoteID, fullName string) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	return s.getRepoNameFallback(sess, remoteID, fullName)
}

func (s storage) getRepoNameFallback(e *xorm.Session, remoteID model.ForgeRemoteID, fullName string) (*model.Repo, error) {
	repo, err := s.getRepoForgeID(e, remoteID)
	if errors.Is(err, types.RecordNotExist) {
		return s.getRepoName(e, fullName)
	}
	return repo, err
}

func (s storage) GetRepoName(fullName string) (*model.Repo, error) {
	sess := s.engine.NewSession()
	defer sess.Close()
	repo, err := s.getRepoName(sess, fullName)
	if errors.Is(err, types.RecordNotExist) {
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
	return repo, wrapGet(e.Where("LOWER(full_name) = ?", strings.ToLower(fullName)).Get(repo))
}

func (s storage) GetRepoCount() (int64, error) {
	return s.engine.Where(builder.Eq{"active": true}).Count(new(model.Repo))
}

func (s storage) CreateRepo(repo *model.Repo) error {
	switch {
	case repo.Name == "":
		return fmt.Errorf("repo name is empty")
	case repo.Owner == "":
		return fmt.Errorf("repo owner is empty")
	case repo.FullName == "":
		return fmt.Errorf("repo full name is empty")
	}
	// only Insert set auto created ID back to object
	_, err := s.engine.Insert(repo)
	return err
}

func (s storage) UpdateRepo(repo *model.Repo) error {
	_, err := s.engine.ID(repo.ID).AllCols().Update(repo)
	return err
}

func (s storage) DeleteRepo(repo *model.Repo) error {
	return s.deleteRepo(s.engine.NewSession(), repo)
}

func (s storage) deleteRepo(sess *xorm.Session, repo *model.Repo) error {
	const batchSize = perPage
	if _, err := sess.Where("repo_id = ?", repo.ID).Delete(new(model.Config)); err != nil {
		return err
	}
	if _, err := sess.Where("repo_id = ?", repo.ID).Delete(new(model.Perm)); err != nil {
		return err
	}
	if _, err := sess.Where("repo_id = ?", repo.ID).Delete(new(model.Registry)); err != nil {
		return err
	}
	if _, err := sess.Where("repo_id = ?", repo.ID).Delete(new(model.Secret)); err != nil {
		return err
	}
	if _, err := sess.Where("repo_id = ?", repo.ID).Delete(new(model.Redirection)); err != nil {
		return err
	}

	// delete related pipelines
	for startPipelines := 0; ; startPipelines += batchSize {
		pipelineIDs := make([]int64, 0, batchSize)
		if err := sess.Limit(batchSize, startPipelines).Table("pipelines").Cols("id").Where("repo_id = ?", repo.ID).Find(&pipelineIDs); err != nil {
			return err
		}
		if len(pipelineIDs) == 0 {
			break
		}

		for i := range pipelineIDs {
			if err := s.deletePipeline(sess, pipelineIDs[i]); err != nil {
				return err
			}
		}
	}

	return wrapDelete(sess.ID(repo.ID).Delete(new(model.Repo)))
}

// RepoList list all repos where permissions for specific user are stored
// TODO: paginate
func (s storage) RepoList(user *model.User, owned, active bool) ([]*model.Repo, error) {
	repos := make([]*model.Repo, 0)
	sess := s.engine.Table("repos").
		Join("INNER", "perms", "perms.repo_id = repos.id").
		Where("perms.user_id = ?", user.ID)
	if owned {
		sess = sess.And(builder.Eq{"perms.push": true}.Or(builder.Eq{"perms.admin": true}))
	}
	if active {
		sess = sess.And(builder.Eq{"repos.active": true})
	}
	return repos, sess.
		Asc("full_name").
		Find(&repos)
}

// RepoListAll list all repos.
func (s storage) RepoListAll(active bool, p *model.ListOptions) ([]*model.Repo, error) {
	repos := make([]*model.Repo, 0)
	sess := s.paginate(p).Table("repos")
	if active {
		sess = sess.And(builder.Eq{"repos.active": true})
	}
	return repos, sess.
		Asc("full_name").
		Find(&repos)
}

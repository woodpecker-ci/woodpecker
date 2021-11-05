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
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func (s storage) GetBuild(id int64) (*model.Build, error) {
	build := &model.Build{}
	return build, wrapGet(s.engine.ID(id).Get(build))
}

func (s storage) GetBuildNumber(repo *model.Repo, num int64) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Number: num,
	}
	return build, wrapGet(s.engine.Get(build))
}

func (s storage) GetBuildRef(repo *model.Repo, ref string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Ref:    ref,
	}
	return build, wrapGet(s.engine.Get(build))
}

func (s storage) GetBuildCommit(repo *model.Repo, sha, branch string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
		Commit: sha,
	}
	return build, wrapGet(s.engine.Get(build))
}

func (s storage) GetBuildLast(repo *model.Repo, branch string) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
		Event:  "push",
	}
	return build, wrapGet(s.engine.Desc("build_number").Get(build))
}

func (s storage) GetBuildLastBefore(repo *model.Repo, branch string, num int64) (*model.Build, error) {
	build := &model.Build{
		RepoID: repo.ID,
		Branch: branch,
	}
	return build, wrapGet(s.engine.
		Desc("build_number").
		Where("build_id < ?", num).
		Get(build))
}

func (s storage) GetBuildList(repo *model.Repo, page int) ([]*model.Build, error) {
	builds := make([]*model.Build, 0, perPage)
	return builds, s.engine.Where("build_repo_id = ?", repo.ID).
		Desc("build_number").
		Limit(perPage, perPage*(page-1)).
		Find(&builds)
}

func (s storage) GetBuildCount() (int64, error) {
	return s.engine.Count(new(model.Build))
}

func (s storage) CreateBuild(build *model.Build, procList ...*model.Proc) error {
	sess := s.engine.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	// increment counter
	if _, err := sess.ID(build.RepoID).Incr("repo_counter").Update(new(model.Repo)); err != nil {
		return err
	}

	repo := new(model.Repo)
	if err := wrapGet(sess.ID(build.RepoID).Get(repo)); err != nil {
		return err
	}

	build.Number = repo.Counter
	build.Created = time.Now().UTC().Unix()
	build.Enqueued = build.Created
	// only Insert set auto created ID back to object
	if _, err := sess.Insert(build); err != nil {
		return err
	}

	for i := range procList {
		// only Insert set auto created ID back to object
		if _, err := sess.Insert(procList[i]); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func (s storage) UpdateBuild(build *model.Build) error {
	_, err := s.engine.ID(build.ID).AllCols().Update(build)
	return err
}

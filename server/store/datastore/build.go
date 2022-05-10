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

	"xorm.io/xorm"

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

func (s storage) GetActiveBuildList(repo *model.Repo, page int) ([]*model.Build, error) {
	builds := make([]*model.Build, 0, perPage)
	query := s.engine.
		Where("build_repo_id = ?", repo.ID).
		Where("build_status = ? or build_status = ?", model.StatusPending, model.StatusRunning, model.StatusBlocked).
		Desc("build_number")
	if page > 0 {
		query = query.Limit(perPage, perPage*(page-1))
	}
	return builds, query.Find(&builds)
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

	// calc build number
	var number int64
	if _, err := sess.SQL("SELECT MAX(build_number) FROM `builds` WHERE build_repo_id = ?", build.RepoID).Get(&number); err != nil {
		return err
	}
	build.Number = number + 1

	build.Created = time.Now().UTC().Unix()
	build.Enqueued = build.Created
	// only Insert set auto created ID back to object
	if _, err := sess.Insert(build); err != nil {
		return err
	}

	for i := range procList {
		procList[i].BuildID = build.ID
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

func deleteBuild(sess *xorm.Session, buildID int64) error {
	// delete related procs
	for startProcs := 0; ; startProcs += perPage {
		procIDs := make([]int64, 0, perPage)
		if err := sess.Limit(perPage, startProcs).Table("procs").Cols("proc_id").Where("proc_build_id = ?", buildID).Find(&procIDs); err != nil {
			return err
		}
		if len(procIDs) == 0 {
			break
		}

		for i := range procIDs {
			if err := deleteProc(sess, procIDs[i]); err != nil {
				return err
			}
		}
	}
	if _, err := sess.Where("build_id = ?", buildID).Delete(new(model.BuildConfig)); err != nil {
		return err
	}
	_, err := sess.ID(buildID).Delete(new(model.Build))
	return err
}

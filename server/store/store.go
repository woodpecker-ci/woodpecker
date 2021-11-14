// Copyright 2018 Drone.IO Inc.
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

package store

import (
	"io"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO: CreateX func should return new object to not indirect let storage change an existing object (alter ID etc...)

type Store interface {
	// GetUser gets a user by unique ID.
	GetUser(int64) (*model.User, error)

	// GetUserLogin gets a user by unique Login name.
	GetUserLogin(string) (*model.User, error)

	// GetUserList gets a list of all users in the system.
	// TODO: paginate
	GetUserList() ([]*model.User, error)

	// GetUserCount gets a count of all users in the system.
	GetUserCount() (int64, error)

	// CreateUser creates a new user account.
	CreateUser(*model.User) error

	// UpdateUser updates a user account.
	UpdateUser(*model.User) error

	// DeleteUser deletes a user account.
	DeleteUser(*model.User) error

	// GetRepo gets a repo by unique ID.
	GetRepo(int64) (*model.Repo, error)

	// GetRepoName gets a repo by its full name.
	GetRepoName(string) (*model.Repo, error)

	// GetRepoCount gets a count of all repositories in the system.
	GetRepoCount() (int64, error)

	// CreateRepo creates a new repository.
	CreateRepo(*model.Repo) error

	// UpdateRepo updates a user repository.
	UpdateRepo(*model.Repo) error

	// DeleteRepo deletes a user repository.
	DeleteRepo(*model.Repo) error

	// GetBuild gets a build by unique ID.
	GetBuild(int64) (*model.Build, error)

	// GetBuildNumber gets a build by number.
	GetBuildNumber(*model.Repo, int64) (*model.Build, error)

	// GetBuildRef gets a build by its ref.
	GetBuildRef(*model.Repo, string) (*model.Build, error)

	// GetBuildCommit gets a build by its commit sha.
	GetBuildCommit(*model.Repo, string, string) (*model.Build, error)

	// GetBuildLast gets the last build for the branch.
	GetBuildLast(*model.Repo, string) (*model.Build, error)

	// GetBuildLastBefore gets the last build before build number N.
	GetBuildLastBefore(*model.Repo, string, int64) (*model.Build, error)

	// GetBuildList gets a list of builds for the repository
	// TODO: paginate
	GetBuildList(*model.Repo, int) ([]*model.Build, error)

	// GetBuildQueue gets a list of build in queue.
	GetBuildQueue() ([]*model.Feed, error)

	// GetBuildCount gets a count of all builds in the system.
	GetBuildCount() (int64, error)

	// CreateBuild creates a new build and jobs.
	CreateBuild(*model.Build, ...*model.Proc) error

	// UpdateBuild updates a build.
	UpdateBuild(*model.Build) error

	//
	// new functions
	//

	UserFeed(*model.User) ([]*model.Feed, error)

	// RepoList TODO: paginate
	RepoList(user *model.User, owned bool) ([]*model.Repo, error)
	RepoListLatest(*model.User) ([]*model.Feed, error)
	// RepoBatch Sync batch of repos (with permissions) to store (create if not exist else update)
	RepoBatch([]*model.Repo) error

	PermFind(user *model.User, repo *model.Repo) (*model.Perm, error)
	PermUpsert(perm *model.Perm) error
	PermDelete(perm *model.Perm) error
	PermFlush(user *model.User, before int64) error

	ConfigsForBuild(buildID int64) ([]*model.Config, error)
	ConfigFindIdentical(repoID int64, hash string) (*model.Config, error)
	ConfigFindApproved(*model.Config) (bool, error)
	ConfigCreate(*model.Config) error
	BuildConfigCreate(*model.BuildConfig) error

	SenderFind(*model.Repo, string) (*model.Sender, error)
	// SenderList TODO: paginate
	SenderList(*model.Repo) ([]*model.Sender, error)
	SenderCreate(*model.Sender) error
	SenderUpdate(*model.Sender) error
	SenderDelete(*model.Sender) error

	SecretFind(*model.Repo, string) (*model.Secret, error)
	SecretList(*model.Repo) ([]*model.Secret, error)
	SecretCreate(*model.Secret) error
	SecretUpdate(*model.Secret) error
	SecretDelete(*model.Secret) error

	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo) ([]*model.Registry, error)
	RegistryCreate(*model.Registry) error
	RegistryUpdate(*model.Registry) error
	RegistryDelete(*model.Registry) error

	ProcLoad(int64) (*model.Proc, error)
	ProcFind(*model.Build, int) (*model.Proc, error)
	ProcChild(*model.Build, int, string) (*model.Proc, error)
	ProcList(*model.Build) ([]*model.Proc, error)
	ProcCreate([]*model.Proc) error
	ProcUpdate(*model.Proc) error
	ProcClear(*model.Build) error

	LogFind(*model.Proc) (io.ReadCloser, error)
	// TODO: since we do ReadAll in any case a ioReader is not the best idear
	// so either find a way to write log in chunks by xorm ...
	LogSave(*model.Proc, io.Reader) error

	FileList(*model.Build) ([]*model.File, error)
	FileFind(*model.Proc, string) (*model.File, error)
	FileRead(*model.Proc, string) (io.ReadCloser, error)
	FileCreate(*model.File, io.Reader) error

	// TaskList TODO: paginate & opt filter
	TaskList() ([]*model.Task, error)
	TaskInsert(*model.Task) error
	TaskDelete(string) error

	Ping() error
	Close() error
	Migrate() error
}

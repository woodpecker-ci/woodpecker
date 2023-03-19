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

//go:generate go install github.com/vektra/mockery/v2@latest
//go:generate mockery --name Store --output mocks --case underscore

import (
	"io"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO: CreateX func should return new object to not indirect let storage change an existing object (alter ID etc...)

type Store interface {
	// Users
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

	// Repos
	// GetRepo gets a repo by unique ID.
	GetRepo(int64) (*model.Repo, error)
	// GetRepoForgeID gets a repo by its forge ID.
	GetRepoForgeID(model.ForgeRemoteID) (*model.Repo, error)
	// GetRepoNameFallback gets the repo by its forge ID and if this doesn't exist by its full name.
	GetRepoNameFallback(remoteID model.ForgeRemoteID, fullName string) (*model.Repo, error)
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

	// Redirections
	// GetRedirection returns the redirection for the given full repo name
	GetRedirection(string) (*model.Redirection, error)
	// CreateRedirection creates a redirection
	CreateRedirection(redirection *model.Redirection) error
	// HasRedirectionForRepo checks if there's a redirection for the given repo and full name
	HasRedirectionForRepo(int64, string) (bool, error)

	// Pipelines
	// GetPipeline gets a pipeline by unique ID.
	GetPipeline(int64) (*model.Pipeline, error)
	// GetPipelineNumber gets a pipeline by number.
	GetPipelineNumber(*model.Repo, int64) (*model.Pipeline, error)
	// GetPipelineRef gets a pipeline by its ref.
	GetPipelineRef(*model.Repo, string) (*model.Pipeline, error)
	// GetPipelineCommit gets a pipeline by its commit sha.
	GetPipelineCommit(*model.Repo, string, string) (*model.Pipeline, error)
	// GetPipelineLast gets the last pipeline for the branch.
	GetPipelineLast(*model.Repo, string) (*model.Pipeline, error)
	// GetPipelineLastBefore gets the last pipeline before pipeline number N.
	GetPipelineLastBefore(*model.Repo, string, int64) (*model.Pipeline, error)
	// GetPipelineList gets a list of pipelines for the repository
	// TODO: paginate
	GetPipelineList(*model.Repo, int) ([]*model.Pipeline, error)
	// GetPipelineList gets a list of the active pipelines for the repository
	GetActivePipelineList(repo *model.Repo, page int) ([]*model.Pipeline, error)
	// GetPipelineQueue gets a list of pipelines in queue.
	GetPipelineQueue() ([]*model.Feed, error)
	// GetPipelineCount gets a count of all pipelines in the system.
	GetPipelineCount() (int64, error)
	// CreatePipeline creates a new pipeline and steps.
	CreatePipeline(*model.Pipeline, ...*model.Step) error
	// UpdatePipeline updates a pipeline.
	UpdatePipeline(*model.Pipeline) error

	// Feeds
	UserFeed(*model.User) ([]*model.Feed, error)

	// Repositories
	// RepoList TODO: paginate
	RepoList(user *model.User, owned bool) ([]*model.Repo, error)
	RepoListLatest(*model.User) ([]*model.Feed, error)
	// RepoBatch Sync batch of repos from SCM (with permissions) to store (create if not exist else update)
	RepoBatch([]*model.Repo) error

	// Permissions
	PermFind(user *model.User, repo *model.Repo) (*model.Perm, error)
	PermUpsert(perm *model.Perm) error
	PermDelete(perm *model.Perm) error
	PermFlush(user *model.User, before int64) error

	// Configs
	ConfigsForPipeline(pipelineID int64) ([]*model.Config, error)
	ConfigFindIdentical(repoID int64, hash string) (*model.Config, error)
	ConfigFindApproved(*model.Config) (bool, error)
	ConfigCreate(*model.Config) error
	PipelineConfigCreate(*model.PipelineConfig) error

	// Secrets
	// GetSecret return secret by ID, make sure the function that calls this performs permission checks!
	GetSecret(secretID int64) (*model.Secret, error)
	SecretFind(*model.Repo, string) (*model.Secret, error)
	SecretList(*model.Repo, bool) ([]*model.Secret, error)
	SecretListAll() ([]*model.Secret, error)
	SecretCreate(*model.Secret) error
	SecretUpdate(*model.Secret) error
	SecretDelete(*model.Secret) error
	OrgSecretFind(string, string) (*model.Secret, error)
	OrgSecretList(string) ([]*model.Secret, error)
	GlobalSecretFind(string) (*model.Secret, error)
	GlobalSecretList() ([]*model.Secret, error)

	// Registries
	RegistryFind(*model.Repo, string) (*model.Registry, error)
	RegistryList(*model.Repo) ([]*model.Registry, error)
	RegistryCreate(*model.Registry) error
	RegistryUpdate(*model.Registry) error
	RegistryDelete(repo *model.Repo, addr string) error

	// Steps
	StepLoad(int64) (*model.Step, error)
	StepFind(*model.Pipeline, int) (*model.Step, error)
	StepChild(*model.Pipeline, int, string) (*model.Step, error)
	StepList(*model.Pipeline) ([]*model.Step, error)
	StepCreate([]*model.Step) error
	StepUpdate(*model.Step) error
	StepClear(*model.Pipeline) error

	// Logs
	LogFind(*model.Step) (io.ReadCloser, error)
	// TODO: since we do ReadAll in any case a ioReader is not the best idea
	// so either find a way to write log in chunks by xorm ...
	LogSave(*model.Step, io.Reader) error

	// Files
	FileList(*model.Pipeline) ([]*model.File, error)
	FileFind(*model.Step, string) (*model.File, error)
	FileRead(*model.Step, string) (io.ReadCloser, error)
	FileCreate(*model.File, io.Reader) error

	// Tasks
	// TaskList TODO: paginate & opt filter
	TaskList() ([]*model.Task, error)
	TaskInsert(*model.Task) error
	TaskDelete(string) error

	// ServerConfig
	ServerConfigGet(string) (string, error)
	ServerConfigSet(string, string) error
	ServerConfigDelete(string) error

	// Cron
	CronCreate(*model.Cron) error
	CronFind(*model.Repo, int64) (*model.Cron, error)
	CronList(*model.Repo) ([]*model.Cron, error)
	CronUpdate(*model.Repo, *model.Cron) error
	CronDelete(*model.Repo, int64) error
	CronListNextExecute(int64, int64) ([]*model.Cron, error)
	CronGetLock(*model.Cron, int64) (bool, error)

	// Agent
	AgentCreate(*model.Agent) error
	AgentFind(int64) (*model.Agent, error)
	AgentFindByToken(string) (*model.Agent, error)
	AgentList() ([]*model.Agent, error)
	AgentUpdate(*model.Agent) error
	AgentDelete(*model.Agent) error

	// Store operations
	Ping() error
	Close() error
	Migrate() error
}

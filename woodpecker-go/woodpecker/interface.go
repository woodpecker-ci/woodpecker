// Copyright 2022 Woodpecker Authors
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

package woodpecker

import (
	"net/http"
)

// Client is used to communicate with a Woodpecker server.
type Client interface {
	// SetClient sets the http.Client.
	SetClient(*http.Client)

	// SetAddress sets the server address.
	SetAddress(string)

	// Self returns the currently authenticated user.
	Self() (*User, error)

	// User returns a user by login.
	User(string) (*User, error)

	// UserList returns a list of all registered users.
	UserList() ([]*User, error)

	// UserPost creates a new user account.
	UserPost(*User) (*User, error)

	// UserPatch updates a user account.
	UserPatch(*User) (*User, error)

	// UserDel deletes a user account.
	UserDel(string) error

	// Repo returns a repository by name.
	Repo(repoID int64) (*Repo, error)

	// RepoLookup returns a repository id by the owner and name.
	RepoLookup(repoFullName string) (int64, error)

	// RepoList returns a list of all repositories to which the user has explicit
	// access in the host system.
	RepoList() ([]*Repo, error)

	// RepoListOpts returns a list of all repositories to which the user has
	// explicit access in the host system.
	RepoListOpts(bool, bool) ([]*Repo, error)

	// RepoPost activates a repository.
	RepoPost(forgeRemoteID int64) (*Repo, error)

	// RepoPatch updates a repository.
	RepoPatch(repoID int64, repo *RepoPatch) (*Repo, error)

	// RepoMove moves the repository
	RepoMove(repoID int64, dst string) error

	// RepoChown updates a repository owner.
	RepoChown(repoID int64) (*Repo, error)

	// RepoRepair repairs the repository hooks.
	RepoRepair(repoID int64) error

	// RepoDel deletes a repository.
	RepoDel(repoID int64) error

	// Pipeline returns a repository pipeline by number.
	Pipeline(repoID int64, pipeline int) (*Pipeline, error)

	// PipelineLast returns the latest repository pipeline by branch. An empty branch
	// will result in the default branch.
	PipelineLast(repoID int64, branch string) (*Pipeline, error)

	// PipelineList returns a list of recent pipelines for the
	// the specified repository.
	PipelineList(repoID int64) ([]*Pipeline, error)

	// PipelineQueue returns a list of enqueued pipelines.
	PipelineQueue() ([]*Activity, error)

	// PipelineCreate returns creates a pipeline on specified branch.
	PipelineCreate(repoID int64, opts *PipelineOptions) (*Pipeline, error)

	// PipelineStart re-starts a stopped pipeline.
	PipelineStart(repoID int64, num int, params map[string]string) (*Pipeline, error)

	// PipelineStop stops the given pipeline.
	PipelineStop(repoID int64, pipeline int) error

	// PipelineApprove approves a blocked pipeline.
	PipelineApprove(repoID int64, pipeline int) (*Pipeline, error)

	// PipelineDecline declines a blocked pipeline.
	PipelineDecline(repoID int64, pipeline int) (*Pipeline, error)

	// PipelineKill force kills the running pipeline.
	PipelineKill(repoID int64, pipeline int) error

	// PipelineLogs returns the logs for the given pipeline
	PipelineLogs(repoID int64, pipeline, step int) ([]*Logs, error)

	// Deploy triggers a deployment for an existing pipeline using the specified
	// target environment.
	Deploy(repoID int64, pipeline int, env string, params map[string]string) (*Pipeline, error)

	// LogsPurge purges the pipeline logs for the specified pipeline.
	LogsPurge(repoID int64, pipeline int) error

	// Registry returns a registry by hostname.
	Registry(repoID int64, hostname string) (*Registry, error)

	// RegistryList returns a list of all repository registries.
	RegistryList(repoID int64) ([]*Registry, error)

	// RegistryCreate creates a registry.
	RegistryCreate(repoID int64, registry *Registry) (*Registry, error)

	// RegistryUpdate updates a registry.
	RegistryUpdate(repoID int64, registry *Registry) (*Registry, error)

	// RegistryDelete deletes a registry.
	RegistryDelete(repoID int64, hostname string) error

	// Secret returns a secret by name.
	Secret(repoID int64, secret string) (*Secret, error)

	// SecretList returns a list of all repository secrets.
	SecretList(repoID int64) ([]*Secret, error)

	// SecretCreate creates a secret.
	SecretCreate(repoID int64, secret *Secret) (*Secret, error)

	// SecretUpdate updates a secret.
	SecretUpdate(repoID int64, secret *Secret) (*Secret, error)

	// SecretDelete deletes a secret.
	SecretDelete(repoID int64, secret string) error

	// OrgSecret returns an organization secret by name.
	OrgSecret(owner, secret string) (*Secret, error)

	// OrgSecretList returns a list of all organization secrets.
	OrgSecretList(owner string) ([]*Secret, error)

	// OrgSecretCreate creates an organization secret.
	OrgSecretCreate(owner string, secret *Secret) (*Secret, error)

	// OrgSecretUpdate updates an organization secret.
	OrgSecretUpdate(owner string, secret *Secret) (*Secret, error)

	// OrgSecretDelete deletes an organization secret.
	OrgSecretDelete(owner, secret string) error

	// GlobalSecret returns an global secret by name.
	GlobalSecret(secret string) (*Secret, error)

	// GlobalSecretList returns a list of all global secrets.
	GlobalSecretList() ([]*Secret, error)

	// GlobalSecretCreate creates a global secret.
	GlobalSecretCreate(secret *Secret) (*Secret, error)

	// GlobalSecretUpdate updates a global secret.
	GlobalSecretUpdate(secret *Secret) (*Secret, error)

	// GlobalSecretDelete deletes a global secret.
	GlobalSecretDelete(secret string) error

	// QueueInfo returns the queue state.
	QueueInfo() (*Info, error)

	// LogLevel returns the current logging level
	LogLevel() (*LogLevel, error)

	// SetLogLevel sets the server's logging level
	SetLogLevel(logLevel *LogLevel) (*LogLevel, error)

	// CronList list all cron jobs of a repo
	CronList(repoID int64) ([]*Cron, error)

	// CronGet get a specific cron job of a repo by id
	CronGet(repoID, cronID int64) (*Cron, error)

	// CronDelete delete a specific cron job of a repo by id
	CronDelete(repoID, cronID int64) error

	// CronCreate create a new cron job in a repo
	CronCreate(repoID int64, cron *Cron) (*Cron, error)

	// CronUpdate update an existing cron job of a repo
	CronUpdate(repoID int64, cron *Cron) (*Cron, error)

	// AgentList returns a list of all registered agents
	AgentList() ([]*Agent, error)

	// Agent returns an agent by id
	Agent(int64) (*Agent, error)

	// AgentCreate creates a new agent
	AgentCreate(*Agent) (*Agent, error)

	// AgentUpdate updates an existing agent
	AgentUpdate(*Agent) (*Agent, error)

	// AgentDelete deletes an agent
	AgentDelete(int64) error

	// AgentTasksList returns a list of all tasks executed by an agent
	AgentTasksList(int64) ([]*Task, error)
}

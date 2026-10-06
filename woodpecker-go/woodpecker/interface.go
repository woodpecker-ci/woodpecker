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
	"context"
	"net/http"
)

// Client is used to communicate with a Woodpecker server.
type Client interface {
	// SetClient sets the http.Client.
	SetClient(*http.Client)

	// SetAddress sets the server address.
	SetAddress(string)

	// Self returns the currently authenticated user.
	Self(ctx context.Context) (*User, error)

	// User returns a user by login.
	// It is recommended to specify forgeID (default is 1).
	User(ctx context.Context, login string, forgeID ...int64) (*User, error)

	// UserList returns a list of all registered users.
	UserList(ctx context.Context, opt UserListOptions) ([]*User, error)

	// UserPost creates a new user account.
	UserPost(ctx context.Context, user *User) (*User, error)

	// UserPatch updates a user account.
	UserPatch(ctx context.Context, user *User) (*User, error)

	// UserDel deletes a user account.
	// It is recommended to specify forgeID (default is 1).
	UserDel(ctx context.Context, login string, forgeID ...int64) error

	// Repo returns a repository by name.
	Repo(ctx context.Context, repoID int64) (*Repo, error)

	// RepoLookup returns a repository id by the owner and name.
	RepoLookup(ctx context.Context, repoFullName string) (*Repo, error)

	// RepoList returns a list of all repositories to which the user has explicit
	// access in the host system.
	RepoList(ctx context.Context, opt RepoListOptions) ([]*Repo, error)

	// RepoPost activates a repository.
	RepoPost(ctx context.Context, opt RepoPostOptions) (*Repo, error)

	// RepoPatch updates a repository.
	RepoPatch(ctx context.Context, repoID int64, repo *RepoPatch) (*Repo, error)

	// RepoMove moves the repository
	RepoMove(ctx context.Context, repoID int64, opt RepoMoveOptions) error

	// RepoChown updates a repository owner.
	RepoChown(ctx context.Context, repoID int64) (*Repo, error)

	// RepoRepair repairs the repository hooks.
	RepoRepair(ctx context.Context, repoID int64) error

	// RepoDel deletes a repository.
	RepoDel(ctx context.Context, repoID int64) error

	// Pipeline returns a repository pipeline by number.
	Pipeline(ctx context.Context, repoID, pipeline int64) (*Pipeline, error)

	// PipelineLast returns the latest repository pipeline.
	PipelineLast(ctx context.Context, repoID int64, opt PipelineLastOptions) (*Pipeline, error)

	// PipelineList returns a list of recent pipelines for the
	// the specified repository.
	PipelineList(ctx context.Context, repoID int64, opt PipelineListOptions) ([]*Pipeline, error)

	PipelineDelete(ctx context.Context, repoID, pipeline int64) error

	// PipelineQueue returns a list of enqueued pipelines.
	PipelineQueue(ctx context.Context) ([]*Feed, error)

	// PipelineCreate returns creates a pipeline on specified branch.
	PipelineCreate(ctx context.Context, repoID int64, opts *PipelineOptions) (*Pipeline, error)

	// PipelineStart re-starts a stopped pipeline.
	PipelineStart(ctx context.Context, repoID, num int64, opt PipelineStartOptions) (*Pipeline, error)

	// PipelineStop stops the given pipeline.
	PipelineStop(ctx context.Context, repoID, pipeline int64) error

	// PipelineApprove approves a blocked pipeline.
	PipelineApprove(ctx context.Context, repoID, pipeline int64) (*Pipeline, error)

	// PipelineDecline declines a blocked pipeline.
	PipelineDecline(ctx context.Context, repoID, pipeline int64) (*Pipeline, error)

	// PipelineMetadata returns metadata for a pipeline.
	PipelineMetadata(ctx context.Context, repoID int64, pipelineNumber int) ([]byte, error)

	// StepLogEntries returns the LogEntries for the given pipeline step
	StepLogEntries(ctx context.Context, repoID, pipeline, stepID int64) ([]*LogEntry, error)

	// Deploy triggers a deployment for an existing pipeline using the specified
	// target environment.
	Deploy(ctx context.Context, repoID, pipeline int64, opt DeployOptions) (*Pipeline, error)

	// LogsPurge purges the pipeline logs for the specified pipeline.
	LogsPurge(ctx context.Context, repoID, pipeline int64) error

	// StepLogsPurge purges the pipeline logs for the specified step.
	StepLogsPurge(ctx context.Context, repoID, pipelineNumber, stepID int64) error

	// Registry returns a registry by hostname.
	Registry(ctx context.Context, repoID int64, hostname string) (*Registry, error)

	// RegistryList returns a list of all repository registries.
	RegistryList(ctx context.Context, repoID int64, opt RegistryListOptions) ([]*Registry, error)

	// RegistryCreate creates a registry.
	RegistryCreate(ctx context.Context, repoID int64, registry *Registry) (*Registry, error)

	// RegistryUpdate updates a registry.
	RegistryUpdate(ctx context.Context, repoID int64, registry *Registry) (*Registry, error)

	// RegistryDelete deletes a registry.
	RegistryDelete(ctx context.Context, repoID int64, hostname string) error

	// OrgRegistry returns an organization registry by address.
	OrgRegistry(ctx context.Context, orgID int64, registry string) (*Registry, error)

	// OrgRegistryList returns a list of all organization registries.
	OrgRegistryList(ctx context.Context, orgID int64, opt RegistryListOptions) ([]*Registry, error)

	// OrgRegistryCreate creates an organization registry.
	OrgRegistryCreate(ctx context.Context, orgID int64, registry *Registry) (*Registry, error)

	// OrgRegistryUpdate updates an organization registry.
	OrgRegistryUpdate(ctx context.Context, orgID int64, registry *Registry) (*Registry, error)

	// OrgRegistryDelete deletes an organization registry.
	OrgRegistryDelete(ctx context.Context, orgID int64, registry string) error

	// GlobalRegistry returns an global registry by address.
	GlobalRegistry(ctx context.Context, registry string) (*Registry, error)

	// GlobalRegistryList returns a list of all global registries.
	GlobalRegistryList(ctx context.Context, opt RegistryListOptions) ([]*Registry, error)

	// GlobalRegistryCreate creates a global registry.
	GlobalRegistryCreate(ctx context.Context, registry *Registry) (*Registry, error)

	// GlobalRegistryUpdate updates a global registry.
	GlobalRegistryUpdate(ctx context.Context, registry *Registry) (*Registry, error)

	// GlobalRegistryDelete deletes a global registry.
	GlobalRegistryDelete(ctx context.Context, registry string) error

	// Secret returns a secret by name.
	Secret(ctx context.Context, repoID int64, secret string) (*Secret, error)

	// SecretList returns a list of all repository secrets.
	SecretList(ctx context.Context, repoID int64, opt SecretListOptions) ([]*Secret, error)

	// SecretCreate creates a secret.
	SecretCreate(ctx context.Context, repoID int64, secret *Secret) (*Secret, error)

	// SecretUpdate updates a secret.
	SecretUpdate(ctx context.Context, repoID int64, secret *Secret) (*Secret, error)

	// SecretDelete deletes a secret.
	SecretDelete(ctx context.Context, repoID int64, secret string) error

	// Org returns an organization by name.
	Org(ctx context.Context, orgID int64) (*Org, error)

	// OrgLookup returns an organization id by name.
	OrgLookup(ctx context.Context, orgName string) (*Org, error)

	// OrgList returns a list of all organizations.
	OrgList(ctx context.Context, opt ListOptions) ([]*Org, error)

	// OrgSecret returns an organization secret by name.
	OrgSecret(ctx context.Context, orgID int64, secret string) (*Secret, error)

	// OrgSecretList returns a list of all organization secrets.
	OrgSecretList(ctx context.Context, orgID int64, opt SecretListOptions) ([]*Secret, error)

	// OrgSecretCreate creates an organization secret.
	OrgSecretCreate(ctx context.Context, orgID int64, secret *Secret) (*Secret, error)

	// OrgSecretUpdate updates an organization secret.
	OrgSecretUpdate(ctx context.Context, orgID int64, secret *Secret) (*Secret, error)

	// OrgSecretDelete deletes an organization secret.
	OrgSecretDelete(ctx context.Context, orgID int64, secret string) error

	// GlobalSecret returns an global secret by name.
	GlobalSecret(ctx context.Context, secret string) (*Secret, error)

	// GlobalSecretList returns a list of all global secrets.
	GlobalSecretList(ctx context.Context, opt SecretListOptions) ([]*Secret, error)

	// GlobalSecretCreate creates a global secret.
	GlobalSecretCreate(ctx context.Context, secret *Secret) (*Secret, error)

	// GlobalSecretUpdate updates a global secret.
	GlobalSecretUpdate(ctx context.Context, secret *Secret) (*Secret, error)

	// GlobalSecretDelete deletes a global secret.
	GlobalSecretDelete(ctx context.Context, secret string) error

	// QueueInfo returns the queue state.
	QueueInfo(ctx context.Context) (*Info, error)

	// LogLevel returns the current logging level.
	LogLevel(ctx context.Context) (*LogLevel, error)

	// SetLogLevel sets the server's logging level.
	SetLogLevel(ctx context.Context, logLevel *LogLevel) (*LogLevel, error)

	// CronList list all cron jobs of a repo.
	CronList(ctx context.Context, repoID int64, opt CronListOptions) ([]*Cron, error)

	// CronGet get a specific cron job of a repo by id.
	CronGet(ctx context.Context, repoID, cronID int64) (*Cron, error)

	// CronDelete delete a specific cron job of a repo by id.
	CronDelete(ctx context.Context, repoID, cronID int64) error

	// CronCreate create a new cron job in a repo.
	CronCreate(ctx context.Context, repoID int64, cron *Cron) (*Cron, error)

	// CronUpdate update an existing cron job of a repo.
	CronUpdate(ctx context.Context, repoID int64, cron *Cron) (*Cron, error)

	// AgentList returns a page of registered agents. The server never
	// returns more than one page, so a caller that needs every agent has to
	// walk the pages until one comes back short.
	AgentList(ctx context.Context, opt AgentListOptions) ([]*Agent, error)

	// Agent returns an agent by id.
	Agent(ctx context.Context, agentID int64) (*Agent, error)

	// AgentCreate creates a new agent.
	AgentCreate(ctx context.Context, agent *Agent) (*Agent, error)

	// AgentUpdate updates an existing agent.
	AgentUpdate(ctx context.Context, agent *Agent) (*Agent, error)

	// AgentDelete deletes an agent.
	AgentDelete(ctx context.Context, agentID int64) error

	// AgentTasksList returns a list of all tasks executed by an agent.
	AgentTasksList(ctx context.Context, agentID int64) ([]*Task, error)
}

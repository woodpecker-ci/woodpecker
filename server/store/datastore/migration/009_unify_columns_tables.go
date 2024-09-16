// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"fmt"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type configV009 struct {
	ID     int64  `xorm:"pk autoincr 'config_id'"`
	RepoID int64  `xorm:"UNIQUE(s) 'config_repo_id'"`
	Hash   string `xorm:"UNIQUE(s) 'config_hash'"`
	Name   string `xorm:"UNIQUE(s) 'config_name'"`
	Data   []byte `xorm:"LONGBLOB 'config_data'"`
}

func (configV009) TableName() string {
	return "config"
}

type cronV009 struct {
	ID        int64  `xorm:"pk autoincr 'i_d'"`
	Name      string `xorm:"name UNIQUE(s) INDEX"`
	RepoID    int64  `xorm:"repo_id UNIQUE(s) INDEX"`
	CreatorID int64  `xorm:"creator_id INDEX"`
	NextExec  int64  `xorm:"next_exec"`
	Schedule  string `xorm:"schedule NOT NULL"`
	Created   int64  `xorm:"created NOT NULL DEFAULT 0"`
	Branch    string `xorm:"branch"`
}

func (cronV009) TableName() string {
	return "crons"
}

type permV009 struct {
	UserID int64 `xorm:"UNIQUE(s) INDEX NOT NULL 'perm_user_id'"`
	RepoID int64 `xorm:"UNIQUE(s) INDEX NOT NULL 'perm_repo_id'"`
	Pull   bool  `xorm:"perm_pull"`
	Push   bool  `xorm:"perm_push"`
	Admin  bool  `xorm:"perm_admin"`
	Synced int64 `xorm:"perm_synced"`
}

func (permV009) TableName() string {
	return "perms"
}

type pipelineV009 struct {
	ID         int64                  `xorm:"pk autoincr 'pipeline_id'"`
	RepoID     int64                  `xorm:"UNIQUE(s) INDEX 'pipeline_repo_id'"`
	Number     int64                  `xorm:"UNIQUE(s) 'pipeline_number'"`
	Author     string                 `xorm:"INDEX 'pipeline_author'"`
	Parent     int64                  `xorm:"pipeline_parent"`
	Event      model.WebhookEvent     `xorm:"pipeline_event"`
	Status     model.StatusValue      `xorm:"INDEX 'pipeline_status'"`
	Errors     []*types.PipelineError `xorm:"json 'pipeline_errors'"`
	Created    int64                  `xorm:"pipeline_created"`
	Started    int64                  `xorm:"pipeline_started"`
	Finished   int64                  `xorm:"pipeline_finished"`
	Deploy     string                 `xorm:"pipeline_deploy"`
	DeployTask string                 `xorm:"pipeline_deploy_task"`
	Commit     string                 `xorm:"pipeline_commit"`
	Branch     string                 `xorm:"pipeline_branch"`
	Ref        string                 `xorm:"pipeline_ref"`
	Refspec    string                 `xorm:"pipeline_refspec"`
	Title      string                 `xorm:"pipeline_title"`
	Message    string                 `xorm:"TEXT 'pipeline_message'"`
	Timestamp  int64                  `xorm:"pipeline_timestamp"`
	Sender     string                 `xorm:"pipeline_sender"` // uses reported user for webhooks and name of cron for cron pipelines
	Avatar     string                 `xorm:"pipeline_avatar"`
	Email      string                 `xorm:"pipeline_email"`
	ForgeURL   string                 `xorm:"pipeline_forge_url"`
	Reviewer   string                 `xorm:"pipeline_reviewer"`
	Reviewed   int64                  `xorm:"pipeline_reviewed"`
}

func (pipelineV009) TableName() string {
	return "pipelines"
}

type redirectionV009 struct {
	ID int64 `xorm:"pk autoincr 'redirection_id'"`
}

func (r redirectionV009) TableName() string {
	return "redirections"
}

type registryV009 struct {
	ID       int64  `xorm:"pk autoincr 'registry_id'"`
	RepoID   int64  `xorm:"UNIQUE(s) INDEX 'registry_repo_id'"`
	Address  string `xorm:"UNIQUE(s) INDEX 'registry_addr'"`
	Username string `xorm:"varchar(2000) 'registry_username'"`
	Password string `xorm:"TEXT 'registry_password'"`
}

func (registryV009) TableName() string {
	return "registry"
}

type repoV009 struct {
	ID           int64                `xorm:"pk autoincr 'repo_id'"`
	UserID       int64                `xorm:"repo_user_id"`
	OrgID        int64                `xorm:"repo_org_id"`
	Owner        string               `xorm:"UNIQUE(name) 'repo_owner'"`
	Name         string               `xorm:"UNIQUE(name) 'repo_name'"`
	FullName     string               `xorm:"UNIQUE 'repo_full_name'"`
	Avatar       string               `xorm:"varchar(500) 'repo_avatar'"`
	ForgeURL     string               `xorm:"varchar(1000) 'repo_forge_url'"`
	Clone        string               `xorm:"varchar(1000) 'repo_clone'"`
	CloneSSH     string               `xorm:"varchar(1000) 'repo_clone_ssh'"`
	Branch       string               `xorm:"varchar(500) 'repo_branch'"`
	SCMKind      model.SCMKind        `xorm:"varchar(50) 'repo_scm'"`
	PREnabled    bool                 `xorm:"DEFAULT TRUE 'repo_pr_enabled'"`
	Timeout      int64                `xorm:"repo_timeout"`
	Visibility   model.RepoVisibility `xorm:"varchar(10) 'repo_visibility'"`
	IsSCMPrivate bool                 `xorm:"repo_private"`
	IsTrusted    bool                 `xorm:"repo_trusted"`
	IsGated      bool                 `xorm:"repo_gated"`
	IsActive     bool                 `xorm:"repo_active"`
	AllowPull    bool                 `xorm:"repo_allow_pr"`
	AllowDeploy  bool                 `xorm:"repo_allow_deploy"`
	Config       string               `xorm:"varchar(500) 'repo_config_path'"`
	Hash         string               `xorm:"varchar(500) 'repo_hash'"`
}

func (repoV009) TableName() string {
	return "repos"
}

type secretV009 struct {
	ID     int64                `xorm:"pk autoincr 'secret_id'"`
	OrgID  int64                `xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_org_id'"`
	RepoID int64                `xorm:"NOT NULL DEFAULT 0 UNIQUE(s) INDEX 'secret_repo_id'"`
	Name   string               `xorm:"NOT NULL UNIQUE(s) INDEX 'secret_name'"`
	Value  string               `xorm:"TEXT 'secret_value'"`
	Images []string             `xorm:"json 'secret_images'"`
	Events []model.WebhookEvent `xorm:"json 'secret_events'"`
}

func (secretV009) TableName() string {
	return "secrets"
}

type stepV009 struct {
	ID         int64             `xorm:"pk autoincr 'step_id'"`
	UUID       string            `xorm:"INDEX 'step_uuid'"`
	PipelineID int64             `xorm:"UNIQUE(s) INDEX 'step_pipeline_id'"`
	PID        int               `xorm:"UNIQUE(s) 'step_pid'"`
	PPID       int               `xorm:"step_ppid"`
	Name       string            `xorm:"step_name"`
	State      model.StatusValue `xorm:"step_state"`
	Error      string            `xorm:"TEXT 'step_error'"`
	Failure    string            `xorm:"step_failure"`
	ExitCode   int               `xorm:"step_exit_code"`
	Started    int64             `xorm:"step_started"`
	Stopped    int64             `xorm:"step_stopped"`
	Type       model.StepType    `xorm:"step_type"`
}

func (stepV009) TableName() string {
	return "steps"
}

type taskV009 struct {
	ID           string                       `xorm:"PK UNIQUE 'task_id'"`
	Data         []byte                       `xorm:"LONGBLOB 'task_data'"`
	Labels       map[string]string            `xorm:"json 'task_labels'"`
	Dependencies []string                     `xorm:"json 'task_dependencies'"`
	RunOn        []string                     `xorm:"json 'task_run_on'"`
	DepStatus    map[string]model.StatusValue `xorm:"json 'task_dep_status'"`
}

func (taskV009) TableName() string {
	return "tasks"
}

type userV009 struct {
	ID     int64  `xorm:"pk autoincr 'user_id'"`
	Login  string `xorm:"UNIQUE 'user_login'"`
	Token  string `xorm:"TEXT 'user_token'"`
	Secret string `xorm:"TEXT 'user_secret'"`
	Expiry int64  `xorm:"user_expiry"`
	Email  string `xorm:" varchar(500) 'user_email'"`
	Avatar string `xorm:" varchar(500) 'user_avatar'"`
	Admin  bool   `xorm:"user_admin"`
	Hash   string `xorm:"UNIQUE varchar(500) 'user_hash'"`
	OrgID  int64  `xorm:"user_org_id"`
}

func (userV009) TableName() string {
	return "users"
}

type workflowV009 struct {
	ID         int64             `xorm:"pk autoincr 'workflow_id'"`
	PipelineID int64             `xorm:"UNIQUE(s) INDEX 'workflow_pipeline_id'"`
	PID        int               `xorm:"UNIQUE(s) 'workflow_pid'"`
	Name       string            `xorm:"workflow_name"`
	State      model.StatusValue `xorm:"workflow_state"`
	Error      string            `xorm:"TEXT 'workflow_error'"`
	Started    int64             `xorm:"workflow_started"`
	Stopped    int64             `xorm:"workflow_stopped"`
	AgentID    int64             `xorm:"workflow_agent_id"`
	Platform   string            `xorm:"workflow_platform"`
	Environ    map[string]string `xorm:"json 'workflow_environ'"`
	AxisID     int               `xorm:"workflow_axis_id"`
}

func (workflowV009) TableName() string {
	return "workflows"
}

type serverConfigV009 struct {
	Key   string `xorm:"pk 'key'"`
	Value string `xorm:"value"`
}

func (serverConfigV009) TableName() string {
	return "server_config"
}

var unifyColumnsTables = xormigrate.Migration{
	ID: "unify-columns-tables",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := sess.Sync(new(configV009), new(cronV009), new(permV009), new(pipelineV009), new(redirectionV009), new(registryV009), new(repoV009), new(secretV009), new(stepV009), new(taskV009), new(userV009), new(workflowV009), new(serverConfigV009)); err != nil {
			return fmt.Errorf("sync models failed: %w", err)
		}

		// Config
		if err := renameColumn(sess, "config", "config_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "config", "config_repo_id", "repo_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "config", "config_hash", "hash"); err != nil {
			return err
		}
		if err := renameColumn(sess, "config", "config_name", "name"); err != nil {
			return err
		}
		if err := renameColumn(sess, "config", "config_data", "data"); err != nil {
			return err
		}
		if err := renameTable(sess, "config", "configs"); err != nil {
			return err
		}

		// PipelineConfig
		if err := renameTable(sess, "pipeline_config", "pipeline_configs"); err != nil {
			return err
		}

		// Cron
		if err := renameColumn(sess, "crons", "i_d", "id"); err != nil {
			return err
		}

		// Forge
		if err := renameTable(sess, "forge", "forges"); err != nil {
			return err
		}

		// Perm
		if err := renameColumn(sess, "perms", "perm_user_id", "user_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "perms", "perm_repo_id", "repo_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "perms", "perm_pull", "pull"); err != nil {
			return err
		}
		if err := renameColumn(sess, "perms", "perm_push", "push"); err != nil {
			return err
		}
		if err := renameColumn(sess, "perms", "perm_admin", "admin"); err != nil {
			return err
		}
		if err := renameColumn(sess, "perms", "perm_synced", "synced"); err != nil {
			return err
		}

		// Pipeline
		if err := renameColumn(sess, "pipelines", "pipeline_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_repo_id", "repo_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_number", "number"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_author", "author"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_parent", "parent"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_event", "event"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_status", "status"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_errors", "errors"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_created", "created"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_started", "started"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_finished", "finished"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_deploy", "deploy"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_deploy_task", "deploy_task"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_commit", "commit"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_branch", "branch"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_ref", "ref"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_refspec", "refspec"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_title", "title"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_message", "message"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_timestamp", "timestamp"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_sender", "sender"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_avatar", "avatar"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_email", "email"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_forge_url", "forge_url"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_reviewer", "reviewer"); err != nil {
			return err
		}
		if err := renameColumn(sess, "pipelines", "pipeline_reviewed", "reviewed"); err != nil {
			return err
		}

		// Redirection
		if err := renameColumn(sess, "redirections", "redirection_id", "id"); err != nil {
			return err
		}

		// Registry
		if err := renameColumn(sess, "registry", "registry_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "registry", "registry_repo_id", "repo_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "registry", "registry_addr", "address"); err != nil {
			return err
		}
		if err := renameColumn(sess, "registry", "registry_username", "username"); err != nil {
			return err
		}
		if err := renameColumn(sess, "registry", "registry_password", "password"); err != nil {
			return err
		}
		if err := renameTable(sess, "registry", "registries"); err != nil {
			return err
		}

		// Repo
		if err := renameColumn(sess, "repos", "repo_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_user_id", "user_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_org_id", "org_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_owner", "owner"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_name", "name"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_full_name", "full_name"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_avatar", "avatar"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_forge_url", "forge_url"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_clone", "clone"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_clone_ssh", "clone_ssh"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_branch", "branch"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_scm", "scm"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_pr_enabled", "pr_enabled"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_timeout", "timeout"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_visibility", "visibility"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_private", "private"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_trusted", "trusted"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_gated", "gated"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_active", "active"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_allow_pr", "allow_pr"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_allow_deploy", "allow_deploy"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_config_path", "config_path"); err != nil {
			return err
		}
		if err := renameColumn(sess, "repos", "repo_hash", "hash"); err != nil {
			return err
		}

		// Secrets
		if err := renameColumn(sess, "secrets", "secret_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "secrets", "secret_org_id", "org_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "secrets", "secret_repo_id", "repo_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "secrets", "secret_name", "name"); err != nil {
			return err
		}
		if err := renameColumn(sess, "secrets", "secret_value", "value"); err != nil {
			return err
		}
		if err := renameColumn(sess, "secrets", "secret_images", "images"); err != nil {
			return err
		}
		if err := renameColumn(sess, "secrets", "secret_events", "events"); err != nil {
			return err
		}

		// ServerConfig
		if err := renameTable(sess, "server_config", "server_configs"); err != nil {
			return err
		}

		// Step
		if err := renameColumn(sess, "steps", "step_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_uuid", "uuid"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_pipeline_id", "pipeline_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_pid", "pid"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_ppid", "ppid"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_name", "name"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_state", "state"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_error", "error"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_failure", "failure"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_exit_code", "exit_code"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_started", "started"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_stopped", "stopped"); err != nil {
			return err
		}
		if err := renameColumn(sess, "steps", "step_type", "type"); err != nil {
			return err
		}

		// Task
		if err := renameColumn(sess, "tasks", "task_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "tasks", "task_data", "data"); err != nil {
			return err
		}
		if err := renameColumn(sess, "tasks", "task_labels", "labels"); err != nil {
			return err
		}
		if err := renameColumn(sess, "tasks", "task_dependencies", "dependencies"); err != nil {
			return err
		}
		if err := renameColumn(sess, "tasks", "task_run_on", "run_on"); err != nil {
			return err
		}
		if err := renameColumn(sess, "tasks", "task_dep_status", "dependencies_status"); err != nil {
			return err
		}

		// User
		if err := renameColumn(sess, "users", "user_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_login", "login"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_token", "token"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_secret", "secret"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_expiry", "expiry"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_email", "email"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_avatar", "avatar"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_admin", "admin"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_hash", "hash"); err != nil {
			return err
		}
		if err := renameColumn(sess, "users", "user_org_id", "org_id"); err != nil {
			return err
		}

		// Workflow
		if err := renameColumn(sess, "workflows", "workflow_id", "id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_pipeline_id", "pipeline_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_pid", "pid"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_name", "name"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_state", "state"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_error", "error"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_started", "started"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_stopped", "stopped"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_agent_id", "agent_id"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_platform", "platform"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_environ", "environ"); err != nil {
			return err
		}
		if err := renameColumn(sess, "workflows", "workflow_axis_id", "axis_id"); err != nil {
			return err
		}

		return nil
	},
}

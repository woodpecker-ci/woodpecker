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

package exec

import (
	"time"

	"github.com/urfave/cli/v3"
)

var flags = []cli.Flag{
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_LOCAL"),
		Name:    "local",
		Usage:   "run from local directory",
		Value:   true,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_REPO_PATH"),
		Name:    "repo-path",
		Usage:   "path to local repository",
	},
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_TIMEOUT"),
		Name:    "timeout",
		Usage:   "pipeline timeout",
		Value:   time.Hour,
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_VOLUMES"),
		Name:    "volumes",
		Usage:   "pipeline volumes",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_NETWORKS"),
		Name:    "network",
		Usage:   "external networks",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_PREFIX"),
		Name:    "prefix",
		Value:   "woodpecker",
		Usage:   "prefix used for containers, volumes, networks, ... created by woodpecker",
		Hidden:  true,
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_PLUGINS_PRIVILEGED"),
		Name:    "plugins-privileged",
		Usage:   "Allow plugins to run in privileged mode, if environment variable is defined but empty there will be none",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND"),
		Name:    "backend-engine",
		Usage:   "backend engine to run pipelines on",
		Value:   "auto-detect",
	},

	//
	// backend options for pipeline compiler
	//
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NO_PROXY", "NO_PROXY", "no_proxy"),
		Usage:   "if set, pass the environment variable down as \"NO_PROXY\" to steps",
		Name:    "backend-no-proxy",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_HTTP_PROXY", "HTTP_PROXY", "http_proxy"),
		Usage:   "if set, pass the environment variable down as \"HTTP_PROXY\" to steps",
		Name:    "backend-http-proxy",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_HTTPS_PROXY", "HTTPS_PROXY", "https_proxy"),
		Usage:   "if set, pass the environment variable down as \"HTTPS_PROXY\" to steps",
		Name:    "backend-https-proxy",
	},

	//
	// Please note the below flags should match the flags from
	// pipeline/frontend/metadata.go and should be kept synchronized.
	//

	//
	// workspace default
	//
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_WORKSPACE_BASE"),
		Name:    "workspace-base",
		Value:   "/woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_WORKSPACE_PATH"),
		Name:    "workspace-path",
		Value:   "src",
	},
	//
	// netrc parameters
	//
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_NETRC_USERNAME"),
		Name:    "netrc-username",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_NETRC_PASSWORD"),
		Name:    "netrc-password",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_NETRC_MACHINE"),
		Name:    "netrc-machine",
	},
	//
	// metadata parameters
	//
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_SYSTEM_PLATFORM"),
		Name:    "system-platform",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_SYSTEM_NAME"),
		Name:    "system-name",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_SYSTEM_URL"),
		Name:    "system-url",
		Value:   "https://github.com/woodpecker-ci/woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO"),
		Name:    "repo",
		Usage:   "full repo name",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_REMOTE_ID"),
		Name:    "repo-remote-id",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_URL"),
		Name:    "repo-url",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_CLONE_URL"),
		Name:    "repo-clone-url",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_CLONE_SSH_URL"),
		Name:    "repo-clone-ssh-url",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_PRIVATE"),
		Name:    "repo-private",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("CI_REPO_TRUSTED"),
		Name:    "repo-trusted",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_NUMBER"),
		Name:    "pipeline-number",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_PARENT"),
		Name:    "pipeline-parent",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_CREATED"),
		Name:    "pipeline-created",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_STARTED"),
		Name:    "pipeline-started",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_FINISHED"),
		Name:    "pipeline-finished",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_STATUS"),
		Name:    "pipeline-status",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_EVENT"),
		Name:    "pipeline-event",
		Value:   "manual",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_URL"),
		Name:    "pipeline-url",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_DEPLOY_TARGET", "CI_PIPELINE_TARGET"), // TODO: remove CI_PIPELINE_TARGET in 3.x
		Name:    "pipeline-deploy-to",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_DEPLOY_TASK", "CI_PIPELINE_TASK"), // TODO: remove CI_PIPELINE_TASK in 3.x
		Name:    "pipeline-deploy-task",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_SHA"),
		Name:    "commit-sha",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_REF"),
		Name:    "commit-ref",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_REFSPEC"),
		Name:    "commit-refspec",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_BRANCH"),
		Name:    "commit-branch",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_MESSAGE"),
		Name:    "commit-message",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_AUTHOR_NAME"),
		Name:    "commit-author-name",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_AUTHOR_AVATAR"),
		Name:    "commit-author-avatar",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_AUTHOR_EMAIL"),
		Name:    "commit-author-email",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_NUMBER"),
		Name:    "prev-pipeline-number",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_CREATED"),
		Name:    "prev-pipeline-created",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_STARTED"),
		Name:    "prev-pipeline-started",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_FINISHED"),
		Name:    "prev-pipeline-finished",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_STATUS"),
		Name:    "prev-pipeline-status",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_EVENT"),
		Name:    "prev-pipeline-event",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_URL"),
		Name:    "prev-pipeline-url",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_SHA"),
		Name:    "prev-commit-sha",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_REF"),
		Name:    "prev-commit-ref",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_REFSPEC"),
		Name:    "prev-commit-refspec",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_BRANCH"),
		Name:    "prev-commit-branch",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_MESSAGE"),
		Name:    "prev-commit-message",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_AUTHOR_NAME"),
		Name:    "prev-commit-author-name",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_AUTHOR_AVATAR"),
		Name:    "prev-commit-author-avatar",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_AUTHOR_EMAIL"),
		Name:    "prev-commit-author-email",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_WORKFLOW_NAME"),
		Name:    "workflow-name",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_WORKFLOW_NUMBER"),
		Name:    "workflow-number",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_STEP_NAME"),
		Name:    "step-name",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("CI_ENV"),
		Name:    "env",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_FORGE_TYPE"),
		Name:    "forge-type",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_FORGE_URL"),
		Name:    "forge-url",
	},
}

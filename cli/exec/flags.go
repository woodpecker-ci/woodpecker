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
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_METADATA_FILE"),
		Name:    "metadata-file",
		Usage:   "path to pipeline metadata file (normally downloaded from UI). Parameters can be adjusted by applying additional cli flags",
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
		Usage:   "Set the metadata environment variable \"CI_SYSTEM_PLATFORM\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_SYSTEM_HOST"),
		Name:    "system-host",
		Usage:   "Set the metadata environment variable \"CI_SYSTEM_HOST\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_SYSTEM_NAME"),
		Name:    "system-name",
		Usage:   "Set the metadata environment variable \"CI_SYSTEM_NAME\".",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_SYSTEM_URL"),
		Name:    "system-url",
		Usage:   "Set the metadata environment variable \"CI_SYSTEM_URL\".",
		Value:   "https://github.com/woodpecker-ci/woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO"),
		Name:    "repo",
		Usage:   "Set the full name to derive metadata environment variables \"CI_REPO\", \"CI_REPO_NAME\" and \"CI_REPO_OWNER\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_REMOTE_ID"),
		Name:    "repo-remote-id",
		Usage:   "Set the metadata environment variable \"CI_REPO_REMOTE_ID\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_URL"),
		Name:    "repo-url",
		Usage:   "Set the metadata environment variable \"CI_REPO_URL\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_SCM"),
		Name:    "repo-scm",
		Usage:   "Set the metadata environment variable \"CI_REPO_SCM\".",
		Value:   "git",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_DEFAULT_BRANCH"),
		Name:    "repo-default-branch",
		Usage:   "Set the metadata environment variable \"CI_REPO_DEFAULT_BRANCH\".",
		Value:   "main",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_CLONE_URL"),
		Name:    "repo-clone-url",
		Usage:   "Set the metadata environment variable \"CI_REPO_CLONE_URL\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_CLONE_SSH_URL"),
		Name:    "repo-clone-ssh-url",
		Usage:   "Set the metadata environment variable \"CI_REPO_CLONE_SSH_URL\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_REPO_PRIVATE"),
		Name:    "repo-private",
		Usage:   "Set the metadata environment variable \"CI_REPO_PRIVATE\".",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("CI_REPO_TRUSTED"),
		Name:    "repo-trusted",
		Usage:   "Set the metadata environment variable \"CI_REPO_TRUSTED\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_NUMBER"),
		Name:    "pipeline-number",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_NUMBER\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_PARENT"),
		Name:    "pipeline-parent",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_PARENT\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_CREATED"),
		Name:    "pipeline-created",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_CREATED\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PIPELINE_STARTED"),
		Name:    "pipeline-started",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_STARTED\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_EVENT"),
		Name:    "pipeline-event",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_EVENT\".",
		Value:   "manual",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_FORGE_URL"),
		Name:    "pipeline-url",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_FORGE_URL\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_DEPLOY_TARGET"),
		Name:    "pipeline-deploy-to",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_DEPLOY_TARGET\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_DEPLOY_TASK"),
		Name:    "pipeline-deploy-task",
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_DEPLOY_TASK\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PIPELINE_FILES"),
		Usage:   "Set the metadata environment variable \"CI_PIPELINE_FILES\", either json formatted list of strings, or comma separated string list.",
		Name:    "pipeline-changed-files",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_SHA"),
		Name:    "commit-sha",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_SHA\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_REF"),
		Name:    "commit-ref",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_REF\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_REFSPEC"),
		Name:    "commit-refspec",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_REFSPEC\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_BRANCH"),
		Name:    "commit-branch",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_BRANCH\".",
		Value:   "main",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_MESSAGE"),
		Name:    "commit-message",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_MESSAGE\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_AUTHOR"),
		Name:    "commit-author-name",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_AUTHOR\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_AUTHOR_AVATAR"),
		Name:    "commit-author-avatar",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_AUTHOR_AVATAR\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_COMMIT_AUTHOR_EMAIL"),
		Name:    "commit-author-email",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_AUTHOR_EMAIL\".",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("CI_COMMIT_PULL_REQUEST_LABELS"),
		Name:    "commit-pull-labels",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_PULL_REQUEST_LABELS\".",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("CI_COMMIT_PRERELEASE"),
		Name:    "commit-release-is-pre",
		Usage:   "Set the metadata environment variable \"CI_COMMIT_PRERELEASE\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_NUMBER"),
		Name:    "prev-pipeline-number",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_NUMBER\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_CREATED"),
		Name:    "prev-pipeline-created",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_CREATED\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_STARTED"),
		Name:    "prev-pipeline-started",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_STARTED\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_FINISHED"),
		Name:    "prev-pipeline-finished",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_FINISHED\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_STATUS"),
		Name:    "prev-pipeline-status",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_STATUS\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_EVENT"),
		Name:    "prev-pipeline-event",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_EVENT\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_FORGE_URL"),
		Name:    "prev-pipeline-url",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_FORGE_URL\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_DEPLOY_TARGET"),
		Name:    "prev-pipeline-deploy-to",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_DEPLOY_TARGET\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_PIPELINE_DEPLOY_TASK"),
		Name:    "prev-pipeline-deploy-task",
		Usage:   "Set the metadata environment variable \"CI_PREV_PIPELINE_DEPLOY_TASK\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_SHA"),
		Name:    "prev-commit-sha",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_SHA\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_REF"),
		Name:    "prev-commit-ref",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_REF\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_REFSPEC"),
		Name:    "prev-commit-refspec",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_REFSPEC\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_BRANCH"),
		Name:    "prev-commit-branch",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_BRANCH\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_MESSAGE"),
		Name:    "prev-commit-message",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_MESSAGE\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_AUTHOR"),
		Name:    "prev-commit-author-name",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_AUTHOR\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_AUTHOR_AVATAR"),
		Name:    "prev-commit-author-avatar",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_AUTHOR_AVATAR\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_PREV_COMMIT_AUTHOR_EMAIL"),
		Name:    "prev-commit-author-email",
		Usage:   "Set the metadata environment variable \"CI_PREV_COMMIT_AUTHOR_EMAIL\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_WORKFLOW_NAME"),
		Name:    "workflow-name",
		Usage:   "Set the metadata environment variable \"CI_WORKFLOW_NAME\".",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("CI_WORKFLOW_NUMBER"),
		Name:    "workflow-number",
		Usage:   "Set the metadata environment variable \"CI_WORKFLOW_NUMBER\".",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("CI_ENV"),
		Name:    "env",
		Usage:   "Set the metadata environment variable \"CI_ENV\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_FORGE_TYPE"),
		Name:    "forge-type",
		Usage:   "Set the metadata environment variable \"CI_FORGE_TYPE\".",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("CI_FORGE_URL"),
		Name:    "forge-url",
		Usage:   "Set the metadata environment variable \"CI_FORGE_URL\".",
	},
}

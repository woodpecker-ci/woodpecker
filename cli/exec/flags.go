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

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

var flags = []cli.Flag{
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_LOCAL"},
		Name:    "local",
		Usage:   "run from local directory",
		Value:   true,
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_TIMEOUT"},
		Name:    "timeout",
		Usage:   "pipeline timeout",
		Value:   time.Hour,
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_VOLUMES"},
		Name:    "volumes",
		Usage:   "pipeline volumes",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_NETWORKS"},
		Name:    "network",
		Usage:   "external networks",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_PREFIX"},
		Name:    "prefix",
		Value:   "woodpecker",
		Usage:   "prefix used for containers, volumes, networks, ... created by woodpecker",
		Hidden:  true,
	},
	&cli.StringSliceFlag{
		Name:  "privileged",
		Usage: "privileged plugins",
		Value: cli.NewStringSlice(constant.PrivilegedPlugins...),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND"},
		Name:    "backend-engine",
		Usage:   "backend engine to run pipelines on",
		Value:   "auto-detect",
	},

	//
	// Please note the below flags should match the flags from
	// pipeline/frontend/metadata.go and should be kept synchronized.
	//

	//
	// workspace default
	//
	&cli.StringFlag{
		EnvVars: []string{"CI_WORKSPACE_BASE"},
		Name:    "workspace-base",
		Value:   "/woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_WORKSPACE_PATH"},
		Name:    "workspace-path",
		Value:   "src",
	},
	//
	// netrc parameters
	//
	&cli.StringFlag{
		EnvVars: []string{"CI_NETRC_USERNAME"},
		Name:    "netrc-username",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_NETRC_PASSWORD"},
		Name:    "netrc-password",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_NETRC_MACHINE"},
		Name:    "netrc-machine",
	},
	//
	// metadata parameters
	//
	&cli.StringFlag{
		EnvVars: []string{"CI_SYSTEM_PLATFORM"},
		Name:    "system-platform",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_SYSTEM_NAME"},
		Name:    "system-name",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_SYSTEM_LINK"},
		Name:    "system-link",
		Value:   "https://github.com/woodpecker-ci/woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_REPO_NAME"},
		Name:    "repo-name",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_REPO_LINK"},
		Name:    "repo-link",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_REPO_REMOTE"},
		Name:    "repo-remote-url",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_REPO_PRIVATE"},
		Name:    "repo-private",
	},
	&cli.IntFlag{
		EnvVars: []string{"CI_PIPELINE_NUMBER"},
		Name:    "pipeline-number",
	},
	&cli.IntFlag{
		EnvVars: []string{"CI_PIPELINE_PARENT"},
		Name:    "pipeline-parent",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PIPELINE_CREATED"},
		Name:    "pipeline-created",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PIPELINE_STARTED"},
		Name:    "pipeline-started",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PIPELINE_FINISHED"},
		Name:    "pipeline-finished",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PIPELINE_STATUS"},
		Name:    "pipeline-status",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PIPELINE_EVENT"},
		Name:    "pipeline-event",
		Value:   "manual",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PIPELINE_LINK"},
		Name:    "pipeline-link",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PIPELINE_TARGET"},
		Name:    "pipeline-target",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_SHA"},
		Name:    "commit-sha",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_REF"},
		Name:    "commit-ref",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_REFSPEC"},
		Name:    "commit-refspec",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_BRANCH"},
		Name:    "commit-branch",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_MESSAGE"},
		Name:    "commit-message",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_AUTHOR_NAME"},
		Name:    "commit-author-name",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_AUTHOR_AVATAR"},
		Name:    "commit-author-avatar",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_COMMIT_AUTHOR_EMAIL"},
		Name:    "commit-author-email",
	},
	&cli.IntFlag{
		EnvVars: []string{"CI_PREV_PIPELINE_NUMBER"},
		Name:    "prev-pipeline-number",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PREV_PIPELINE_CREATED"},
		Name:    "prev-pipeline-created",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PREV_PIPELINE_STARTED"},
		Name:    "prev-pipeline-started",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PREV_PIPELINE_FINISHED"},
		Name:    "prev-pipeline-finished",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_PIPELINE_STATUS"},
		Name:    "prev-pipeline-status",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_PIPELINE_EVENT"},
		Name:    "prev-pipeline-event",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_PIPELINE_LINK"},
		Name:    "prev-pipeline-link",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_SHA"},
		Name:    "prev-commit-sha",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_REF"},
		Name:    "prev-commit-ref",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_REFSPEC"},
		Name:    "prev-commit-refspec",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_BRANCH"},
		Name:    "prev-commit-branch",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_MESSAGE"},
		Name:    "prev-commit-message",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_AUTHOR_NAME"},
		Name:    "prev-commit-author-name",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_AUTHOR_AVATAR"},
		Name:    "prev-commit-author-avatar",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_COMMIT_AUTHOR_EMAIL"},
		Name:    "prev-commit-author-email",
	},
	&cli.IntFlag{
		EnvVars: []string{"CI_STEP_NUMBER"},
		Name:    "step-number",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"CI_ENV"},
		Name:    "env",
	},

	// TODO: add flags of backends

	// backend k8s
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_NAMESPACE"},
		Name:    "backend-k8s-namespace",
		Usage:   "backend k8s namespace",
		Value:   "woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_VOLUME_SIZE"},
		Name:    "backend-k8s-volume-size",
		Usage:   "backend k8s volume size (default 10G)",
		Value:   "10G",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_STORAGE_CLASS"},
		Name:    "backend-k8s-storage-class",
		Usage:   "backend k8s storage class",
		Value:   "",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_K8S_STORAGE_RWX"},
		Name:    "backend-k8s-storage-rwx",
		Usage:   "backend k8s storage access mode, should ReadWriteMany (RWX) instead of ReadWriteOnce (RWO) be used? (default: true)",
		Value:   true,
	},
}

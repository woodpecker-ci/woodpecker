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
		Usage:   "build from local directory",
		Value:   true,
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_TIMEOUT"},
		Name:    "timeout",
		Usage:   "build timeout",
		Value:   time.Hour,
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_VOLUMES"},
		Name:    "volumes",
		Usage:   "build volumes",
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
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DOCKER_NETWORK"},
		Name:    "backend-docker-network",
		Usage:   "attach pipeline containers (steps) to an existing docker network",
		Value:   "",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_DOCKER_IPV6"},
		Name:    "backend-docker-ipv6",
		Usage:   "enable ipv6 for pipeline containers (steps)",
		Value:   false,
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
		EnvVars: []string{"CI_SYSTEM_ARCH"},
		Name:    "system-arch",
		Value:   "linux/amd64",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_SYSTEM_NAME"},
		Name:    "system-name",
		Value:   "pipec",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_SYSTEM_LINK"},
		Name:    "system-link",
		Value:   "https://github.com/cncd/pipec",
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
		EnvVars: []string{"CI_BUILD_NUMBER"},
		Name:    "build-number",
	},
	&cli.IntFlag{
		EnvVars: []string{"CI_PARENT_BUILD_NUMBER"},
		Name:    "parent-build-number",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_BUILD_CREATED"},
		Name:    "build-created",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_BUILD_STARTED"},
		Name:    "build-started",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_BUILD_FINISHED"},
		Name:    "build-finished",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_BUILD_STATUS"},
		Name:    "build-status",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_BUILD_EVENT"},
		Name:    "build-event",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_BUILD_LINK"},
		Name:    "build-link",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_BUILD_TARGET"},
		Name:    "build-target",
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
		EnvVars: []string{"CI_PREV_BUILD_NUMBER"},
		Name:    "prev-build-number",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PREV_BUILD_CREATED"},
		Name:    "prev-build-created",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PREV_BUILD_STARTED"},
		Name:    "prev-build-started",
	},
	&cli.Int64Flag{
		EnvVars: []string{"CI_PREV_BUILD_FINISHED"},
		Name:    "prev-build-finished",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_BUILD_STATUS"},
		Name:    "prev-build-status",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_BUILD_EVENT"},
		Name:    "prev-build-event",
	},
	&cli.StringFlag{
		EnvVars: []string{"CI_PREV_BUILD_LINK"},
		Name:    "prev-build-link",
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
		EnvVars: []string{"CI_JOB_NUMBER"},
		Name:    "job-number",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"CI_ENV"},
		Name:    "env",
	},
}

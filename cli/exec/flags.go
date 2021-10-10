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

	"github.com/urfave/cli"
)

var flags = []cli.Flag{
	cli.BoolTFlag{
		EnvVar: "WOODPECKER_LOCAL",
		Name:   "local",
		Usage:  "build from local directory",
	},
	cli.DurationFlag{
		EnvVar: "WOODPECKER_TIMEOUT",
		Name:   "timeout",
		Usage:  "build timeout",
		Value:  time.Hour,
	},
	cli.StringSliceFlag{
		EnvVar: "WOODPECKER_VOLUMES",
		Name:   "volumes",
		Usage:  "build volumes",
	},
	cli.StringSliceFlag{
		EnvVar: "WOODPECKER_NETWORKS",
		Name:   "network",
		Usage:  "external networks",
	},
	cli.StringFlag{
		EnvVar: "WOODPECKER_DOCKER_PREFIX",
		Name:   "prefix",
		Value:  "woodpecker",
		Usage:  "prefix containers created by woodpecker",
		Hidden: true,
	},
	cli.StringSliceFlag{
		Name:  "privileged",
		Usage: "privileged plugins",
		Value: &cli.StringSlice{
			"plugins/docker",
			"plugins/gcr",
			"plugins/ecr",
		},
	},

	//
	// Please note the below flags should match the flags from
	// pipeline/frontend/metadata.go and should be kept synchronized.
	//

	//
	// workspace default
	//
	cli.StringFlag{
		EnvVar: "WOODPECKER_WORKSPACE_BASE",
		Name:   "workspace-base",
		Value:  "/woodpecker",
	},
	cli.StringFlag{
		EnvVar: "WOODPECKER_WORKSPACE_PATH",
		Name:   "workspace-path",
		Value:  "src",
	},
	//
	// netrc parameters
	//
	cli.StringFlag{
		EnvVar: "WOODPECKER_NETRC_USERNAME",
		Name:   "netrc-username",
	},
	cli.StringFlag{
		EnvVar: "WOODPECKER_NETRC_PASSWORD",
		Name:   "netrc-password",
	},
	cli.StringFlag{
		EnvVar: "WOODPECKER_NETRC_MACHINE",
		Name:   "netrc-machine",
	},
	//
	// metadata parameters
	//
	cli.StringFlag{
		EnvVar: "CI_SYSTEM_ARCH",
		Name:   "system-arch",
		Value:  "linux/amd64",
	},
	cli.StringFlag{
		EnvVar: "CI_SYSTEM_NAME",
		Name:   "system-name",
		Value:  "woodpecker-cli",
	},
	cli.StringFlag{
		EnvVar: "CI_SYSTEM_LINK",
		Name:   "system-link",
		Value:  "https://github.com/woodpecker-ci/woodpecker",
	},
	cli.StringFlag{
		EnvVar: "CI_REPO_NAME",
		Name:   "repo-name",
	},
	cli.StringFlag{
		EnvVar: "CI_REPO_LINK",
		Name:   "repo-link",
	},
	cli.StringFlag{
		EnvVar: "CI_REPO_REMOTE",
		Name:   "repo-remote-url",
	},
	cli.StringFlag{
		EnvVar: "CI_REPO_PRIVATE",
		Name:   "repo-private",
	},
	cli.IntFlag{
		EnvVar: "CI_BUILD_NUMBER",
		Name:   "build-number",
	},
	cli.IntFlag{
		EnvVar: "CI_PARENT_BUILD_NUMBER",
		Name:   "parent-build-number",
	},
	cli.Int64Flag{
		EnvVar: "CI_BUILD_CREATED",
		Name:   "build-created",
	},
	cli.Int64Flag{
		EnvVar: "CI_BUILD_STARTED",
		Name:   "build-started",
	},
	cli.Int64Flag{
		EnvVar: "CI_BUILD_FINISHED",
		Name:   "build-finished",
	},
	cli.StringFlag{
		EnvVar: "CI_BUILD_STATUS",
		Name:   "build-status",
	},
	cli.StringFlag{
		EnvVar: "CI_BUILD_EVENT",
		Name:   "build-event",
	},
	cli.StringFlag{
		EnvVar: "CI_BUILD_LINK",
		Name:   "build-link",
	},
	cli.StringFlag{
		EnvVar: "CI_BUILD_TARGET",
		Name:   "build-target",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_SHA",
		Name:   "commit-sha",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_REF",
		Name:   "commit-ref",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_REFSPEC",
		Name:   "commit-refspec",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_BRANCH",
		Name:   "commit-branch",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_MESSAGE",
		Name:   "commit-message",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_AUTHOR",
		Name:   "commit-author-name",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_AUTHOR_AVATAR",
		Name:   "commit-author-avatar",
	},
	cli.StringFlag{
		EnvVar: "CI_COMMIT_AUTHOR_EMAIL",
		Name:   "commit-author-email",
	},
	cli.IntFlag{
		EnvVar: "CI_PREV_BUILD_NUMBER",
		Name:   "prev-build-number",
	},
	cli.Int64Flag{
		EnvVar: "CI_PREV_BUILD_CREATED",
		Name:   "prev-build-created",
	},
	cli.Int64Flag{
		EnvVar: "CI_PREV_BUILD_STARTED",
		Name:   "prev-build-started",
	},
	cli.Int64Flag{
		EnvVar: "CI_PREV_BUILD_FINISHED",
		Name:   "prev-build-finished",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_BUILD_STATUS",
		Name:   "prev-build-status",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_BUILD_EVENT",
		Name:   "prev-build-event",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_BUILD_LINK",
		Name:   "prev-build-link",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_SHA",
		Name:   "prev-commit-sha",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_REF",
		Name:   "prev-commit-ref",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_REFSPEC",
		Name:   "prev-commit-refspec",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_BRANCH",
		Name:   "prev-commit-branch",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_MESSAGE",
		Name:   "prev-commit-message",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_AUTHOR",
		Name:   "prev-commit-author-name",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_AUTHOR_AVATAR",
		Name:   "prev-commit-author-avatar",
	},
	cli.StringFlag{
		EnvVar: "CI_PREV_COMMIT_AUTHOR_EMAIL",
		Name:   "prev-commit-author-email",
	},
	cli.IntFlag{
		EnvVar: "CI_BUILD_JOB_NUMBER",
		Name:   "job-number",
	},
	cli.StringSliceFlag{
		EnvVar: "WOODPECKER_ENV",
		Name:   "env",
	},
}

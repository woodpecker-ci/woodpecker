// Copyright 2019 Laszlo Fogas
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

package main

import (
	"os"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/shared/constant"
)

var flags = []cli.Flag{
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_LOG_LEVEL"},
		Name:    "log-level",
		Usage:   "set logging level",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_DEBUG_PRETTY"},
		Name:    "pretty",
		Usage:   "enable pretty-printed debug output",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_DEBUG_NOCOLOR"},
		Name:    "nocolor",
		Usage:   "disable colored debug output",
		Value:   true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_HOST"},
		Name:    "server-host",
		Usage:   "server fully qualified url (<scheme>://<host>)",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER_ADDR"},
		Name:    "server-addr",
		Usage:   "server address",
		Value:   ":8000",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER_CERT"},
		Name:    "server-cert",
		Usage:   "server ssl cert path",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER_KEY"},
		Name:    "server-key",
		Usage:   "server ssl key path",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_LETS_ENCRYPT_EMAIL"},
		Name:    "lets-encrypt-email",
		Usage:   "let's encrypt email",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_LETS_ENCRYPT"},
		Name:    "lets-encrypt",
		Usage:   "enable let's encrypt",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_GRPC_ADDR"},
		Name:    "grpc-addr",
		Usage:   "grpc address",
		Value:   ":9000",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_ADMIN"},
		Name:    "admin",
		Usage:   "list of admin users",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_ORGS"},
		Name:    "orgs",
		Usage:   "list of approved organizations",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_REPO_OWNERS"},
		Name:    "repo-owners",
		Usage:   "List of syncable repo owners",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_OPEN"},
		Name:    "open",
		Usage:   "enable open user registration",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_AUTHENTICATE_PUBLIC_REPOS"},
		Name:    "authenticate-public-repos",
		Usage:   "Always use authentication to clone repositories even if they are public. Needed if the SCM requires to always authenticate as used by many companies.",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS"},
		Name:    "default-cancel-previous-pipeline-events",
		Usage:   "List of event names that will be canceled when a new pipeline for the same context (tag, branch) is created.",
		Value:   cli.NewStringSlice("push", "pull_request"),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DEFAULT_CLONE_IMAGE"},
		Name:    "default-clone-image",
		Usage:   "The default docker image to be used when cloning the repo",
		Value:   constant.DefaultCloneImage,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DOCS"},
		Name:    "docs",
		Usage:   "link to user documentation",
		Value:   "https://woodpecker-ci.org/",
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_SESSION_EXPIRES"},
		Name:    "session-expires",
		Usage:   "session expiration time",
		Value:   time.Hour * 72,
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_ESCALATE"},
		Name:    "escalate",
		Usage:   "images to run in privileged mode",
		Value:   cli.NewStringSlice(constant.PrivilegedPlugins...),
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_VOLUME"},
		Name:    "volume",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DOCKER_CONFIG"},
		Name:    "docker-config",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_ENVIRONMENT"},
		Name:    "environment",
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_NETWORK"},
		Name:    "network",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_AGENT_SECRET"},
		Name:     "agent-secret",
		Usage:    "server-agent shared password",
		FilePath: os.Getenv("WOODPECKER_AGENT_SECRET_FILE"),
	},
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_KEEPALIVE_MIN_TIME"},
		Name:    "keepalive-min-time",
		Usage:   "server-side enforcement policy on the minimum amount of time a client should wait before sending a keepalive ping.",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SECRET_ENDPOINT"},
		Name:    "secret-service",
		Usage:   "secret plugin endpoint",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_REGISTRY_ENDPOINT"},
		Name:    "registry-service",
		Usage:   "registry plugin endpoint",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_CONFIG_SERVICE_ENDPOINT"},
		Name:    "config-service-endpoint",
		Usage:   "url used for calling configuration service endpoint",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DATABASE_DRIVER"},
		Name:    "driver",
		Usage:   "database driver",
		Value:   "sqlite3",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_DATABASE_DATASOURCE"},
		Name:     "datasource",
		Usage:    "database driver configuration string",
		Value:    "woodpecker.sqlite",
		FilePath: os.Getenv("WOODPECKER_DATABASE_DATASOURCE_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_PROMETHEUS_AUTH_TOKEN"},
		Name:     "prometheus-auth-token",
		Usage:    "token to secure prometheus metrics endpoint",
		Value:    "",
		FilePath: os.Getenv("WOODPECKER_PROMETHEUS_AUTH_TOKEN_FILE"),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_STATUS_CONTEXT", "WOODPECKER_GITHUB_CONTEXT", "WOODPECKER_GITEA_CONTEXT"},
		Name:    "status-context",
		Usage:   "status context prefix",
		Value:   "ci/woodpecker",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_STATUS_CONTEXT_FORMAT"},
		Name:    "status-context-format",
		Usage:   "status context format",
		Value:   "{{ .context }}/{{ .event }}/{{ .pipeline }}",
	},
	//
	// resource limit parameters
	//
	&cli.DurationFlag{
		EnvVars: []string{"WOODPECKER_FORGE_TIMEOUT"},
		Name:    "forge-timeout",
		Usage:   "how many seconds before timeout when fetching the Woodpecker configuration from a Forge",
		Value:   time.Second * 3,
	},
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_LIMIT_MEM_SWAP"},
		Name:    "limit-mem-swap",
		Usage:   "maximum swappable memory allowed in bytes",
	},
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_LIMIT_MEM"},
		Name:    "limit-mem",
		Usage:   "maximum memory allowed in bytes",
	},
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_LIMIT_SHM_SIZE"},
		Name:    "limit-shm-size",
		Usage:   "docker compose /dev/shm allowed in bytes",
	},
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_LIMIT_CPU_QUOTA"},
		Name:    "limit-cpu-quota",
		Usage:   "impose a cpu quota",
	},
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_LIMIT_CPU_SHARES"},
		Name:    "limit-cpu-shares",
		Usage:   "change the cpu shares",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_LIMIT_CPU_SET"},
		Name:    "limit-cpu-set",
		Usage:   "set the cpus allowed to execute containers",
	},
	//
	// GitHub
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITHUB"},
		Name:    "github",
		Usage:   "github driver is enabled",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_GITHUB_URL"},
		Name:    "github-server",
		Usage:   "github server address",
		Value:   "https://github.com",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GITHUB_CLIENT"},
		Name:     "github-client",
		Usage:    "github oauth2 client id",
		FilePath: os.Getenv("WOODPECKER_GITHUB_CLIENT_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GITHUB_SECRET"},
		Name:     "github-secret",
		Usage:    "github oauth2 client secret",
		FilePath: os.Getenv("WOODPECKER_GITHUB_SECRET_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITHUB_MERGE_REF"},
		Name:    "github-merge-ref",
		Usage:   "github pull requests use merge ref",
		Value:   true,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITHUB_SKIP_VERIFY"},
		Name:    "github-skip-verify",
		Usage:   "github skip ssl verification",
	},
	//
	// Gogs
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GOGS"},
		Name:    "gogs",
		Usage:   "gogs driver is enabled",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_GOGS_URL"},
		Name:    "gogs-server",
		Usage:   "gogs server address",
		Value:   "https://try.gogs.io",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GOGS_GIT_USERNAME"},
		Name:     "gogs-git-username",
		Usage:    "gogs service account username",
		FilePath: os.Getenv("WOODPECKER_GOGS_GIT_USERNAME_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GOGS_GIT_PASSWORD"},
		Name:     "gogs-git-password",
		Usage:    "gogs service account password",
		FilePath: os.Getenv("WOODPECKER_GOGS_GIT_PASSWORD_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GOGS_PRIVATE_MODE"},
		Name:    "gogs-private-mode",
		Usage:   "gogs private mode enabled",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GOGS_SKIP_VERIFY"},
		Name:    "gogs-skip-verify",
		Usage:   "gogs skip ssl verification",
	},
	//
	// Gitea
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITEA"},
		Name:    "gitea",
		Usage:   "gitea driver is enabled",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_GITEA_URL"},
		Name:    "gitea-server",
		Usage:   "gitea server address",
		Value:   "https://try.gitea.io",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GITEA_CLIENT"},
		Name:     "gitea-client",
		Usage:    "gitea oauth2 client id",
		FilePath: os.Getenv("WOODPECKER_GITEA_CLIENT_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GITEA_SECRET"},
		Name:     "gitea-secret",
		Usage:    "gitea oauth2 client secret",
		FilePath: os.Getenv("WOODPECKER_GITEA_SECRET_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITEA_SKIP_VERIFY"},
		Name:    "gitea-skip-verify",
		Usage:   "gitea skip ssl verification",
	},
	//
	// Bitbucket
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BITBUCKET"},
		Name:    "bitbucket",
		Usage:   "bitbucket driver is enabled",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_BITBUCKET_CLIENT"},
		Name:     "bitbucket-client",
		Usage:    "bitbucket oauth2 client id",
		FilePath: os.Getenv("WOODPECKER_BITBUCKET_CLIENT_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_BITBUCKET_SECRET"},
		Name:     "bitbucket-secret",
		Usage:    "bitbucket oauth2 client secret",
		FilePath: os.Getenv("WOODPECKER_BITBUCKET_SECRET_FILE"),
	},
	//
	// Gitlab
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITLAB"},
		Name:    "gitlab",
		Usage:   "gitlab driver is enabled",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_GITLAB_URL"},
		Name:    "gitlab-server",
		Usage:   "gitlab server address",
		Value:   "https://gitlab.com",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GITLAB_CLIENT"},
		Name:     "gitlab-client",
		Usage:    "gitlab oauth2 client id",
		FilePath: os.Getenv("WOODPECKER_GITLAB_CLIENT_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GITLAB_SECRET"},
		Name:     "gitlab-secret",
		Usage:    "gitlab oauth2 client secret",
		FilePath: os.Getenv("WOODPECKER_GITLAB_SECRET_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITLAB_SKIP_VERIFY"},
		Name:    "gitlab-skip-verify",
		Usage:   "gitlab skip ssl verification",
	},
	//
	// Bitbucket Stash
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_STASH"},
		Name:    "stash",
		Usage:   "stash driver is enabled",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_STASH_URL"},
		Name:    "stash-server",
		Usage:   "stash server address",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_STASH_CONSUMER_KEY"},
		Name:     "stash-consumer-key",
		Usage:    "stash oauth1 consumer key",
		FilePath: os.Getenv("WOODPECKER_STASH_CONSUMER_KEY_FILE"),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_STASH_CONSUMER_RSA"},
		Name:    "stash-consumer-rsa",
		Usage:   "stash oauth1 private key file",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_STASH_CONSUMER_RSA_STRING"},
		Name:    "stash-consumer-rsa-string",
		Usage:   "stash oauth1 private key string",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_STASH_GIT_USERNAME"},
		Name:     "stash-git-username",
		Usage:    "stash service account username",
		FilePath: os.Getenv("WOODPECKER_STASH_GIT_USERNAME_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_STASH_GIT_PASSWORD"},
		Name:     "stash-git-password",
		Usage:    "stash service account password",
		FilePath: os.Getenv("WOODPECKER_STASH_GIT_PASSWORD_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_STASH_SKIP_VERIFY"},
		Name:    "stash-skip-verify",
		Usage:   "stash skip ssl verification",
	},
	//
	// Coding
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_CODING"},
		Name:    "coding",
		Usage:   "coding driver is enabled",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_CODING_URL"},
		Name:    "coding-server",
		Usage:   "coding server address",
		Value:   "https://coding.net",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_CODING_CLIENT"},
		Name:     "coding-client",
		Usage:    "coding oauth2 client id",
		FilePath: os.Getenv("WOODPECKER_CODING_CLIENT_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_CODING_SECRET"},
		Name:     "coding-secret",
		Usage:    "coding oauth2 client secret",
		FilePath: os.Getenv("WOODPECKER_CODING_SECRET_FILE"),
	},
	&cli.StringSliceFlag{
		EnvVars: []string{"WOODPECKER_CODING_SCOPE"},
		Name:    "coding-scope",
		Usage:   "coding oauth scope",
		Value: cli.NewStringSlice(
			"user",
			"project",
			"project:depot",
		),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_CODING_GIT_MACHINE"},
		Name:    "coding-git-machine",
		Usage:   "coding machine name",
		Value:   "git.coding.net",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_CODING_GIT_USERNAME"},
		Name:     "coding-git-username",
		Usage:    "coding machine user username",
		FilePath: os.Getenv("WOODPECKER_CODING_GIT_USERNAME_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_CODING_GIT_PASSWORD"},
		Name:     "coding-git-password",
		Usage:    "coding machine user password",
		FilePath: os.Getenv("WOODPECKER_CODING_GIT_PASSWORD_FILE"),
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_CODING_SKIP_VERIFY"},
		Name:    "coding-skip-verify",
		Usage:   "coding skip ssl verification",
	},
	//
	// development flags
	//
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DEV_WWW_PROXY"},
		Name:    "www-proxy",
		Usage:   "serve the website by using a proxy (used for development)",
		Hidden:  true,
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DEV_OAUTH_HOST"},
		Name:    "server-dev-oauth-host",
		Usage:   "server fully qualified url (<scheme>://<host>) used for oauth redirect (used for development)",
		Value:   "",
		Hidden:  true,
	},
	//
	// misc
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_FLAT_PERMISSIONS"},
		Name:    "flat-permissions",
		Usage:   "no forge call for permissions should be made",
		Hidden:  true,
		// TODO(485) temporary workaround to not hit api rate limits
	},
	//
	// secrets encryption in DB
	//
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_ENCRYPTION_KEY"},
		Name:     "encryption-raw-key",
		Usage:    "Raw encryption key",
		FilePath: os.Getenv("WOODPECKER_ENCRYPTION_KEY_FILE"),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE"},
		Name:    "encryption-tink-keyset",
		Usage:   "Google tink AEAD-compatible keyset file to encrypt secrets in DB",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_ENCRYPTION_DISABLE"},
		Name:    "encryption-disable-flag",
		Usage:   "Flag to decrypt all encrypted data and disable encryption on server",
	},
}

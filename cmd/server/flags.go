// Copyright 2023 Woodpecker Authors
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
	"time"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cmd/common"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

var flags = append([]cli.Flag{
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_LOG_XORM"),
		Name:    "log-xorm",
		Usage:   "enable xorm logging",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_LOG_XORM_SQL"),
		Name:    "log-xorm-sql",
		Usage:   "enable xorm sql command logging",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_HOST"),
		Name:    "server-host",
		Usage:   "server fully qualified url (<scheme>://<host>[/<prefixpath>])",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_WEBHOOK_HOST"),
		Name:    "server-webhook-host",
		Usage:   "server fully qualified url for forge's Webhooks (<scheme>://<host>[/<prefixpath>])",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SERVER_ADDR"),
		Name:    "server-addr",
		Usage:   "server address",
		Value:   ":8000",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SERVER_ADDR_TLS"),
		Name:    "server-addr-tls",
		Usage:   "port https with tls (:443)",
		Value:   ":443",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SERVER_CERT"),
		Name:    "server-cert",
		Usage:   "server ssl cert path",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SERVER_KEY"),
		Name:    "server-key",
		Usage:   "server ssl key path",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_CUSTOM_CSS_FILE"),
		Name:    "custom-css-file",
		Usage:   "file path for the server to serve a custom .CSS file, used for customizing the UI",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_CUSTOM_JS_FILE"),
		Name:    "custom-js-file",
		Usage:   "file path for the server to serve a custom .JS file, used for customizing the UI",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_LETS_ENCRYPT_EMAIL"),
		Name:    "lets-encrypt-email",
		Usage:   "let's encrypt email",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_LETS_ENCRYPT"),
		Name:    "lets-encrypt",
		Usage:   "enable let's encrypt",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GRPC_ADDR"),
		Name:    "grpc-addr",
		Usage:   "grpc address",
		Value:   ":9000",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GRPC_SECRET"),
		Name:    "grpc-secret",
		Usage:   "grpc jwt secret",
		Value:   "secret",
		// TODO: FilePath: os.Getenv("WOODPECKER_GRPC_SECRET_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_METRICS_SERVER_ADDR"),
		Name:    "metrics-server-addr",
		Usage:   "metrics server address",
		Value:   "",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_ADMIN"),
		Name:    "admin",
		Usage:   "list of admin users",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_ORGS"),
		Name:    "orgs",
		Usage:   "list of approved organizations",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_REPO_OWNERS"),
		Name:    "repo-owners",
		Usage:   "List of syncable repo owners",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_OPEN"),
		Name:    "open",
		Usage:   "enable open user registration",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_AUTHENTICATE_PUBLIC_REPOS"),
		Name:    "authenticate-public-repos",
		Usage:   "Always use authentication to clone repositories even if they are public. Needed if the SCM requires to always authenticate as used by many companies.",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS"),
		Name:    "default-cancel-previous-pipeline-events",
		Usage:   "List of event names that will be canceled when a new pipeline for the same context (tag, branch) is created.",
		Value:   []string{"push", "pull_request"},
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_DEFAULT_CLONE_IMAGE"),
		Name:    "default-clone-image",
		Usage:   "The default docker image to be used when cloning the repo",
		Value:   constant.DefaultCloneImage,
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_DEFAULT_PIPELINE_TIMEOUT"),
		Name:    "default-pipeline-timeout",
		Usage:   "The default time in minutes for a repo in minutes before a pipeline gets killed",
		Value:   60,
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_MAX_PIPELINE_TIMEOUT"),
		Name:    "max-pipeline-timeout",
		Usage:   "The maximum time in minutes you can set in the repo settings before a pipeline gets killed",
		Value:   120,
	},
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_SESSION_EXPIRES"),
		Name:    "session-expires",
		Usage:   "session expiration time",
		Value:   time.Hour * 72,
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_ESCALATE"),
		Name:    "escalate",
		Usage:   "images to run in privileged mode",
		Value:   constant.PrivilegedPlugins,
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_VOLUME"),
		Name:    "volume",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_DOCKER_CONFIG"),
		Name:    "docker-config",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_ENVIRONMENT"),
		Name:    "environment",
	},
	&cli.StringSliceFlag{
		Sources: cli.EnvVars("WOODPECKER_NETWORK"),
		Name:    "network",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_AGENT_SECRET"),
		Name:    "agent-secret",
		Usage:   "server-agent shared password",
		// TODO: FilePath: os.Getenv("WOODPECKER_AGENT_SECRET_FILE"),
	},
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_KEEPALIVE_MIN_TIME"),
		Name:    "keepalive-min-time",
		Usage:   "server-side enforcement policy on the minimum amount of time a client should wait before sending a keepalive ping.",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_SECRET_ENDPOINT"),
		Name:    "secret-service",
		Usage:   "secret plugin endpoint",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_REGISTRY_ENDPOINT"),
		Name:    "registry-service",
		Usage:   "registry plugin endpoint",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_CONFIG_SERVICE_ENDPOINT"),
		Name:    "config-service-endpoint",
		Usage:   "url used for calling configuration service endpoint",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_DATABASE_DRIVER"),
		Name:    "driver",
		Usage:   "database driver",
		Value:   "sqlite3",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_DATABASE_DATASOURCE"),
		Name:    "datasource",
		Usage:   "database driver configuration string",
		Value:   "woodpecker.sqlite",
		// TODO: FilePath: os.Getenv("WOODPECKER_DATABASE_DATASOURCE_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_PROMETHEUS_AUTH_TOKEN"),
		Name:    "prometheus-auth-token",
		Usage:   "token to secure prometheus metrics endpoint",
		Value:   "",
		// TODO: FilePath: os.Getenv("WOODPECKER_PROMETHEUS_AUTH_TOKEN_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_STATUS_CONTEXT", "WOODPECKER_GITHUB_CONTEXT", "WOODPECKER_GITEA_CONTEXT"),
		Name:    "status-context",
		Usage:   "status context prefix",
		Value:   "ci/woodpecker",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_STATUS_CONTEXT_FORMAT"),
		Name:    "status-context-format",
		Usage:   "status context format",
		Value:   "{{ .context }}/{{ .event }}/{{ .workflow }}{{if not (eq .axis_id 0)}}/{{.axis_id}}{{end}}",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_MIGRATIONS_ALLOW_LONG"),
		Name:    "migrations-allow-long",
		Value:   false,
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_ENABLE_SWAGGER"),
		Name:    "enable-swagger",
		Value:   true,
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
	// resource limit parameters
	//
	&cli.DurationFlag{
		Sources: cli.EnvVars("WOODPECKER_FORGE_TIMEOUT"),
		Name:    "forge-timeout",
		Usage:   "how many seconds before timeout when fetching the Woodpecker configuration from a Forge",
		Value:   time.Second * 3,
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_LIMIT_MEM_SWAP"),
		Name:    "limit-mem-swap",
		Usage:   "maximum swappable memory allowed in bytes",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_LIMIT_MEM"),
		Name:    "limit-mem",
		Usage:   "maximum memory allowed in bytes",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_LIMIT_SHM_SIZE"),
		Name:    "limit-shm-size",
		Usage:   "docker compose /dev/shm allowed in bytes",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_LIMIT_CPU_QUOTA"),
		Name:    "limit-cpu-quota",
		Usage:   "impose a cpu quota",
	},
	&cli.IntFlag{
		Sources: cli.EnvVars("WOODPECKER_LIMIT_CPU_SHARES"),
		Name:    "limit-cpu-shares",
		Usage:   "change the cpu shares",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_LIMIT_CPU_SET"),
		Name:    "limit-cpu-set",
		Usage:   "set the cpus allowed to execute containers",
	},
	//
	// GitHub
	//
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITHUB"),
		Name:    "github",
		Usage:   "github driver is enabled",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITHUB_URL"),
		Name:    "github-server",
		Usage:   "github server address",
		Value:   "https://github.com",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITHUB_CLIENT"),
		Name:    "github-client",
		Usage:   "github oauth2 client id",
		// TODO: FilePath: os.Getenv("WOODPECKER_GITHUB_CLIENT_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITHUB_SECRET"),
		Name:    "github-secret",
		Usage:   "github oauth2 client secret",
		// TODO: FilePath: os.Getenv("WOODPECKER_GITHUB_SECRET_FILE"),
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITHUB_MERGE_REF"),
		Name:    "github-merge-ref",
		Usage:   "github pull requests use merge ref",
		Value:   true,
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITHUB_SKIP_VERIFY"),
		Name:    "github-skip-verify",
		Usage:   "github skip ssl verification",
	},
	//
	// Gitea
	//
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITEA"),
		Name:    "gitea",
		Usage:   "gitea driver is enabled",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITEA_URL"),
		Name:    "gitea-server",
		Usage:   "gitea server address",
		Value:   "https://try.gitea.io",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITEA_CLIENT"),
		Name:    "gitea-client",
		Usage:   "gitea oauth2 client id",
		// TODO: FilePath: os.Getenv("WOODPECKER_GITEA_CLIENT_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITEA_SECRET"),
		Name:    "gitea-secret",
		Usage:   "gitea oauth2 client secret",
		// TODO: FilePath: os.Getenv("WOODPECKER_GITEA_SECRET_FILE"),
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITEA_SKIP_VERIFY"),
		Name:    "gitea-skip-verify",
		Usage:   "gitea skip ssl verification",
	},
	//
	// Bitbucket
	//
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BITBUCKET"),
		Name:    "bitbucket",
		Usage:   "bitbucket driver is enabled",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BITBUCKET_CLIENT"),
		Name:    "bitbucket-client",
		Usage:   "bitbucket oauth2 client id",
		// TODO: FilePath: os.Getenv("WOODPECKER_BITBUCKET_CLIENT_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_BITBUCKET_SECRET"),
		Name:    "bitbucket-secret",
		Usage:   "bitbucket oauth2 client secret",
		// TODO: FilePath: os.Getenv("WOODPECKER_BITBUCKET_SECRET_FILE"),
	},
	//
	// Gitlab
	//
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITLAB"),
		Name:    "gitlab",
		Usage:   "gitlab driver is enabled",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITLAB_URL"),
		Name:    "gitlab-server",
		Usage:   "gitlab server address",
		Value:   "https://gitlab.com",
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITLAB_CLIENT"),
		Name:    "gitlab-client",
		Usage:   "gitlab oauth2 client id",
		// TODO: FilePath: os.Getenv("WOODPECKER_GITLAB_CLIENT_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_GITLAB_SECRET"),
		Name:    "gitlab-secret",
		Usage:   "gitlab oauth2 client secret",
		// TODO: FilePath: os.Getenv("WOODPECKER_GITLAB_SECRET_FILE"),
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_GITLAB_SKIP_VERIFY"),
		Name:    "gitlab-skip-verify",
		Usage:   "gitlab skip ssl verification",
	},
	//
	// development flags
	//
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_DEV_WWW_PROXY"),
		Name:    "www-proxy",
		Usage:   "serve the website by using a proxy (used for development)",
		Hidden:  true,
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_DEV_OAUTH_HOST"),
		Name:    "server-dev-oauth-host",
		Usage:   "server fully qualified url (<scheme>://<host>[/<prefixpath>]) used for oauth redirect (used for development)",
		Value:   "",
		Hidden:  true,
	},
	//
	// secrets encryption in DB
	//
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_ENCRYPTION_KEY"),
		Name:    "encryption-raw-key",
		Usage:   "Raw encryption key",
		// TODO: FilePath: os.Getenv("WOODPECKER_ENCRYPTION_KEY_FILE"),
	},
	&cli.StringFlag{
		Sources: cli.EnvVars("WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE"),
		Name:    "encryption-tink-keyset",
		Usage:   "Google tink AEAD-compatible keyset file to encrypt secrets in DB",
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_ENCRYPTION_DISABLE"),
		Name:    "encryption-disable-flag",
		Usage:   "Flag to decrypt all encrypted data and disable encryption on server",
	},
}, common.GlobalLoggerFlags...)

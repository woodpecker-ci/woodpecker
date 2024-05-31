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

	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger"
)

var flags = append([]cli.Flag{
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_LOG_XORM"},
		Name:    "log-xorm",
		Usage:   "enable xorm logging",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_LOG_XORM_SQL"},
		Name:    "log-xorm-sql",
		Usage:   "enable xorm sql command logging",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_HOST"},
		Name:    "server-host",
		Usage:   "server fully qualified url. Format: <scheme>://<host>[/<prefixpath>]",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER_ADDR"},
		Name:    "server-addr",
		Usage:   "server address",
		Value:   ":8000",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_SERVER_ADDR_TLS"},
		Name:    "server-addr-tls",
		Usage:   "port https with tls (:443)",
		Value:   ":443",
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
		EnvVars: []string{"WOODPECKER_CUSTOM_CSS_FILE"},
		Name:    "custom-css-file",
		Usage:   "file path for the server to serve a custom .CSS file, used for customizing the UI",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_CUSTOM_JS_FILE"},
		Name:    "custom-js-file",
		Usage:   "file path for the server to serve a custom .JS file, used for customizing the UI",
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
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_GRPC_SECRET"},
		Name:     "grpc-secret",
		Usage:    "grpc jwt secret",
		Value:    "secret",
		FilePath: os.Getenv("WOODPECKER_GRPC_SECRET_FILE"),
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_METRICS_SERVER_ADDR"},
		Name:    "metrics-server-addr",
		Usage:   "metrics server address",
		Value:   "",
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
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_DEFAULT_PIPELINE_TIMEOUT"},
		Name:    "default-pipeline-timeout",
		Usage:   "The default time in minutes for a repo in minutes before a pipeline gets killed",
		Value:   60,
	},
	&cli.Int64Flag{
		EnvVars: []string{"WOODPECKER_MAX_PIPELINE_TIMEOUT"},
		Name:    "max-pipeline-timeout",
		Usage:   "The maximum time in minutes you can set in the repo settings before a pipeline gets killed",
		Value:   120,
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
		Value:    datasourceDefaultValue(),
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
		Value:   "{{ .context }}/{{ .event }}/{{ .workflow }}{{if not (eq .axis_id 0)}}/{{.axis_id}}{{end}}",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_MIGRATIONS_ALLOW_LONG"},
		Name:    "migrations-allow-long",
		Value:   false,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_ENABLE_SWAGGER"},
		Name:    "enable-swagger",
		Value:   true,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_DISABLE_VERSION_CHECK"},
		Usage:   "Disable version check in admin web ui.",
		Name:    "skip-version-check",
	},
	//
	// backend options for pipeline compiler
	//
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_NO_PROXY", "NO_PROXY", "no_proxy"},
		Usage:   "if set, pass the environment variable down as \"NO_PROXY\" to steps",
		Name:    "backend-no-proxy",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_HTTP_PROXY", "HTTP_PROXY", "http_proxy"},
		Usage:   "if set, pass the environment variable down as \"HTTP_PROXY\" to steps",
		Name:    "backend-http-proxy",
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_BACKEND_HTTPS_PROXY", "HTTPS_PROXY", "https_proxy"},
		Usage:   "if set, pass the environment variable down as \"HTTPS_PROXY\" to steps",
		Name:    "backend-https-proxy",
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
	&cli.UintFlag{
		EnvVars: []string{"WOODPECKER_FORGE_RETRY"},
		Name:    "forge-retry",
		Usage:   "How many retries of fetching the Woodpecker configuration from a forge are done before we fail",
		Value:   3,
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
	&cli.StringFlag{
		Name:    "forge-url",
		Usage:   "url of the forge",
		EnvVars: []string{"WOODPECKER_FORGE_URL", "WOODPECKER_GITHUB_URL", "WOODPECKER_GITLAB_URL", "WOODPECKER_GITEA_URL", "WOODPECKER_FORGEJO_URL", "WOODPECKER_BITBUCKET_URL"},
	},
	&cli.StringFlag{
		Name:     "forge-oauth-client",
		Usage:    "oauth2 client id",
		EnvVars:  []string{"WOODPECKER_FORGE_CLIENT", "WOODPECKER_GITHUB_CLIENT", "WOODPECKER_GITLAB_CLIENT", "WOODPECKER_GITEA_CLIENT", "WOODPECKER_FORGEJO_CLIENT", "WOODPECKER_BITBUCKET_CLIENT", "WOODPECKER_BITBUCKET_DC_CLIENT_ID"},
		FilePath: getFirstNonEmptyEnvVar([]string{"WOODPECKER_FORGE_CLIENT_FILE", "WOODPECKER_GITHUB_CLIENT_FILE", "WOODPECKER_GITLAB_CLIENT_FILE", "WOODPECKER_GITEA_CLIENT_FILE", "WOODPECKER_FORGEJO_CLIENT_FILE", "WOODPECKER_BITBUCKET_CLIENT_FILE", "WOODPECKER_BITBUCKET_DC_CLIENT_ID_FILE"}),
	},
	&cli.StringFlag{
		Name:     "forge-oauth-secret",
		Usage:    "oauth2 client secret",
		EnvVars:  []string{"WOODPECKER_FORGE_SECRET", "WOODPECKER_GITHUB_SECRET", "WOODPECKER_GITLAB_SECRET", "WOODPECKER_GITEA_SECRET", "WOODPECKER_FORGEJO_SECRET", "WOODPECKER_BITBUCKET_SECRET", "WOODPECKER_BITBUCKET_DC_CLIENT_SECRET"},
		FilePath: getFirstNonEmptyEnvVar([]string{"WOODPECKER_FORGE_SECRET_FILE", "WOODPECKER_GITHUB_SECRET_FILE", "WOODPECKER_GITLAB_SECRET_FILE", "WOODPECKER_GITEA_SECRET_FILE", "WOODPECKER_FORGEJO_SECRET_FILE", "WOODPECKER_BITBUCKET_SECRET_FILE", "WOODPECKER_BITBUCKET_DC_CLIENT_SECRET_FILE"}),
	},
	&cli.BoolFlag{
		Name:    "forge-skip-verify",
		Usage:   "skip ssl verification",
		EnvVars: []string{"WOODPECKER_FORGE_SKIP_VERIFY", "WOODPECKER_GITHUB_SKIP_VERIFY", "WOODPECKER_GITLAB_SKIP_VERIFY", "WOODPECKER_GITEA_SKIP_VERIFY", "WOODPECKER_FORGEJO_SKIP_VERIFY", "WOODPECKER_BITBUCKET_SKIP_VERIFY"},
	},
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_EXPERT_FORGE_OAUTH_HOST", "WOODPECKER_DEV_GITEA_OAUTH_URL"}, // TODO: remove WOODPECKER_DEV_GITEA_OAUTH_URL in next major release
		Name:    "forge-oauth-host",
		Usage:   "!!!for experts!!! fully qualified public forge url. Use it if your forge url WOODPECKER_FORGE_URL or WOODPECKER_GITEA_URL, ... isn't a public url. Format: <scheme>://<host>[/<prefixpath>]",
	},
	//
	// Addon
	//
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_ADDON_FORGE"},
		Name:    "addon-forge",
		Usage:   "path to forge addon executable",
	},
	//
	// GitHub
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITHUB"},
		Name:    "github",
		Usage:   "github driver is enabled",
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITHUB_MERGE_REF"},
		Name:    "github-merge-ref",
		Usage:   "github pull requests use merge ref",
		Value:   true,
	},
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITHUB_PUBLIC_ONLY"},
		Name:    "github-public-only",
		Usage:   "github tokens should only get access to public repos",
		Value:   false,
	},
	//
	// Gitea
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITEA"},
		Name:    "gitea",
		Usage:   "gitea driver is enabled",
	},
	//
	// Forgejo
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_FORGEJO"},
		Name:    "forgejo",
		Usage:   "forgejo driver is enabled",
	},
	//
	// Bitbucket
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BITBUCKET"},
		Name:    "bitbucket",
		Usage:   "bitbucket driver is enabled",
	},
	//
	// Gitlab
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_GITLAB"},
		Name:    "gitlab",
		Usage:   "gitlab driver is enabled",
	},
	//
	// Bitbucket DataCenter/Server (previously Stash)
	//
	&cli.BoolFlag{
		EnvVars: []string{"WOODPECKER_BITBUCKET_DC"},
		Name:    "bitbucket-dc",
		Usage:   "Bitbucket DataCenter/Server driver is enabled",
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_BITBUCKET_DC_GIT_USERNAME"},
		Name:     "bitbucket-dc-git-username",
		Usage:    "Bitbucket DataCenter/Server service account username",
		FilePath: os.Getenv("WOODPECKER_BITBUCKET_DC_GIT_USERNAME_FILE"),
	},
	&cli.StringFlag{
		EnvVars:  []string{"WOODPECKER_BITBUCKET_DC_GIT_PASSWORD"},
		Name:     "bitbucket-dc-git-password",
		Usage:    "Bitbucket DataCenter/Server service account password",
		FilePath: os.Getenv("WOODPECKER_BITBUCKET_DC_GIT_PASSWORD_FILE"),
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
	//
	// expert flags
	//
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_EXPERT_WEBHOOK_HOST", "WOODPECKER_WEBHOOK_HOST"}, // TODO: remove WOODPECKER_WEBHOOK_HOST in next major release
		Name:    "server-webhook-host",
		Usage:   "!!!for experts!!! fully qualified woodpecker server url called by forge's webhooks. Format: <scheme>://<host>[/<prefixpath>]",
	},
	// TODO: remove in next major release
	&cli.StringFlag{
		EnvVars: []string{"WOODPECKER_DEV_OAUTH_HOST"},
		Name:    "server-dev-oauth-host-deprecated",
		Usage:   "DEPRECATED: use WOODPECKER_EXPERT_FORGE_OAUTH_HOST instead\nfully qualified url used for oauth redirects. Format: <scheme>://<host>[/<prefixpath>]",
		Value:   "",
		Hidden:  true,
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
}, logger.GlobalLoggerFlags...)

// If woodpecker is running inside a container the default value for
// the datasource is different from running outside a container.
func datasourceDefaultValue() string {
	_, found := os.LookupEnv("WOODPECKER_IN_CONTAINER")
	if found {
		return "/var/lib/woodpecker/woodpecker.sqlite"
	}
	return "woodpecker.sqlite"
}

func getFirstNonEmptyEnvVar(envVars []string) string {
	for _, envVar := range envVars {
		val := os.Getenv(envVar)
		if val != "" {
			return val
		}
	}
	return ""
}

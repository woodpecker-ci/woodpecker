---
toc_max_heading_level: 3
---

# Environment variables

## Server

The following list describes all available server configuration options.

### `WOODPECKER_LOG_LEVEL`

> Default: empty

Configures the logging level. Possible values are `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled` and empty.

### `WOODPECKER_LOG_FILE`

> Default: `stderr`

Output destination for logs.
'stdout' and 'stderr' can be used as special keywords.

### `WOODPECKER_DATABASE_LOG`

> Default: `false`

Enable logging in database engine (currently xorm).

### `WOODPECKER_DATABASE_LOG_SQL`

> Default: `false`

Enable logging of sql commands.

### `WOODPECKER_DATABASE_MAX_CONNECTIONS`

> Default: `100`

Max database connections xorm is allowed create.

### `WOODPECKER_DATABASE_IDLE_CONNECTIONS`

> Default: `2`

Amount of database connections xorm will hold open.

### `WOODPECKER_DATABASE_CONNECTION_TIMEOUT`

> Default: `3 Seconds`

Time an active database connection is allowed to stay open.

### `WOODPECKER_DEBUG_PRETTY`

> Default: `false`

Enable pretty-printed debug output.

### `WOODPECKER_DEBUG_NOCOLOR`

> Default: `true`

Disable colored debug output.

### `WOODPECKER_HOST`

> Default: empty

Server fully qualified URL of the user-facing hostname, port (if not default for HTTP/HTTPS) and path prefix.

Examples:

- `WOODPECKER_HOST=http://woodpecker.example.org`
- `WOODPECKER_HOST=http://example.org/woodpecker`
- `WOODPECKER_HOST=http://example.org:1234/woodpecker`

### `WOODPECKER_SERVER_ADDR`

> Default: `:8000`

Configures the HTTP listener port.

### `WOODPECKER_SERVER_ADDR_TLS`

> Default: `:443`

Configures the HTTPS listener port when SSL is enabled.

### `WOODPECKER_SERVER_CERT`

> Default: empty

Path to an SSL certificate used by the server to accept HTTPS requests.

Example: `WOODPECKER_SERVER_CERT=/path/to/cert.pem`

### `WOODPECKER_SERVER_KEY`

> Default: empty

Path to an SSL certificate key used by the server to accept HTTPS requests.

Example: `WOODPECKER_SERVER_KEY=/path/to/key.pem`

### `WOODPECKER_CUSTOM_CSS_FILE`

> Default: empty

File path for the server to serve a custom .CSS file, used for customizing the UI.
Can be used for showing banner messages, logos, or environment-specific hints (a.k.a. white-labeling).
The file must be UTF-8 encoded, to ensure all special characters are preserved.

Example: `WOODPECKER_CUSTOM_CSS_FILE=/usr/local/www/woodpecker.css`

### `WOODPECKER_CUSTOM_JS_FILE`

> Default: empty

File path for the server to serve a custom .JS file, used for customizing the UI.
Can be used for showing banner messages, logos, or environment-specific hints (a.k.a. white-labeling).
The file must be UTF-8 encoded, to ensure all special characters are preserved.

Example: `WOODPECKER_CUSTOM_JS_FILE=/usr/local/www/woodpecker.js`

### `WOODPECKER_GRPC_ADDR`

> Default: `:9000`

Configures the gRPC listener port.

### `WOODPECKER_GRPC_SECRET`

> Default: `secret`

Configures the gRPC JWT secret.

### `WOODPECKER_GRPC_SECRET_FILE`

> Default: empty

Read the value for `WOODPECKER_GRPC_SECRET` from the specified filepath.

### `WOODPECKER_METRICS_SERVER_ADDR`

> Default: empty

Configures an unprotected metrics endpoint. An empty value disables the metrics endpoint completely.

Example: `:9001`

### `WOODPECKER_ADMIN`

> Default: empty

Comma-separated list of admin accounts.

Example: `WOODPECKER_ADMIN=user1,user2`

### `WOODPECKER_ORGS`

> Default: empty

Comma-separated list of approved organizations.

Example: `org1,org2`

### `WOODPECKER_REPO_OWNERS`

> Default: empty

Repositories by those owners will be allowed to be used in woodpecker.

Example: `user1,user2`

### `WOODPECKER_OPEN`

> Default: `false`

Enable to allow user registration.

### `WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`

> Default: `false`

Always use authentication to clone repositories even if they are public. Needed if the forge requires to always authenticate as used by many companies.

### `WOODPECKER_DEFAULT_ALLOW_PULL_REQUESTS`

> Default: `true`

The default setting for allowing pull requests on a repo.

### `WOODPECKER_DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS`

> Default: `pull_request, push`

List of event names that will be canceled when a new pipeline for the same context (tag, branch) is created.

### `WOODPECKER_DEFAULT_CLONE_PLUGIN`

> Default is defined in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/main/shared/constant/constant.go)

The default docker image to be used when cloning the repo.

It is also added to the trusted clone plugin list.

### `WOODPECKER_DEFAULT_WORKFLOW_LABELS`

> By default run workflows on any agent if no label conditions are set in workflow definition.

You can specify default label/platform conditions that will be used for agent selection for workflows that does not have labels conditions set.

Example: `platform=linux/amd64,backend=docker`

### `WOODPECKER_DEFAULT_PIPELINE_TIMEOUT`

> 60 (minutes)

The default time for a repo in minutes before a pipeline gets killed

### `WOODPECKER_MAX_PIPELINE_TIMEOUT`

> 120 (minutes)

The maximum time in minutes you can set in the repo settings before a pipeline gets killed

### `WOODPECKER_SESSION_EXPIRES`

> Default: `72h`

Configures the session expiration time.
Context: when someone does log into Woodpecker, a temporary session token is created.
As long as the session is valid (until it expires or log-out),
a user can log into Woodpecker, without re-authentication.

### `WOODPECKER_PLUGINS_PRIVILEGED`

Docker images to run in privileged mode. Only change if you are sure what you do!

You should specify the tag of your images too, as this enforces exact matches.

### WOODPECKER_PLUGINS_TRUSTED_CLONE

> Defaults are defined in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/main/shared/constant/constant.go)

Plugins which are trusted to handle the Git credential info in clone steps.
If a clone step use an image not in this list, Git credentials will not be injected and users have to use other methods (e.g. secrets) to clone non-public repos.

You should specify the tag of your images too, as this enforces exact matches.

<!--
### `WOODPECKER_VOLUME`
> Default: empty

Comma-separated list of Docker volumes that are mounted into every pipeline step.

Example: `WOODPECKER_VOLUME=/path/on/host:/path/in/container:rw`|
-->

### `WOODPECKER_DOCKER_CONFIG`

> Default: empty

Configures a specific private registry config for all pipelines.

Example: `WOODPECKER_DOCKER_CONFIG=/home/user/.docker/config.json`

<!--
### `WOODPECKER_ENVIRONMENT`
> Default: empty

TODO

### `WOODPECKER_NETWORK`
> Default: empty

Comma-separated list of Docker networks that are attached to every pipeline step.

Example: `WOODPECKER_NETWORK=network1,network2`
-->

### `WOODPECKER_AGENT_SECRET`

> Default: empty

A shared secret used by server and agents to authenticate communication. A secret can be generated by `openssl rand -hex 32`.

### `WOODPECKER_AGENT_SECRET_FILE`

> Default: empty

Read the value for `WOODPECKER_AGENT_SECRET` from the specified filepath

### `WOODPECKER_DISABLE_USER_AGENT_REGISTRATION`

> Default: false

By default, users can create new agents for their repos they have admin access to.
If an instance admin doesn't want this feature enabled, they can disable the API and hide the Web UI elements.

:::note
You should set this option if you have, for example,
global secrets and don't trust your users to create a rogue agent and pipeline for secret extraction.
:::

### `WOODPECKER_KEEPALIVE_MIN_TIME`

> Default: empty

Server-side enforcement policy on the minimum amount of time a client should wait before sending a keepalive ping.

Example: `WOODPECKER_KEEPALIVE_MIN_TIME=10s`

### `WOODPECKER_DATABASE_DRIVER`

> Default: `sqlite3`

The database driver name. Possible values are `sqlite3`, `mysql` or `postgres`.

### `WOODPECKER_DATABASE_DATASOURCE`

> Default: `woodpecker.sqlite` if not running inside a container, `/var/lib/woodpecker/woodpecker.sqlite` if running inside a container

The database connection string. The default value is the path of the embedded SQLite database file.

Example:

```bash
# MySQL
# https://github.com/go-sql-driver/mysql#dsn-data-source-name
WOODPECKER_DATABASE_DATASOURCE=root:password@tcp(1.2.3.4:3306)/woodpecker?parseTime=true

# PostgreSQL
# https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING
WOODPECKER_DATABASE_DATASOURCE=postgres://root:password@1.2.3.4:5432/woodpecker?sslmode=disable
```

### `WOODPECKER_DATABASE_DATASOURCE_FILE`

> Default: empty

Read the value for `WOODPECKER_DATABASE_DATASOURCE` from the specified filepath

### `WOODPECKER_PROMETHEUS_AUTH_TOKEN`

> Default: empty

Token to secure the Prometheus metrics endpoint.
Must be set to enable the endpoint.

### `WOODPECKER_PROMETHEUS_AUTH_TOKEN_FILE`

> Default: empty

Read the value for `WOODPECKER_PROMETHEUS_AUTH_TOKEN` from the specified filepath

### `WOODPECKER_STATUS_CONTEXT`

> Default: `ci/woodpecker`

Context prefix Woodpecker will use to publish status messages to SCM. You probably will only need to change it if you run multiple Woodpecker instances for a single repository.

### `WOODPECKER_STATUS_CONTEXT_FORMAT`

> Default: `{{ .context }}/{{ .event }}/{{ .workflow }}{{if not (eq .axis_id 0)}}/{{.axis_id}}{{end}}`

Template for the status messages published to forges, uses [Go templates](https://pkg.go.dev/text/template) as template language.
Supported variables:

- `context`: Woodpecker's context (see `WOODPECKER_STATUS_CONTEXT`)
- `event`: the event which started the pipeline
- `workflow`: the workflow's name
- `owner`: the repo's owner
- `repo`: the repo's name

---

### `WOODPECKER_CONFIG_SERVICE_ENDPOINT`

> Default: empty

Specify a configuration service endpoint, see [Configuration Extension](../40-advanced/100-external-configuration-api.md)

### `WOODPECKER_FORGE_TIMEOUT`

> Default: 5s

Specify timeout when fetching the Woodpecker configuration from forge. See <https://pkg.go.dev/time#ParseDuration> for syntax reference.

### `WOODPECKER_FORGE_RETRY`

> Default: 3

Specify how many retries of fetching the Woodpecker configuration from a forge are done before we fail.

### `WOODPECKER_ENABLE_SWAGGER`

> Default: true

Enable the Swagger UI for API documentation.

### `WOODPECKER_DISABLE_VERSION_CHECK`

> Default: false

Disable version check in admin web UI.

### `WOODPECKER_LOG_STORE`

> Default: `database`

Where to store logs. Possible values: `database` or `file`.

### `WOODPECKER_LOG_STORE_FILE_PATH`

> Default empty

Directory to store logs in if [`WOODPECKER_LOG_STORE`](#woodpecker_log_store) is `file`.

## Agent

## Forge

### `WOODPECKER_GITHUB_...`

See [GitHub configuration](../11-forges/20-github.md#configuration)

### `WOODPECKER_GITEA_...`

See [Gitea configuration](../11-forges/30-gitea.md#configuration)

### `WOODPECKER_BITBUCKET_...`

See [Bitbucket configuration](../11-forges/50-bitbucket.md#configuration)

### `WOODPECKER_GITLAB_...`

See [GitLab configuration](../11-forges/40-gitlab.md#configuration)

### `WOODPECKER_ADDON_FORGE`

See [addon forges](../11-forges/100-addon.md).

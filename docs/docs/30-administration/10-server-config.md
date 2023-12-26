# Server configuration

## User registration

Woodpecker does not have its own user registry; users are provided from your [forge](./11-forges/10-overview.md) (using OAuth2).

Registration is closed by default (`WOODPECKER_OPEN=false`). If registration is open (`WOODPECKER_OPEN=true`) then every user with an account at the configured forge can login to Woodpecker.

To open registration:

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_OPEN=true
```

You can **also restrict** registration, by keep registration closed and ...\
... **adding** new **users manually** via the CLI: `woodpecker-cli user add`, or\
... allowing specific **admin users** via the `WOODPECKER_ADMIN` setting, or\
by open registration and **filter by organization** membership through the `WOODPECKER_ORGS` setting.

### To close registration, but allow specific admin users

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_OPEN=false
+     - WOODPECKER_ADMIN=johnsmith,janedoe
```

### To only allow registration of users, who are members of approved organizations

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_OPEN=true
+     - WOODPECKER_ORGS=dolores,dogpatch
```

## Administrators

Administrators should also be enumerated in your configuration.

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_ADMIN=johnsmith,janedoe
```

## Filtering repositories

Woodpecker operates with the user's OAuth permission. Due to the coarse permission handling of GitHub, you may end up syncing more repos into Woodpecker than preferred.

Use the `WOODPECKER_REPO_OWNERS` variable to filter which GitHub user's repos should be synced only. You typically want to put here your company's GitHub name.

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_REPO_OWNERS=mycompany,mycompanyossgithubuser
```

## Global registry setting

If you want to make available a specific private registry to all pipelines, use the `WOODPECKER_DOCKER_CONFIG` server configuration.
Point it to your server's docker config.

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    volumes:
      - [...]
+     - /home/user/.docker/config.json:/root/.docker/config.json:ro
    environment:
      - [...]
+     - WOODPECKER_DOCKER_CONFIG=/root/.docker/config.json
```

## Handling sensitive data in docker-compose and docker-swarm

To handle sensitive data in docker-compose or docker-swarm configurations there are several options:

For docker-compose you can use a .env file next to your compose configuration to store the secrets outside of the compose file. While this separates configuration from secrets it is still not very secure.

Alternatively use docker-secrets. As it may be difficult to use docker secrets for environment variables woodpecker allows to read sensible data from files by providing a `*_FILE` option of all sensible configuration variables. Woodpecker will try to read the value directly from this file. Keep in mind that when the original environment variable gets specified at the same time it will override the value read from the file.

```diff title="docker-compose.yml"
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_AGENT_SECRET_FILE=/run/secrets/woodpecker-agent-secret
+   secrets:
+     - woodpecker-agent-secret
+
+secrets:
+  woodpecker-agent-secret:
+    external: true
```

Store a value to a docker secret like this:
`echo "my_agent_secret_key" | docker secret create woodpecker-agent-secret -`

or generate a random one like this:

`openssl rand -hex 32 | docker secret create woodpecker-agent-secret -`

## Custom Javascript and CSS Styling (a.k.a. white-labeling)

Woodpecker supports custom styling of the Web UI by providing custom JS and CSS files.
These files must be present in the server's filesystem.
They can be backed in a Docker image or mounted from a ConfigMap inside a Kubernetes environment.
The configuration variables are independent of each other, which means it can be just one file present, or both.

```text
WOODPECKER_CUSTOM_CSS_FILE=/usr/local/www/woodpecker.css
WOODPECKER_CUSTOM_CSS_FILE=/usr/local/www/woodpecker.js
```

The examples below show how to place a banner message in the top navigation bar of Woodpecker.

### woodpecker.css

```css
.banner-message {
  position: absolute;
  width: 280px;
  height: 40px;
  margin-left: 240px;
  margin-top: 5px;
  padding-top: 5px;
  font-weight: bold;
  background: red no-repeat;
  text-align: center;
}
```

### woodpecker.js

```javascript
// place/copy a minified version of jQuery or ZeptoJS here ...
!(function () {
  'use strict';
  function e() {} /*...*/
})();

$().ready(function () {
  $('.app nav img').first().htmlAfter("<div class='banner-message'>This is a demo banner message :)</div>");
});
```

## All server configuration options

The following list describes all available server configuration options.

### `WOODPECKER_LOG_LEVEL`

> Default: empty

Configures the logging level. Possible values are `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled` and empty.

### `WOODPECKER_LOG_FILE`

> Default: stderr

Output destination for logs.
'stdout' and 'stderr' can be used as special keywords.

### `WOODPECKER_LOG_XORM`

> Default: `false`

Enable XORM logs.

### `WOODPECKER_LOG_XORM_SQL`

> Default: `false`

Enable XORM SQL command logs.

### `WOODPECKER_DEBUG_PRETTY`

> Default: `false`

Enable pretty-printed debug output.

### `WOODPECKER_DEBUG_NOCOLOR`

> Default: `true`

Disable colored debug output.

### `WOODPECKER_HOST`

> Default: empty

Server fully qualified URL of the user-facing hostname and path prefix.

Example: `WOODPECKER_HOST=http://woodpecker.example.org` or `WOODPECKER_HOST=http://example.org/woodpecker`

### `WOODPECKER_WEBHOOK_HOST`

> Default: value from `WOODPECKER_HOST` config env

Server fully qualified URL of the Webhook-facing hostname and path prefix.

Example: `WOODPECKER_WEBHOOK_HOST=http://woodpecker-server.cicd.svc.cluster.local:8000`

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

### `WOODPECKER_LETS_ENCRYPT`

> Default: `false`

Automatically generates an SSL certificate using Let's Encrypt, and configures the server to accept HTTPS requests.

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

Comma-separated list of syncable repo owners. ???

Example: `user1,user2`

### `WOODPECKER_OPEN`

> Default: `false`

Enable to allow user registration.

### `WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`

> Default: `false`

Always use authentication to clone repositories even if they are public. Needed if the forge requires to always authenticate as used by many companies.

### `WOODPECKER_DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS`

> Default: `pull_request, push`

List of event names that will be canceled when a new pipeline for the same context (tag, branch) is created.

### `WOODPECKER_DEFAULT_CLONE_IMAGE`

> Default is defined in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/main/shared/constant/constant.go)

The default docker image to be used when cloning the repo

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

### `WOODPECKER_ESCALATE`

> Defaults are defined in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/main/shared/constant/constant.go)

Docker images to run in privileged mode. Only change if you are sure what you do!

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

### `WOODPECKER_KEEPALIVE_MIN_TIME`

> Default: empty

Server-side enforcement policy on the minimum amount of time a client should wait before sending a keepalive ping.

Example: `WOODPECKER_KEEPALIVE_MIN_TIME=10s`

### `WOODPECKER_DATABASE_DRIVER`

> Default: `sqlite3`

The database driver name. Possible values are `sqlite3`, `mysql` or `postgres`.

### `WOODPECKER_DATABASE_DATASOURCE`

> Default: `woodpecker.sqlite`

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

### `WOODPECKER_ENCRYPTION_KEY`

> Default: empty

Encryption key used to encrypt secrets in DB. See [secrets encryption](./40-encryption.md)

### `WOODPECKER_ENCRYPTION_KEY_FILE`

> Default: empty

Read the value for `WOODPECKER_ENCRYPTION_KEY` from the specified filepath

### `WOODPECKER_ENCRYPTION_TINK_KEYSET_FILE`

> Default: empty

Filepath to encryption keyset used to encrypt secrets in DB. See [secrets encryption](./40-encryption.md)

### `WOODPECKER_ENCRYPTION_DISABLE`

> Default: empty

Boolean flag to decrypt secrets in DB and disable server encryption. See [secrets encryption](./40-encryption.md)

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

### WOODPECKER_ADDONS

> Default: empty

List of addon files. See [addons](./75-addons/00-overview.md).

---

### `WOODPECKER_LIMIT_MEM_SWAP`

> Default: `0`

The maximum amount of memory a single pipeline container is allowed to swap to disk, configured in bytes. There is no limit if `0`.

### `WOODPECKER_LIMIT_MEM`

> Default: `0`

The maximum amount of memory a single pipeline container can use, configured in bytes. There is no limit if `0`.

### `WOODPECKER_LIMIT_SHM_SIZE`

> Default: `0`

The maximum amount of memory of `/dev/shm` allowed in bytes. There is no limit if `0`.

### `WOODPECKER_LIMIT_CPU_QUOTA`

> Default: `0`

The number of microseconds per CPU period that the container is limited to before throttled. There is no limit if `0`.

### `WOODPECKER_LIMIT_CPU_SHARES`

> Default: `0`

The relative weight vs. other containers.

### `WOODPECKER_LIMIT_CPU_SET`

> Default: empty

Comma-separated list to limit the specific CPUs or cores a pipeline container can use.

Example: `WOODPECKER_LIMIT_CPU_SET=1,2`

### `WOODPECKER_CONFIG_SERVICE_ENDPOINT`

> Default: empty

Specify a configuration service endpoint, see [Configuration Extension](./100-external-configuration-api.md)

### `WOODPECKER_FORGE_TIMEOUT`

> Default: 3sec

Specify how many seconds before timeout when fetching the Woodpecker configuration from a Forge

### `WOODPECKER_ENABLE_SWAGGER`

> Default: true

Enable the Swagger UI for API documentation.

---

### `WOODPECKER_GITHUB_...`

See [GitHub configuration](forges/github/#configuration)

### `WOODPECKER_GITEA_...`

See [Gitea configuration](forges/gitea/#configuration)

### `WOODPECKER_BITBUCKET_...`

See [Bitbucket configuration](forges/bitbucket/#configuration)

### `WOODPECKER_GITLAB_...`

See [Gitlab configuration](forges/gitlab/#configuration)

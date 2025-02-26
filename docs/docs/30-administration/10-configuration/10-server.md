---
toc_max_heading_level: 3
---

# Server

## Forge and User configuration

Woodpecker does not have its own user registration. Users are provided by your [forge](./12-forges/11-overview.md) (using OAuth2). The registration is closed by default (`WOODPECKER_OPEN=false`). If the registration is open, any user with an account can log in to Woodpecker with the configured forge.

You can also restrict the registration:

- closed registration and manually managing users with the CLI `woodpecker-cli user`
- open registration and allowing certain admin users with the setting `WOODPECKER_ADMIN`

  ```ini
  WOODPECKER_OPEN=false
  WOODPECKER_ADMIN=john.smith,jane_doe
  ```

- open registration and filtering by organizational affiliation with the setting `WOODPECKER_ORGS`

  ```ini
  WOODPECKER_OPEN=true
  WOODPECKER_ORGS=dolores,dog-patch
  ```

Administrators should also be explicitly set in your configuration.

```ini
WOODPECKER_ADMIN=john.smith,jane_doe
```

## Repository configuration

Woodpecker works with the user's OAuth permissions on the forge. By default Woodpecker will synchronize all repositories the user has access to. Use the variable `WOODPECKER_REPO_OWNERS` to filter which repos should only be synchronized by GitHub users. Normally you should enter the GitHub name of your company here.

```ini
WOODPECKER_REPO_OWNERS=my_company,my_company_oss_github_user
```

## Databases

The default database engine of Woodpecker is an embedded SQLite database which requires zero installation or configuration. But you can replace it with a MySQL/MariaDB or PostgreSQL database. There are also some fundamentals to keep in mind:

- Woodpecker does not create your database automatically. If you are using the MySQL or Postgres driver you will need to manually create your database using `CREATE DATABASE`.

- Woodpecker does not perform data archival; it considered out-of-scope for the project. Woodpecker is rather conservative with the amount of data it stores, however, you should expect the database logs to grow the size of your database considerably.

- Woodpecker automatically handles database migration, including the initial creation of tables and indexes. New versions of Woodpecker will automatically upgrade the database unless otherwise specified in the release notes.

- Woodpecker does not perform database backups. This should be handled by separate third party tools provided by your database vendor of choice.

### SQLite

By default Woodpecker uses a SQLite database stored under `/var/lib/woodpecker/`. If using containers, you can mount a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the SQLite database.

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
+    volumes:
+      - woodpecker-server-data:/var/lib/woodpecker/
```

### MySQL/MariaDB

The below example demonstrates MySQL database configuration. See the official driver [documentation](https://github.com/go-sql-driver/mysql#dsn-data-source-name) for configuration options and examples.
The minimum version of MySQL/MariaDB required is determined by the `go-sql-driver/mysql` - see [it's README](https://github.com/go-sql-driver/mysql#requirements) for more information.

```ini
WOODPECKER_DATABASE_DRIVER=mysql
WOODPECKER_DATABASE_DATASOURCE=root:password@tcp(1.2.3.4:3306)/woodpecker?parseTime=true
```

### PostgreSQL

The below example demonstrates Postgres database configuration. See the official driver [documentation](https://www.postgresql.org/docs/current/static/libpq-connect.html#LIBPQ-CONNSTRING) for configuration options and examples.
Please use Postgres versions equal or higher than **11**.

```ini
WOODPECKER_DATABASE_DRIVER=postgres
WOODPECKER_DATABASE_DATASOURCE=postgres://root:password@1.2.3.4:5432/postgres?sslmode=disable
```

## External Configuration API

To provide additional management and preprocessing capabilities for pipeline configurations Woodpecker supports an HTTP API which can be enabled to call an external config service.
Before the run or restart of any pipeline Woodpecker will make a POST request to an external HTTP API sending the current repository, build information and all current config files retrieved from the repository. The external API can then send back new pipeline configurations that will be used immediately or respond with `HTTP 204` to tell the system to use the existing configuration.

Every request sent by Woodpecker is signed using a [http-signature](https://datatracker.ietf.org/doc/html/rfc9421) by a private key (ed25519) generated on the first start of the Woodpecker server. You can get the public key for the verification of the http-signature from `http(s)://your-woodpecker-server/api/signature/public-key`.

A simplistic example configuration service can be found here: [https://github.com/woodpecker-ci/example-config-service](https://github.com/woodpecker-ci/example-config-service)

:::warning
You need to trust the external config service as it is getting secret information about the repository and pipeline and has the ability to change pipeline configs that could run malicious tasks.
:::

### Configuration

```ini title="Server"
WOODPECKER_CONFIG_SERVICE_ENDPOINT=https://example.com/ciconfig
```

#### Example request made by Woodpecker

```json
{
  "repo": {
    "id": 100,
    "uid": "",
    "user_id": 0,
    "namespace": "",
    "name": "woodpecker-test-pipe",
    "slug": "",
    "scm": "git",
    "git_http_url": "",
    "git_ssh_url": "",
    "link": "",
    "default_branch": "",
    "private": true,
    "visibility": "private",
    "active": true,
    "config": "",
    "trusted": false,
    "protected": false,
    "ignore_forks": false,
    "ignore_pulls": false,
    "cancel_pulls": false,
    "timeout": 60,
    "counter": 0,
    "synced": 0,
    "created": 0,
    "updated": 0,
    "version": 0
  },
  "pipeline": {
    "author": "myUser",
    "author_avatar": "https://myforge.com/avatars/d6b3f7787a685fcdf2a44e2c685c7e03",
    "author_email": "my@email.com",
    "branch": "main",
    "changed_files": ["some-file-name.txt"],
    "commit": "2fff90f8d288a4640e90f05049fe30e61a14fd50",
    "created_at": 0,
    "deploy_to": "",
    "enqueued_at": 0,
    "error": "",
    "event": "push",
    "finished_at": 0,
    "id": 0,
    "link_url": "https://myforge.com/myUser/woodpecker-testpipe/commit/2fff90f8d288a4640e90f05049fe30e61a14fd50",
    "message": "test old config\n",
    "number": 0,
    "parent": 0,
    "ref": "refs/heads/main",
    "refspec": "",
    "clone_url": "",
    "reviewed_at": 0,
    "reviewed_by": "",
    "sender": "myUser",
    "signed": false,
    "started_at": 0,
    "status": "",
    "timestamp": 1645962783,
    "title": "",
    "updated_at": 0,
    "verified": false
  },
  "netrc": {
    "machine": "https://example.com",
    "login": "user",
    "password": "password"
  }
}
```

#### Example response structure

```json
{
  "configs": [
    {
      "name": "central-override",
      "data": "steps:\n  - name: backend\n    image: alpine\n    commands:\n      - echo \"Hello there from ConfigAPI\"\n"
    }
  ]
}
```

## TLS

Woodpecker supports SSL configuration by mounting certificates into your container.

```ini
WOODPECKER_SERVER_CERT=/etc/certs/woodpecker.example.com/server.crt
WOODPECKER_SERVER_KEY=/etc/certs/woodpecker.example.com/server.key
```

TLS support is provided using the [ListenAndServeTLS](https://golang.org/pkg/net/http/#ListenAndServeTLS) function from the Go standard library.

### Container configuration

In addition to the ports shown in the [docker-compose](../05-installation/10-docker-compose.md) installation, port `443` must be exposed:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     ports:
+      - 80:80
+      - 443:443
       - 9000:9000
```

Additionally, the certificate and key must be mounted and referenced:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
+      - WOODPECKER_SERVER_CERT=/etc/certs/woodpecker.example.com/server.crt
+      - WOODPECKER_SERVER_KEY=/etc/certs/woodpecker.example.com/server.key
     volumes:
+      - /etc/certs/woodpecker.example.com/server.crt:/etc/certs/woodpecker.example.com/server.crt
+      - /etc/certs/woodpecker.example.com/server.key:/etc/certs/woodpecker.example.com/server.key
```

## Metrics

### Endpoint

Woodpecker is compatible with Prometheus and exposes a `/metrics` endpoint if the environment variable `WOODPECKER_PROMETHEUS_AUTH_TOKEN` is set. Please note that access to the metrics endpoint is restricted and requires the authorization token from the environment variable mentioned above.

```yaml
global:
  scrape_interval: 60s

scrape_configs:
  - job_name: 'woodpecker'
    bearer_token: dummyToken...

    static_configs:
      - targets: ['woodpecker.domain.com']
```

### Authorization

An administrator will need to generate a user API token and configure in the Prometheus configuration file as a bearer token. Please see the following example:

```diff
 global:
   scrape_interval: 60s

 scrape_configs:
   - job_name: 'woodpecker'
+    bearer_token: dummyToken...

     static_configs:
        - targets: ['woodpecker.domain.com']
```

As an alternative, the token can also be read from a file:

```diff
 global:
   scrape_interval: 60s

 scrape_configs:
   - job_name: 'woodpecker'
+    bearer_token_file: /etc/secrets/woodpecker-monitoring-token

     static_configs:
        - targets: ['woodpecker.domain.com']
```

### Reference

List of Prometheus metrics specific to Woodpecker:

```yaml
# HELP woodpecker_pipeline_count Pipeline count.
# TYPE woodpecker_pipeline_count counter
woodpecker_pipeline_count{branch="main",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 3
woodpecker_pipeline_count{branch="dev",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 3
# HELP woodpecker_pipeline_time Build time.
# TYPE woodpecker_pipeline_time gauge
woodpecker_pipeline_time{branch="main",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 116
woodpecker_pipeline_time{branch="dev",pipeline="total",repo="woodpecker-ci/woodpecker",status="success"} 155
# HELP woodpecker_pipeline_total_count Total number of builds.
# TYPE woodpecker_pipeline_total_count gauge
woodpecker_pipeline_total_count 1025
# HELP woodpecker_pending_steps Total number of pending pipeline steps.
# TYPE woodpecker_pending_steps gauge
woodpecker_pending_steps 0
# HELP woodpecker_repo_count Total number of repos.
# TYPE woodpecker_repo_count gauge
woodpecker_repo_count 9
# HELP woodpecker_running_steps Total number of running pipeline steps.
# TYPE woodpecker_running_steps gauge
woodpecker_running_steps 0
# HELP woodpecker_user_count Total number of users.
# TYPE woodpecker_user_count gauge
woodpecker_user_count 1
# HELP woodpecker_waiting_steps Total number of pipeline waiting on deps.
# TYPE woodpecker_waiting_steps gauge
woodpecker_waiting_steps 0
# HELP woodpecker_worker_count Total number of workers.
# TYPE woodpecker_worker_count gauge
woodpecker_worker_count 4
```

## UI customization

Woodpecker supports custom JS and CSS files. These files must be present in the server's filesystem.
They can be backed in a Docker image or mounted from a ConfigMap inside a Kubernetes environment.
The configuration variables are independent of each other, which means it can be just one file present, or both.

```ini
WOODPECKER_CUSTOM_CSS_FILE=/usr/local/www/woodpecker.css
WOODPECKER_CUSTOM_JS_FILE=/usr/local/www/woodpecker.js
```

The examples below show how to place a banner message in the top navigation bar of Woodpecker.

```css title="woodpecker.css"
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

```javascript title="woodpecker.js"
// place/copy a minified version of your preferred lightweight JavaScript library here ...
!(function () {
  'use strict';
  function e() {} /*...*/
})();

$().ready(function () {
  $('.app nav img').first().htmlAfter("<div class='banner-message'>This is a demo banner message :)</div>");
});
```

## Environment variables

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

Specify a configuration service endpoint, see [Configuration Extension](#external-configuration-api)

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

---

### `WOODPECKER_GITHUB_...`

See [GitHub configuration](./12-forges/20-github.md#configuration)

### `WOODPECKER_GITEA_...`

See [Gitea configuration](./12-forges/30-gitea.md#configuration)

### `WOODPECKER_BITBUCKET_...`

See [Bitbucket configuration](./12-forges/50-bitbucket.md#configuration)

### `WOODPECKER_GITLAB_...`

See [GitLab configuration](./12-forges/40-gitlab.md#configuration)

### `WOODPECKER_ADDON_FORGE`

See [addon forges](./12-forges/100-addon.md).

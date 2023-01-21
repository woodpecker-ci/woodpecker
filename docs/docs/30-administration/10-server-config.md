# Server configuration

## User registration

Registration is closed by default. While disabled an administrator needs to add new users manually (exp. `woodpecker-cli user add`).

If registration is open every user with an account at the configured [forges](./11-forges/10-overview.md) can login to Woodpecker.
This example enables open registration for users that are members of approved organizations:

```diff
# docker-compose.yml
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

```diff
# docker-compose.yml
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

```diff
# docker-compose.yml
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

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_DOCKER_CONFIG=/home/user/.docker/config.json
```

## Handling sensitive data in docker-compose and docker-swarm

To handle sensitive data in docker-compose or docker-swarm configurations there are several options:

For docker-compose you can use a .env file next to your compose configuration to store the secrets outside of the compose file. While this separates configuration from secrets it is still not very secure.

Alternatively use docker-secrets. As it may be difficult to use docker secrets for environment variables woodpecker allows to read sensible data from files by providing a `*_FILE` option of all sensible configuration variables. Woodpecker will try to read the value directly from this file. Keep in mind that when the original environment variable gets specified at the same time it will override the value read from the file.

```diff
# docker-compose.yml
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

## All server configuration options

The following list describes all available server configuration options.

### `WOODPECKER_LOG_LEVEL`
> Default: empty

Configures the logging level. Possible values are `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled` and empty.

### `WOODPECKER_DEBUG_PRETTY`
> Default: `false`

Enable pretty-printed debug output.

### `WOODPECKER_DEBUG_NOCOLOR`
> Default: `true`

Disable colored debug output.

### `WOODPECKER_HOST`
> Default: empty

Server fully qualified URL of the user-facing hostname.

Example: `WOODPECKER_HOST=http://woodpecker.example.org`

### `WOODPECKER_SERVER_ADDR`
> Default: `:8000`

Configures the HTTP listener port.

### `WOODPECKER_SERVER_CERT`
> Default: empty

Path to an SSL certificate used by the server to accept HTTPS requests.

Example: `WOODPECKER_SERVER_CERT=/path/to/cert.pem`

### `WOODPECKER_SERVER_KEY`
> Default: empty

Path to an SSL certificate key used by the server to accept HTTPS requests.

Example: `WOODPECKER_SERVER_KEY=/path/to/key.pem`

### `WOODPECKER_LETS_ENCRYPT`
> Default: `false`

Automatically generates an SSL certificate using Let's Encrypt, and configures the server to accept HTTPS requests.

### `WOODPECKER_GRPC_ADDR`
> Default: `:9000`

Configures the gRPC listener port.


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

### `WOODPECKER_DOCS`
> Default: `https://woodpecker-ci.org/`

Link to documentation in the UI.

### `WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`
> Default: `false`

Always use authentication to clone repositories even if they are public. Needed if the forge requires to always authenticate as used by many companies.

### `WOODPECKER_DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS`
> Default: `pull_request, push`

List of event names that will be canceled when a new pipeline for the same context (tag, branch) is created.

### `WOODPECKER_DEFAULT_CLONE_IMAGE`
> Default is defined in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/master/shared/constant/constant.go)

The default docker image to be used when cloning the repo

### `WOODPECKER_SESSION_EXPIRES`
> Default: `72h`

Configures the session expiration time.

### `WOODPECKER_ESCALATE`
> Defaults are defined in [shared/constant/constant.go](https://github.com/woodpecker-ci/woodpecker/blob/master/shared/constant/constant.go)

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
> Default: `{{ .context }}/{{ .event }}/{{ .pipeline }}`

Template for the status messages published to forges, uses [Go templates](https://pkg.go.dev/text/template) as template language.
Supported variables:
- `context`: Woodpecker's context (see `WOODPECKER_STATUS_CONTEXT`)
- `event`: the event which started the pipeline
- `pipeline`: the pipeline's name
- `owner`: the repo's owner
- `repo`: the repo's name

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
> Default: ``

Specify a configuration service endpoint, see [Configuration Extension](./100-external-configuration-api.md)

---

### `WOODPECKER_GITHUB_...`

See [GitHub configuration](forges/github/#configuration)

### `WOODPECKER_GOGS_...`

See [Gogs configuration](forges/gogs/#configuration)

### `WOODPECKER_FORGEJO_...`

See [Forgejo configuration](forges/forgejo/#configuration)

### `WOODPECKER_GITEA_...`

See [Gitea configuration](forges/gitea/#configuration)

### `WOODPECKER_BITBUCKET_...`

See [Bitbucket configuration](forges/bitbucket/#configuration)

### `WOODPECKER_STASH_...`

See [Bitbucket server configuration](forges/bitbucket_server/#configuration)

### `WOODPECKER_GITLAB_...`

See [Gitlab configuration](forges/gitlab/#configuration)

### `WOODPECKER_CODING_...`

See [Coding configuration](forges/coding/#configuration)

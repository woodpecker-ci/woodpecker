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

## Reverse Proxy

### Apache

This guide provides a brief overview for installing Woodpecker server behind the Apache2 web-server. This is an example configuration:

<!-- cspell:ignore apacheconf -->

```apacheconf
ProxyPreserveHost On

RequestHeader set X-Forwarded-Proto "https"

ProxyPass / http://127.0.0.1:8000/
ProxyPassReverse / http://127.0.0.1:8000/
```

You must have these Apache modules installed:

- `proxy`
- `proxy_http`

You must configure Apache to set `X-Forwarded-Proto` when using https.

```diff
 ProxyPreserveHost On

+RequestHeader set X-Forwarded-Proto "https"

 ProxyPass / http://127.0.0.1:8000/
 ProxyPassReverse / http://127.0.0.1:8000/
```

### Nginx

This guide provides a basic overview for installing Woodpecker server behind the Nginx web-server. For more advanced configuration options please consult the official Nginx [documentation](https://docs.nginx.com/nginx/admin-guide).

Example configuration:

```nginx
server {
    listen 80;
    server_name woodpecker.example.com;

    location / {
        proxy_set_header X-Forwarded-For $remote_addr;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header Host $http_host;

        proxy_pass http://127.0.0.1:8000;
        proxy_redirect off;
        proxy_http_version 1.1;
        proxy_buffering off;

        chunked_transfer_encoding off;
    }
}
```

You must configure the proxy to set `X-Forwarded` proxy headers:

```diff
 server {
     listen 80;
     server_name woodpecker.example.com;

     location / {
+        proxy_set_header X-Forwarded-For $remote_addr;
+        proxy_set_header X-Forwarded-Proto $scheme;

         proxy_pass http://127.0.0.1:8000;
         proxy_redirect off;
         proxy_http_version 1.1;
         proxy_buffering off;

         chunked_transfer_encoding off;
     }
 }
```

### Caddy

This guide provides a brief overview for installing Woodpecker server behind the [Caddy web-server](https://caddyserver.com/). This is an example caddyfile proxy configuration:

```caddy
# expose WebUI and API
woodpecker.example.com {
  reverse_proxy woodpecker-server:8000
}

# expose gRPC
woodpecker-agent.example.com {
  reverse_proxy h2c://woodpecker-server:9000
}
```

### Tunnelmole

[Tunnelmole](https://github.com/robbie-cahill/tunnelmole-client) is an open source tunneling tool.

Start by [installing tunnelmole](https://github.com/robbie-cahill/tunnelmole-client#installation).

After the installation, run the following command to start tunnelmole:

```bash
tmole 8000
```

It will start a tunnel and will give a response like this:

```bash
➜  ~ tmole 8000
http://bvdo5f-ip-49-183-170-144.tunnelmole.net is forwarding to localhost:8000
https://bvdo5f-ip-49-183-170-144.tunnelmole.net is forwarding to localhost:8000
```

Set `WOODPECKER_HOST` to the Tunnelmole URL (`xxx.tunnelmole.net`) and start the server.

### Ngrok

[Ngrok](https://ngrok.com/) is a popular closed source tunnelling tool. After installing ngrok, open a new console and run the following command:

```bash
ngrok http 8000
```

Set `WOODPECKER_HOST` to the ngrok URL (usually xxx.ngrok.io) and start the server.

### Traefik

To install the Woodpecker server behind a [Traefik](https://traefik.io/) load balancer, you must expose both the `http` and the `gRPC` ports. Here is a comprehensive example, considering you are running Traefik with docker swarm and want to do TLS termination and automatic redirection from http to https.

<!-- cspell:words redirectscheme certresolver  -->

```yaml
services:
  server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_ADMIN=your_admin_user
      # other settings ...

    networks:
      - dmz # externally defined network, so that traefik can connect to the server
    volumes:
      - woodpecker-server-data:/var/lib/woodpecker/

    deploy:
      labels:
        - traefik.enable=true

        # web server
        - traefik.http.services.woodpecker-service.loadbalancer.server.port=8000

        - traefik.http.routers.woodpecker-secure.rule=Host(`ci.example.com`)
        - traefik.http.routers.woodpecker-secure.tls=true
        - traefik.http.routers.woodpecker-secure.tls.certresolver=letsencrypt
        - traefik.http.routers.woodpecker-secure.entrypoints=web-secure
        - traefik.http.routers.woodpecker-secure.service=woodpecker-service

        - traefik.http.routers.woodpecker.rule=Host(`ci.example.com`)
        - traefik.http.routers.woodpecker.entrypoints=web
        - traefik.http.routers.woodpecker.service=woodpecker-service

        - traefik.http.middlewares.woodpecker-redirect.redirectscheme.scheme=https
        - traefik.http.middlewares.woodpecker-redirect.redirectscheme.permanent=true
        - traefik.http.routers.woodpecker.middlewares=woodpecker-redirect@docker

        #  gRPC service
        - traefik.http.services.woodpecker-grpc.loadbalancer.server.port=9000
        - traefik.http.services.woodpecker-grpc.loadbalancer.server.scheme=h2c

        - traefik.http.routers.woodpecker-grpc-secure.rule=Host(`woodpecker-grpc.example.com`)
        - traefik.http.routers.woodpecker-grpc-secure.tls=true
        - traefik.http.routers.woodpecker-grpc-secure.tls.certresolver=letsencrypt
        - traefik.http.routers.woodpecker-grpc-secure.entrypoints=web-secure
        - traefik.http.routers.woodpecker-grpc-secure.service=woodpecker-grpc

        - traefik.http.routers.woodpecker-grpc.rule=Host(`woodpecker-grpc.example.com`)
        - traefik.http.routers.woodpecker-grpc.entrypoints=web
        - traefik.http.routers.woodpecker-grpc.service=woodpecker-grpc

        - traefik.http.middlewares.woodpecker-grpc-redirect.redirectscheme.scheme=https
        - traefik.http.middlewares.woodpecker-grpc-redirect.redirectscheme.permanent=true
        - traefik.http.routers.woodpecker-grpc.middlewares=woodpecker-grpc-redirect@docker

volumes:
  woodpecker-server-data:
    driver: local

networks:
  dmz:
    external: true
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

### LOG_LEVEL

- Name: `WOODPECKER_LOG_LEVEL`
- Default: `info`

Configures the logging level. Possible values are `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`, `disabled` and empty.

---

### LOG_FILE

- Name: `WOODPECKER_LOG_FILE`
- Default: `stderr`

Output destination for logs.
'stdout' and 'stderr' can be used as special keywords.

---

### DATABASE_LOG

- Name: `WOODPECKER_DATABASE_LOG`
- Default: `false`

Enable logging in database engine (currently xorm).

---

### DATABASE_LOG_SQL

- Name: `WOODPECKER_DATABASE_LOG_SQL`
- Default: `false`

Enable logging of sql commands.

---

### DATABASE_MAX_CONNECTIONS

- Name: `WOODPECKER_DATABASE_MAX_CONNECTIONS`
- Default: `100`

Max database connections xorm is allowed create.

---

### DATABASE_IDLE_CONNECTIONS

- Name: `WOODPECKER_DATABASE_IDLE_CONNECTIONS`
- Default: `2`

Amount of database connections xorm will hold open.

---

### DATABASE_CONNECTION_TIMEOUT

- Name: `WOODPECKER_DATABASE_CONNECTION_TIMEOUT`
- Default: `3 Seconds`

Time an active database connection is allowed to stay open.

---

### DEBUG_PRETTY

- Name: `WOODPECKER_DEBUG_PRETTY`
- Default: `false`

Enable pretty-printed debug output.

---

### DEBUG_NOCOLOR

- Name: `WOODPECKER_DEBUG_NOCOLOR`
- Default: `true`

Disable colored debug output.

---

### HOST

- Name: `WOODPECKER_HOST`
- Default: none

Server fully qualified URL of the user-facing hostname, port (if not default for HTTP/HTTPS) and path prefix.

Examples:

- `WOODPECKER_HOST=http://woodpecker.example.org`
- `WOODPECKER_HOST=http://example.org/woodpecker`
- `WOODPECKER_HOST=http://example.org:1234/woodpecker`

---

### SERVER_ADDR

- Name: `WOODPECKER_SERVER_ADDR`
- Default: `:8000`

Configures the HTTP listener port.

---

### SERVER_ADDR_TLS

- Name: `WOODPECKER_SERVER_ADDR_TLS`
- Default: `:443`

Configures the HTTPS listener port when SSL is enabled.

---

### SERVER_CERT

- Name: `WOODPECKER_SERVER_CERT`
- Default: none

Path to an SSL certificate used by the server to accept HTTPS requests.

Example: `WOODPECKER_SERVER_CERT=/path/to/cert.pem`

---

### SERVER_KEY

- Name: `WOODPECKER_SERVER_KEY`
- Default: none

Path to an SSL certificate key used by the server to accept HTTPS requests.

Example: `WOODPECKER_SERVER_KEY=/path/to/key.pem`

---

### CUSTOM_CSS_FILE

- Name: `WOODPECKER_CUSTOM_CSS_FILE`
- Default: none

File path for the server to serve a custom .CSS file, used for customizing the UI.
Can be used for showing banner messages, logos, or environment-specific hints (a.k.a. white-labeling).
The file must be UTF-8 encoded, to ensure all special characters are preserved.

Example: `WOODPECKER_CUSTOM_CSS_FILE=/usr/local/www/woodpecker.css`

---

### CUSTOM_JS_FILE

- Name: `WOODPECKER_CUSTOM_JS_FILE`
- Default: none

File path for the server to serve a custom .JS file, used for customizing the UI.
Can be used for showing banner messages, logos, or environment-specific hints (a.k.a. white-labeling).
The file must be UTF-8 encoded, to ensure all special characters are preserved.

Example: `WOODPECKER_CUSTOM_JS_FILE=/usr/local/www/woodpecker.js`

---

### GRPC_ADDR

- Name: `WOODPECKER_GRPC_ADDR`
- Default: `:9000`

Configures the gRPC listener port.

---

### GRPC_SECRET

- Name: `WOODPECKER_GRPC_SECRET`
- Default: `secret`

Configures the gRPC JWT secret.

---

### GRPC_SECRET_FILE

- Name: `WOODPECKER_GRPC_SECRET_FILE`
- Default: none

Read the value for `WOODPECKER_GRPC_SECRET` from the specified filepath.

---

### METRICS_SERVER_ADDR

- Name: `WOODPECKER_METRICS_SERVER_ADDR`
- Default: none

Configures an unprotected metrics endpoint. An empty value disables the metrics endpoint completely.

Example: `:9001`

---

### ADMIN

- Name: `WOODPECKER_ADMIN`
- Default: none

Comma-separated list of admin accounts.

Example: `WOODPECKER_ADMIN=user1,user2`

---

### ORGS

- Name: `WOODPECKER_ORGS`
- Default: none

Comma-separated list of approved organizations.

Example: `org1,org2`

---

### REPO_OWNERS

- Name: `WOODPECKER_REPO_OWNERS`
- Default: none

Repositories by those owners will be allowed to be used in woodpecker.

Example: `user1,user2`

---

### OPEN

- Name: `WOODPECKER_OPEN`
- Default: `false`

Enable to allow user registration.

---

### AUTHENTICATE_PUBLIC_REPOS

- Name: `WOODPECKER_AUTHENTICATE_PUBLIC_REPOS`
- Default: `false`

Always use authentication to clone repositories even if they are public. Needed if the forge requires to always authenticate as used by many companies.

---

### DEFAULT_ALLOW_PULL_REQUESTS

- Name: `WOODPECKER_DEFAULT_ALLOW_PULL_REQUESTS`
- Default: `true`

The default setting for allowing pull requests on a repo.

---

### DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS

- Name: `WOODPECKER_DEFAULT_CANCEL_PREVIOUS_PIPELINE_EVENTS`
- Default: `pull_request, push`

List of event names that will be canceled when a new pipeline for the same context (tag, branch) is created.

---

### DEFAULT_CLONE_PLUGIN

- Name: `WOODPECKER_DEFAULT_CLONE_PLUGIN`
- Default: `docker.io/woodpeckerci/plugin-git`

The default docker image to be used when cloning the repo.

It is also added to the trusted clone plugin list.

### DEFAULT_WORKFLOW_LABELS

- Name: `WOODPECKER_DEFAULT_WORKFLOW_LABELS`
- Default: none

You can specify default label/platform conditions that will be used for agent selection for workflows that does not have labels conditions set.

Example: `platform=linux/amd64,backend=docker`

### DEFAULT_PIPELINE_TIMEOUT

- Name: `WOODPECKER_DEFAULT_PIPELINE_TIMEOUT`
- Default: 60

The default time for a repo in minutes before a pipeline gets killed

### MAX_PIPELINE_TIMEOUT

- Name: `WOODPECKER_MAX_PIPELINE_TIMEOUT`
- Default: 120

The maximum time in minutes you can set in the repo settings before a pipeline gets killed

---

### SESSION_EXPIRES

- Name: `WOODPECKER_SESSION_EXPIRES`
- Default: `72h`

Configures the session expiration time.
Context: when someone does log into Woodpecker, a temporary session token is created.
As long as the session is valid (until it expires or log-out),
a user can log into Woodpecker, without re-authentication.

### PLUGINS_PRIVILEGED

- Name: `WOODPECKER_PLUGINS_PRIVILEGED`
- Default: none

Docker images to run in privileged mode. Only change if you are sure what you do!

You should specify the tag of your images too, as this enforces exact matches.

### PLUGINS_TRUSTED_CLONE

- Name: `WOODPECKER_PLUGINS_TRUSTED_CLONE`
- Default: `docker.io/woodpeckerci/plugin-git,docker.io/woodpeckerci/plugin-git,quay.io/woodpeckerci/plugin-git`

Plugins which are trusted to handle the Git credential info in clone steps.
If a clone step use an image not in this list, Git credentials will not be injected and users have to use other methods (e.g. secrets) to clone non-public repos.

You should specify the tag of your images too, as this enforces exact matches.

<!-- ---

### `VOLUME`

- Name: `WOODPECKER_VOLUME`
- Default: none

Comma-separated list of Docker volumes that are mounted into every pipeline step.

Example: `WOODPECKER_VOLUME=/path/on/host:/path/in/container:rw`| -->

---

### DOCKER_CONFIG

- Name: `WOODPECKER_DOCKER_CONFIG`
- Default: none

Configures a specific private registry config for all pipelines.

Example: `WOODPECKER_DOCKER_CONFIG=/home/user/.docker/config.json`

---

### ENVIRONMENT

- Name: `WOODPECKER_ENVIRONMENT`
- Default: none

If you want specific environment variables to be available in all of your pipelines use the `WOODPECKER_ENVIRONMENT` setting on the Woodpecker server. Note that these can't overwrite any existing, built-in variables.

Example: `WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2`

<!-- ---

### NETWORK

- Name: `WOODPECKER_NETWORK`
- Default: none

Comma-separated list of Docker networks that are attached to every pipeline step.

Example: `WOODPECKER_NETWORK=network1,network2` -->

---

### AGENT_SECRET

- Name: `WOODPECKER_AGENT_SECRET`
- Default: none

A shared secret used by server and agents to authenticate communication. A secret can be generated by `openssl rand -hex 32`.

---

### AGENT_SECRET_FILE

- Name: `WOODPECKER_AGENT_SECRET_FILE`
- Default: none

Read the value for `WOODPECKER_AGENT_SECRET` from the specified filepath

---

### DISABLE_USER_AGENT_REGISTRATION

- Name: `WOODPECKER_DISABLE_USER_AGENT_REGISTRATION`
- Default: false

By default, users can create new agents for their repos they have admin access to.
If an instance admin doesn't want this feature enabled, they can disable the API and hide the Web UI elements.

:::note
You should set this option if you have, for example,
global secrets and don't trust your users to create a rogue agent and pipeline for secret extraction.
:::

---

### KEEPALIVE_MIN_TIME

- Name: `WOODPECKER_KEEPALIVE_MIN_TIME`
- Default: none

Server-side enforcement policy on the minimum amount of time a client should wait before sending a keepalive ping.

Example: `WOODPECKER_KEEPALIVE_MIN_TIME=10s`

---

### DATABASE_DRIVER

- Name: `WOODPECKER_DATABASE_DRIVER`
- Default: `sqlite3`

The database driver name. Possible values are `sqlite3`, `mysql` or `postgres`.

---

### DATABASE_DATASOURCE

- Name: `WOODPECKER_DATABASE_DATASOURCE`
- Default: `woodpecker.sqlite` if not running inside a container, `/var/lib/woodpecker/woodpecker.sqlite` if running inside a container

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

---

### DATABASE_DATASOURCE_FILE

- Name: `WOODPECKER_DATABASE_DATASOURCE_FILE`
- Default: none

Read the value for `WOODPECKER_DATABASE_DATASOURCE` from the specified filepath

---

### PROMETHEUS_AUTH_TOKEN

- Name: `WOODPECKER_PROMETHEUS_AUTH_TOKEN`
- Default: none

Token to secure the Prometheus metrics endpoint.
Must be set to enable the endpoint.

---

### PROMETHEUS_AUTH_TOKEN_FILE

- Name: `WOODPECKER_PROMETHEUS_AUTH_TOKEN_FILE`
- Default: none

Read the value for `WOODPECKER_PROMETHEUS_AUTH_TOKEN` from the specified filepath

---

### STATUS_CONTEXT

- Name: `WOODPECKER_STATUS_CONTEXT`
- Default: `ci/woodpecker`

Context prefix Woodpecker will use to publish status messages to SCM. You probably will only need to change it if you run multiple Woodpecker instances for a single repository.

---

### STATUS_CONTEXT_FORMAT

- Name: `WOODPECKER_STATUS_CONTEXT_FORMAT`
- Default: `{{ .context }}/{{ .event }}/{{ .workflow }}{{if not (eq .axis_id 0)}}/{{.axis_id}}{{end}}`

Template for the status messages published to forges, uses [Go templates](https://pkg.go.dev/text/template) as template language.
Supported variables:

- `context`: Woodpecker's context (see `WOODPECKER_STATUS_CONTEXT`)
- `event`: the event which started the pipeline
- `workflow`: the workflow's name
- `owner`: the repo's owner
- `repo`: the repo's name

---

---

### CONFIG_SERVICE_ENDPOINT

- Name: `WOODPECKER_CONFIG_SERVICE_ENDPOINT`
- Default: none

Specify a configuration service endpoint, see [Configuration Extension](#external-configuration-api)

---

### FORGE_TIMEOUT

- Name: `WOODPECKER_FORGE_TIMEOUT`
- Default: 5s

Specify timeout when fetching the Woodpecker configuration from forge. See <https://pkg.go.dev/time#ParseDuration> for syntax reference.

---

### FORGE_RETRY

- Name: `WOODPECKER_FORGE_RETRY`
- Default: 3

Specify how many retries of fetching the Woodpecker configuration from a forge are done before we fail.

---

### ENABLE_SWAGGER

- Name: `WOODPECKER_ENABLE_SWAGGER`
- Default: true

Enable the Swagger UI for API documentation.

---

### DISABLE_VERSION_CHECK

- Name: `WOODPECKER_DISABLE_VERSION_CHECK`
- Default: false

Disable version check in admin web UI.

---

### LOG_STORE

- Name: `WOODPECKER_LOG_STORE`
- Default: `database`

Where to store logs. Possible values: `database` or `file`.

---

### LOG_STORE_FILE_PATH

- Name: `WOODPECKER_LOG_STORE_FILE_PATH`
- Default: none

Directory to store logs in if [`WOODPECKER_LOG_STORE`](#log_store) is `file`.

---

### GITHUB\_\*

See [GitHub configuration](./12-forges/20-github.md#configuration)

---

### GITEA\_\*

See [Gitea configuration](./12-forges/30-gitea.md#configuration)

---

### BITBUCKET\_\*

See [Bitbucket configuration](./12-forges/50-bitbucket.md#configuration)

---

### GITLAB\_\*

See [GitLab configuration](./12-forges/40-gitlab.md#configuration)

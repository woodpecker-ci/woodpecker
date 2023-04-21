# Gitea

Woodpecker comes with built-in support for Gitea. To enable Gitea you should configure the Woodpecker container using the following environment variables:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_GITEA=true
+     - WOODPECKER_GITEA_URL=${WOODPECKER_GITEA_URL}
+     - WOODPECKER_GITEA_CLIENT=${WOODPECKER_GITEA_CLIENT}
+     - WOODPECKER_GITEA_SECRET=${WOODPECKER_GITEA_SECRET}

  woodpecker-agent:
    [...]
```

## Registration

Register your application with Gitea to create your client id and secret. You can find the OAuth applications settings of Gitea at `https://gitea.<host>/user/settings/`. It is very important the authorization callback URL matches your http(s) scheme and hostname exactly with `https://<host>/authorize` as the path.

If you run the Woodpecker CI server on the same host as the Gitea instance, you might also need to allow local connections in Gitea, since version `v1.16`. Otherwise webhooks will fail. Add the following lines to your Gitea configuration (usually at `/etc/gitea/conf/app.ini`).
```ini
...
[webhook]
ALLOWED_HOST_LIST=external,loopback
```
For reference see [Configuration Cheat Sheet](https://docs.gitea.io/en-us/config-cheat-sheet/#webhook-webhook).

![gitea oauth setup](gitea_oauth.gif)

If you run Woodpecker CI and Gitea behind a reverse proxy that authenticates users using specific HTTP header, you should configure Woodpecker CI to forward authentication header to Gitea using the following configuration options (assumes Woodpecker CI listening on http://192.168.1.100:8000/ behind the proxy and authentication header name `X-Forward-Username`):
```WOODPECKER_HOST_INTERNAL=http://192.168.1.100:8000/
WOODPECKER_GITEA_REV_PROXY_AUTH=true
WOODPECKER_GITEA_REV_PROXY_AUTH_HEADER=X-Forward-Username
```
Note: for this to work, Gitea must be configured for reverse proxy authentication and must accept HTTP header auth in API calls (`ENABLE_REVERSE_PROXY_AUTHENTICATION` option enabled).

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_GITEA`
> Default: `false`

Enables the Gitea driver.

### `WOODPECKER_GITEA_URL`
> Default: `https://try.gitea.io`

Configures the Gitea server address.

### `WOODPECKER_GITEA_CLIENT`
> Default: empty

Configures the Gitea OAuth client id. This is used to authorize access.

### `WOODPECKER_GITEA_CLIENT_FILE`
> Default: empty

Read the value for `WOODPECKER_GITEA_CLIENT` from the specified filepath

### `WOODPECKER_GITEA_SECRET`
> Default: empty

Configures the Gitea OAuth client secret. This is used to authorize access.

### `WOODPECKER_GITEA_SECRET_FILE`
> Default: empty

Read the value for `WOODPECKER_GITEA_SECRET` from the specified filepath

### `WOODPECKER_GITEA_SKIP_VERIFY`
> Default: `false`

Configure if SSL verification should be skipped.

### `WOODPECKER_GITEA_REV_PROXY_AUTH`
> Default: false

Enable gitea authentication using HTTP header specified in `WOODPECKER_GITEA_REV_PROXY_AUTH_HEADER` option.

### `WOODPECKER_GITEA_REV_PROXY_AUTH_HEADER`
> Default: empty

HTTP header name with authenticated user login to use when authenicating to Gitea with HTTP header (`WOODPECKER_GITEA_REV_PROXY_AUTH` enabled).

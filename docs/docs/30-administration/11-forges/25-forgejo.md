# Forgejo

Woodpecker comes with built-in support for Forgejo. To enable Forgejo you should configure the Woodpecker container using the following environment variables:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_FORGEJO=true
+     - WOODPECKER_FORGEJO_URL=${WOODPECKER_FORGEJO_URL}
+     - WOODPECKER_FORGEJO_CLIENT=${WOODPECKER_FORGEJO_CLIENT}
+     - WOODPECKER_FORGEJO_SECRET=${WOODPECKER_FORGEJO_SECRET}

  woodpecker-agent:
    [...]
```

## Registration

Register your application with Forgejo to create your client id and secret. You can find the OAuth applications settings of Forgejo at `https://forgejo.<host>/user/settings/`. It is very import the authorization callback URL matches your http(s) scheme and hostname exactly with `https://<host>/authorize` as the path.

If you run the Woodpecker CI server on the same host as the Forgejo instance, you might also need to allow local connections in Forgejo, since version `v1.16`. Otherwise webhooks will fail. Add the following lines to your Forgejo configuration (usually at `/etc/forgejo/conf/app.ini`).
```ini
...
[webhook]
ALLOWED_HOST_LIST=external,loopback
```
For reference see [Configuration Cheat Sheet](https://docs.gitea.io/en-us/config-cheat-sheet/#webhook-webhook).

![forgejo oauth setup](forgejo_oauth.gif)


## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_FORGEJO`
> Default: `false`

Enables the Forgejo driver.

### `WOODPECKER_FORGEJO_URL`
> Default: `https://codeberg.org`

Configures the Forgejo server address.

### `WOODPECKER_FORGEJO_CLIENT`
> Default: empty

Configures the Forgejo OAuth client id. This is used to authorize access.

### `WOODPECKER_FORGEJO_CLIENT_FILE`
> Default: empty

Read the value for `WOODPECKER_FORGEJO_CLIENT` from the specified filepath

### `WOODPECKER_FORGEJO_SECRET`
> Default: empty

Configures the Forgejo OAuth client secret. This is used to authorize access.

### `WOODPECKER_FORGEJO_SECRET_FILE`
> Default: empty

Read the value for `WOODPECKER_FORGEJO_SECRET` from the specified filepath

### `WOODPECKER_FORGEJO_SKIP_VERIFY`
> Default: `false`

Configure if SSL verification should be skipped.

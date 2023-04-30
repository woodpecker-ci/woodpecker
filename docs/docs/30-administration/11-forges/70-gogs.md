# Gogs

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_GOGS`
> Default: `false`

Enables the Gogs driver.

### `WOODPECKER_GOGS_URL`
> Default: `https://try.gogs.io`

Configures the Gogs server address.

### `WOODPECKER_GOGS_GIT_USERNAME`
> Default: empty

This username is used to authenticate and clone all private repositories.

### `WOODPECKER_GOGS_GIT_USERNAME_FILE`
> Default: empty

Read the value for `WOODPECKER_GOGS_GIT_USERNAME` from the specified filepath

### `WOODPECKER_GOGS_GIT_PASSWORD`
> Default: empty

The password is used to authenticate and clone all private repositories.

### `WOODPECKER_GOGS_GIT_PASSWORD_FILE`
> Default: empty

Read the value for `WOODPECKER_GOGS_GIT_PASSWORD` from the specified filepath

### `WOODPECKER_GOGS_SKIP_VERIFY`
> Default: `false`

Configure if SSL verification should be skipped.

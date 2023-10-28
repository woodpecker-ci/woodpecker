# Coding

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_CODING`
>
> Default: `false`

Enables the Coding driver.

### `WOODPECKER_CODING_URL`
>
> Default: `https://coding.net`

Configures the Coding server address.

### `WOODPECKER_CODING_CLIENT`
>
> Default: empty

Configures the Coding OAuth client id. This is used to authorize access.

### `WOODPECKER_CODING_SECRET`
>
> Default: empty

Configures the Coding OAuth client secret. This is used to authorize access.

### `WOODPECKER_CODING_SCOPE`
>
> Default: `user, project, project:depot`

Comma-separated list of OAuth scopes.

### `WOODPECKER_CODING_GIT_MACHINE`
>
> Default: `git.coding.net`

TODO

### `WOODPECKER_CODING_GIT_USERNAME`
>
> Default: empty

This username is used to authenticate and clone all private repositories.

### `WOODPECKER_CODING_GIT_PASSWORD`
>
> Default: empty

The password is used to authenticate and clone all private repositories.

### `WOODPECKER_CODING_SKIP_VERIFY`
>
> Default: `false`

Configure if SSL verification should be skipped.

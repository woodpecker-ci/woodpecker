# GitLab

Woodpecker comes with built-in support for the GitLab version 8.2 and higher. To enable GitLab you should configure the Woodpecker container using the following environment variables:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
+     - WOODPECKER_GITLAB=true
+     - WOODPECKER_GITLAB_URL=http://gitlab.mycompany.com
+     - WOODPECKER_GITLAB_CLIENT=95c0282573633eb25e82
+     - WOODPECKER_GITLAB_SECRET=30f5064039e6b359e075

  woodpecker-agent:
    [...]
```

## Registration

You must register your application with GitLab in order to generate a Client and Secret. Navigate to your account settings and choose Applications from the menu, and click New Application.

Please use `http://woodpecker.mycompany.com/authorize` as the Authorization callback URL. Grant `api` scope to the application.

If you run the Woodpecker CI server on the same host as the GitLab instance, you might also need to allow local connections in GitLab, otherwise API requests will fail. In GitLab, navigate to the Admin dashboard, then go to `Settings > Network > Outbound requests` and enable `Allow requests to the local network from web hooks and services`.

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_GITLAB`
> Default: `false`

Enables the GitLab driver.

### `WOODPECKER_GITLAB_URL`
> Default: `https://gitlab.com`

Configures the GitLab server address.

### `WOODPECKER_GITLAB_CLIENT`
> Default: empty

Configures the GitLab OAuth client id. This is used to authorize access.

### `WOODPECKER_GITLAB_CLIENT_FILE`
> Default: empty

Read the value for `WOODPECKER_GITLAB_CLIENT` from the specified filepath

### `WOODPECKER_GITLAB_SECRET`
> Default: empty

Configures the GitLab OAuth client secret. This is used to authorize access.

### `WOODPECKER_GITLAB_SECRET_FILE`
> Default: empty

Read the value for `WOODPECKER_GITLAB_SECRET` from the specified filepath

### `WOODPECKER_GITLAB_SKIP_VERIFY`
> Default: `false`

Configure if SSL verification should be skipped.

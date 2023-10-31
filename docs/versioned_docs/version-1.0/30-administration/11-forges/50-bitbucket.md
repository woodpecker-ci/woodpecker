# Bitbucket

Woodpecker comes with built-in support for Bitbucket Cloud. To enable Bitbucket Cloud you should configure the Woodpecker container using the following environment variables:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_BITBUCKET=true
+     - WOODPECKER_BITBUCKET_CLIENT=95c0282573633eb25e82
+     - WOODPECKER_BITBUCKET_SECRET=30f5064039e6b359e075

  woodpecker-agent:
    [...]
```

## Registration

You must register your application with Bitbucket in order to generate a client and secret. Navigate to your account settings and choose OAuth from the menu, and click Add Consumer.

Please use the Authorization callback URL:

```nohighlight
http://woodpecker.mycompany.com/authorize
```

Please also be sure to check the following permissions:

```nohighlight
Account:Email
Account:Read
Team Membership:Read
Repositories:Read
Webhooks:Read and Write
```

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_BITBUCKET`
> Default: `false`

Enables the Bitbucket driver.

### `WOODPECKER_BITBUCKET_CLIENT`
> Default: empty

Configures the Bitbucket OAuth client id. This is used to authorize access.

### `WOODPECKER_BITBUCKET_CLIENT_FILE`
> Default: empty

Read the value for `WOODPECKER_BITBUCKET_CLIENT` from the specified filepath

### `WOODPECKER_BITBUCKET_SECRET`
> Default: empty

Configures the Bitbucket OAuth client secret. This is used to authorize access.

### `WOODPECKER_BITBUCKET_SECRET_FILE`
> Default: empty

Read the value for `WOODPECKER_BITBUCKET_SECRET` from the specified filepath

## Missing Features

Merge requests are not currently supported. We are interested in patches to include this functionality.
If you are interested in contributing to Woodpecker and submitting a patch please **contact us** via [Discord](https://discord.gg/fcMQqSMXJy) or [Matrix](https://matrix.to/#/#WoodpeckerCI-Develop:obermui.de).

---
toc_max_heading_level: 2
---

# Bitbucket Datacenter / Server

:::warning
Woodpecker comes with experimental support for Bitbucket Datacenter / Server, formerly known as Atlassian Stash.
:::

To enable Bitbucket Server you should configure the Woodpecker container using the following environment variables:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_BITBUCKET_DC=true
+      - WOODPECKER_BITBUCKET_DC_GIT_USERNAME=foo
+      - WOODPECKER_BITBUCKET_DC_GIT_PASSWORD=bar
+      - WOODPECKER_BITBUCKET_DC_CLIENT_ID=xxx
+      - WOODPECKER_BITBUCKET_DC_CLIENT_SECRET=yyy
+      - WOODPECKER_BITBUCKET_DC_URL=http://stash.mycompany.com

   woodpecker-agent:
     [...]
```

## Service Account

Woodpecker uses `git+https` to clone repositories, however, Bitbucket Server does not currently support cloning repositories with an OAuth token. To work around this limitation, you must create a service account and provide the username and password to Woodpecker. This service account will be used to authenticate and clone private repositories.

## Registration

Woodpecker must be registered with Bitbucket Datacenter / Server.
In the administration section of Bitbucket choose "Application Links" and then "Create link".
Woodpecker should be listed as "External Application" and the direction should be set to "Incoming".
Note the client id and client secret of the registration to be used in the configuration of Woodpecker.

See also [Configure an incoming link](https://confluence.atlassian.com/bitbucketserver/configure-an-incoming-link-1108483657.html).

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_BITBUCKET_DC`

> Default: `false`

Enables the Bitbucket Server driver.

### `WOODPECKER_BITBUCKET_DC_URL`

> Default: empty

Configures the Bitbucket Server address.

### `WOODPECKER_BITBUCKET_DC_CLIENT_ID`

> Default: empty

Configures your Bitbucket Server OAUth 2.0 client id.

### `WOODPECKER_BITBUCKET_DC_CLIENT_SECRET`

> Default: empty

Configures your Bitbucket Server OAUth 2.0 client secret.

### `WOODPECKER_BITBUCKET_DC_GIT_USERNAME`

> Default: empty

This username is used to authenticate and clone all private repositories.

### `WOODPECKER_BITBUCKET_DC_GIT_USERNAME_FILE`

> Default: empty

Read the value for `WOODPECKER_BITBUCKET_DC_GIT_USERNAME` from the specified filepath

### `WOODPECKER_BITBUCKET_DC_GIT_PASSWORD`

> Default: empty

The password is used to authenticate and clone all private repositories.

### `WOODPECKER_BITBUCKET_DC_GIT_PASSWORD_FILE`

> Default: empty

Read the value for `WOODPECKER_BITBUCKET_DC_GIT_PASSWORD` from the specified filepath

### `WOODPECKER_BITBUCKET_DC_SKIP_VERIFY`

> Default: `false`

Configure if SSL verification should be skipped.

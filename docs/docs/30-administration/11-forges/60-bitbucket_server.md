# Bitbucket Server

Woodpecker comes with experimental support for Bitbucket Server, formerly known as Atlassian Stash. To enable Bitbucket Server you should configure the Woodpecker container using the following environment variables:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_STASH=true
+     - WOODPECKER_STASH_GIT_USERNAME=foo
+     - WOODPECKER_STASH_GIT_PASSWORD=bar
+     - WOODPECKER_STASH_CONSUMER_KEY=95c0282573633eb25e82
+     - WOODPECKER_STASH_CONSUMER_RSA=/etc/bitbucket/key.pem
+     - WOODPECKER_STASH_URL=http://stash.mycompany.com
    volumes:
+     - /path/to/key.pem:/path/to/key.pem

  woodpecker-agent:
    [...]
```

## Private Key File

The OAuth process in Bitbucket server requires a private and a public RSA certificate. This is how you create the private RSA certificate.

```nohighlight
openssl genrsa -out /etc/bitbucket/key.pem 1024
```

This stores the private RSA certificate in `key.pem`. The next command generates the public RSA certificate and stores it in `key.pub`.

```nohighlight
openssl rsa -in /etc/bitbucket/key.pem -pubout >> /etc/bitbucket/key.pub
```

Please note that the private key file can be mounted into your Woodpecker container at runtime or as an environment variable

Private key file mounted into your Woodpecker container at runtime as a volume.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
      - WOODPECKER_STASH=true
      - WOODPECKER_STASH_GIT_USERNAME=foo
      - WOODPECKER_STASH_GIT_PASSWORD=bar
      - WOODPECKER_STASH_CONSUMER_KEY=95c0282573633eb25e82
+     - WOODPECKER_STASH_CONSUMER_RSA=/etc/bitbucket/key.pem
      - WOODPECKER_STASH_URL=http://stash.mycompany.com
+  volumes:
+     - /etc/bitbucket/key.pem:/etc/bitbucket/key.pem

  woodpecker-agent:
    [...]
```

Private key as environment variable

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
      - WOODPECKER_STASH=true
      - WOODPECKER_STASH_GIT_USERNAME=foo
      - WOODPECKER_STASH_GIT_PASSWORD=bar
      - WOODPECKER_STASH_CONSUMER_KEY=95c0282573633eb25e82
+     - WOODPECKER_STASH_CONSUMER_RSA_STRING=contentOfPemKeyAsString
      - WOODPECKER_STASH_URL=http://stash.mycompany.com

  woodpecker-agent:
    [...]
```

## Service Account

Woodpecker uses `git+https` to clone repositories, however, Bitbucket Server does not currently support cloning repositories with OAuth token. To work around this limitation, you must create a service account and provide the username and password to Woodpecker. This service account will be used to authenticate and clone private repositories.

## Registration

You must register your application with Bitbucket Server in order to generate a consumer key. Navigate to your account settings and choose Applications from the menu, and click Register new application. Now copy & paste the text value from `/etc/bitbucket/key.pub` into the `Public Key` in the incoming link part of the application registration.

Please use http://woodpecker.mycompany.com/authorize as the Authorization callback URL.

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

### `WOODPECKER_STASH`
> Default: `false`

Enables the Bitbucket Server driver.

### `WOODPECKER_STASH_URL`
> Default: empty

Configures the Bitbucket Server address.

### `WOODPECKER_STASH_CONSUMER_KEY`
> Default: empty

Configures your Bitbucket Server consumer key.

### `WOODPECKER_STASH_CONSUMER_KEY_FILE`
> Default: empty

Read the value for `WOODPECKER_STASH_CONSUMER_KEY` from the specified filepath

### `WOODPECKER_STASH_CONSUMER_RSA`
> Default: empty

Configures the path to your Bitbucket Server private key file.

### `WOODPECKER_STASH_CONSUMER_RSA_STRING`
> Default: empty

Configures your Bitbucket Server private key.

### `WOODPECKER_STASH_GIT_USERNAME`
> Default: empty

This username is used to authenticate and clone all private repositories.

### `WOODPECKER_STASH_GIT_USERNAME_FILE`
> Default: empty

Read the value for `WOODPECKER_STASH_GIT_USERNAME` from the specified filepath

### `WOODPECKER_STASH_GIT_PASSWORD`
> Default: empty

The password is used to authenticate and clone all private repositories.

### `WOODPECKER_STASH_GIT_PASSWORD_FILE`
> Default: empty

Read the value for `WOODPECKER_STASH_GIT_PASSWORD` from the specified filepath

### `WOODPECKER_STASH_SKIP_VERIFY`
> Default: `false`

Configure if SSL verification should be skipped.

# Gitea

Woodpecker comes with built-in support for Gitea. To enable Gitea you should configure the Woodpecker container using the following environment variables:

```diff
version: '3'

services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    ports:
      - 80:8000
      - 9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
    restart: always
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
+     - WOODPECKER_GITEA=true
+     - WOODPECKER_GITEA_URL=${WOODPECKER_GITEA_URL}
+     - WOODPECKER_GITEA_CLIENT=${WOODPECKER_GITEA_CLIENT}
+     - WOODPECKER_GITEA_SECRET=${WOODPECKER_GITEA_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}

  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

## Registration

Register your application with Gitea to create your client id and secret. You can find the OAuth applications settings of Gitea at `https://gitea.<host>/user/settings/`. It is very import the authorization callback URL matches your http(s) scheme and hostname exactly with `https://<host>/authorize` as the path.

![gitea oauth setup](gitea_oauth.gif)


## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

`WOODPECKER_GITEA=true`
: Set to true to enable the Gitea driver.

`WOODPECKER_GITEA_URL=https://try.gitea.io`
: Gitea server address.

`WOODPECKER_GITEA_CLIENT`
: Gitea oauth2 client id.

`WOODPECKER_GITEA_SECRET`
: Gitea oauth2 client secret.

`WOODPECKER_GITEA_CONTEXT=continuous-integration/woodpecker`
: Customize the Gitea status message context

`WOODPECKER_GITEA_GIT_USERNAME`
: Optional. Use a single machine account username to clone all repositories.

`WOODPECKER_GITEA_GIT_PASSWORD`
: Optional. Use a single machine account password to clone all repositories.

`WOODPECKER_GITEA_PRIVATE_MODE=true`
: Set to true if Gitea is running in private mode.

`WOODPECKER_GITEA_SKIP_VERIFY=false`
: Set to true to disable SSL verification.

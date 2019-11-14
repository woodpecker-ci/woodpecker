# Gitlab

Woodpecker comes with built-in support for the GitLab version 8.2 and higher. To enable GitLab you should configure the Woodpecker container using the following environment variables:

```diff
version: '2'

services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    ports:
      - 80:8000
      - 9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
    restart: always
    environment:
+     - DRONE_GITLAB=true
+     - DRONE_GITLAB_CLIENT=95c0282573633eb25e82
+     - DRONE_GITLAB_SECRET=30f5064039e6b359e075
+     - DRONE_GITLAB_URL=http://gitlab.mycompany.com
      - DRONE_SECRET=${DRONE_SECRET}

  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=woodpecker-server:9000
      - DRONE_SECRET=${DRONE_SECRET}
```

## Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

DRONE_GITLAB=true
: Set to true to enable the GitLab driver.

DRONE_GITLAB_URL=`https://gitlab.com`
: GitLab Server address.

DRONE_GITLAB_CLIENT
: GitLab oauth2 client id.

DRONE_GITLAB_SECRET
: GitLab oauth2 client secret.

DRONE_GITLAB_GIT_USERNAME
: Optional. Use a single machine account username to clone all repositories.

DRONE_GITLAB_GIT_PASSWORD
: Optional. Use a single machine account password to clone all repositories.

DRONE_GITLAB_SKIP_VERIFY=false
: Set to true to disable SSL verification.

DRONE_GITLAB_PRIVATE_MODE=false
: Set to true if GitLab is running in private mode.

## Registration

You must register your application with GitLab in order to generate a Client and Secret. Navigate to your account settings and choose Applications from the menu, and click New Application.

Please use `http://woodpecker.mycompany.com/authorize` as the Authorization callback URL. Grant `api` scope to the application.

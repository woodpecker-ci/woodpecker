Drone comes with built-in support for GitHub and GitHub Enterprise. To enable GitHub you should configure the Drone container using the following environment variables:

```diff
version: '2'

services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
      - 80:8000
      - 9000
    volumes:
      - /var/lib/drone:/var/lib/drone/
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
+     - DRONE_GITHUB=true
+     - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
+     - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}

  drone-agent:
    image: drone/agent:{{% version %}}
    restart: always
    depends_on:
      - drone-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=drone-server:9000
      - DRONE_SECRET=${DRONE_SECRET}
```

# Registration

Register your application with GitHub to create your client id and secret. It is very import the authorization callback URL matches your http(s) scheme and hostname exactly with `<scheme>://<host>/authorize` as the path.

Please use this screenshot for reference:

![github oauth setup](github_oauth.png)

# Configuration

This is a full list of configuration options. Please note that many of these options use default configuration values that should work for the majority of installations.

DRONE_GITHUB=true
: Set to true to enable the GitHub driver.

DRONE_GITHUB_URL=`https://github.com`
: GitHub server address.

DRONE_GITHUB_CLIENT
: Github oauth2 client id.

DRONE_GITHUB_SECRET
: Github oauth2 client secret.

DRONE_GITHUB_SCOPE=repo,repo:status,user:email,read:org
: Comma-separated Github oauth scope.

DRONE_GITHUB_GIT_USERNAME
: Optional. Use a single machine account username to clone all repositories.

DRONE_GITHUB_GIT_PASSWORD
: Optional. Use a single machine account password to clone all repositories.

DRONE_GITHUB_PRIVATE_MODE=false
: Set to true if Github is running in private mode.

DRONE_GITHUB_MERGE_REF=true
: Set to true to use the `refs/pulls/%d/merge` vs `refs/pulls/%d/head`

DRONE_GITHUB_CONTEXT=continuous-integration/drone
: Customize the GitHub status message context

DRONE_GITHUB_SKIP_VERIFY=false
: Set to true to disable SSL verification.

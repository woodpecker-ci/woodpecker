# Server configuration

## User registration

Registration is closed by default. While disabled an administrator needs to add new users manually (exp. `woodpecker-cli user add`).

If registration is open every user with an account at the configured [SCM](docs/administration/vcs/overview) can login to Woodpecker.
This example enables open registration for users that are members of approved organizations:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_OPEN=true
+     - WOODPECKER_ORGS=dolores,dogpatch

```

## Administrators

Administrators should also be enumerated in your configuration.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_ADMIN=johnsmith,janedoe
```

## Filtering repositories

Woodpecker operates with the user's OAuth permission. Due to the coarse permission handling of GitHub, you may end up syncing more repos into Woodpecker than preferred.

Use the `WOODPECKER_REPO_OWNERS` variable to filter which GitHub user's repos should be synced only. You typically want to put here your company's GitHub name.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_REPO_OWNERS=mycompany,mycompanyossgithubuser
```

## Global registry setting

If you want to make available a specific private registry to all pipelines, use the `WOODPECKER_DOCKER_CONFIG` server configuration.
Point it to your server's docker config.

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_DOCKER_CONFIG=/home/user/.docker/config.json
```

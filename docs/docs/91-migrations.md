# Migrations

Some versions need some changes to the server configuration or the pipeline configuration files.

## 0.15.0

- Default value for custom pipeline path is now empty / un-set which results in following resolution:

  `.woodpecker/*.yml` -> `.woodpecker.yml` -> `.drone.yml`

  Only projects created after updating will have an empty value by default. Existing projects will stick to the current pipeline path which is `.drone.yml` in most cases.

  Read more about it at the [Project Settings](/docs/usage/project-settings#pipeline-path)

- From version `0.15.0` ongoing there will be three types of docker images: `latest`, `next` and `x.x.x` with an alpine variant for each type like `latest-alpine`.
  If you used `latest` before to try pre-release features you should switch to `next` after this release.

- Dropped support for `DRONE_*` environment variables. The according `WOODPECKER_*` variables must be used instead.
  Additionally some alternative namings have been removed to simplify maintenance:
  - `WOODPECKER_AGENT_SECRET` replaces `WOODPECKER_SECRET`, `DRONE_SECRET`, `WOODPECKER_PASSWORD`, `DRONE_PASSWORD` and `DRONE_AGENT_SECRET`.
  - `WOODPECKER_HOST` replaces `DRONE_HOST` and `DRONE_SERVER_HOST`.
  - `WOODPECKER_DATABASE_DRIVER` replaces `DRONE_DATABASE_DRIVER` and `DATABASE_DRIVER`.
  - `WOODPECKER_DATABASE_DATASOURCE` replaces `DRONE_DATABASE_DATASOURCE` and `DATABASE_CONFIG`.

- Dropped support for `DRONE_*` environment variables in pipeline steps. Pipeline meta-data can be accessed with `CI_*` variables.
  - `CI_*` prefix replaces `DRONE_*`
  - `CI` value is now `woodpecker`
  - `DRONE=true` has been removed
  - Some variables got deprecated and will be removed in future versions. Please migrate to the new names. Same applies for `DRONE_` of them.
    - CI_ARCH => use CI_SYSTEM_ARCH
    - CI_COMMIT => CI_COMMIT_SHA
    - CI_TAG => CI_COMMIT_TAG
    - CI_PULL_REQUEST => CI_COMMIT_PULL_REQUEST
    - CI_REMOTE_URL => use CI_REPO_REMOTE
    - CI_REPO_BRANCH => use CI_REPO_DEFAULT_BRANCH
    - CI_PARENT_BUILD_NUMBER => use CI_BUILD_PARENT
    - CI_BUILD_TARGET => use CI_BUILD_DEPLOY_TARGET
    - CI_DEPLOY_TO => use CI_BUILD_DEPLOY_TARGET
    - CI_COMMIT_AUTHOR_NAME => use CI_COMMIT_AUTHOR
    - CI_PREV_COMMIT_AUTHOR_NAME => use CI_PREV_COMMIT_AUTHOR
    - CI_SYSTEM => use CI_SYSTEM_NAME
    - CI_BRANCH => use CI_COMMIT_BRANCH
    - CI_SOURCE_BRANCH => use CI_COMMIT_SOURCE_BRANCH
    - CI_TARGET_BRANCH => use CI_COMMIT_TARGET_BRANCH

  For all available variables and their descriptions have a look at [built-in-environment-variables](/docs/usage/environment#built-in-environment-variables).

- Prometheus metrics have been changed from `drone_*` to `woodpecker_*`

- Base path has moved from `/var/lib/drone` to `/var/lib/woodpecker`

- Default SQLite database location has changed:
  - `/var/lib/drone/drone.sqlite` -> `/var/lib/woodpecker/woodpecker.sqlite`
  - `drone.sqlite` -> `woodpecker.sqlite`

- Plugin Settings moved into `settings` section:
  ```diff
   pipline:
   something:
     image: my/plugin
  -  setting1: foo
  -  setting2: bar
  +  settings:
  +    setting1: foo
  +    setting2: bar
  ```

- Dropped support for manually setting the agents platform with `WOODPECKER_PLATFORM`. The platform is now automatically detected.

- Use `WOODPECKER_STATUS_CONTEXT` instead of the deprecated options `WOODPECKER_GITHUB_CONTEXT` and `WOODPECKER_GITEA_CONTEXT`.

## 0.14.0

No breaking changes

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
  - Use `CI_REPO_DEFAULT_BRANCH` instead of `DRONE_REPO_BRANCH`
  - Use `CI_COMMIT_BRANCH`, `CI_COMMIT_SOURCE_BRANCH` and `CI_COMMIT_TARGET_BRANCH` instead of `DRONE_BRANCH`, `DRONE_TARGET_BRANCH` and `DRONE_SOURCE_BRANCH` variables
  - Use `CI_COMMIT_AUTHOR` instead of `DRONE_COMMIT_AUTHOR_NAME`
  - Use `CI_COMMIT_TAG` instead of `DRONE_TAG`
  - Use `CI_COMMIT_PULL_REQUEST` instead of `DRONE_PULL_REQUEST`
  - Use `CI_BUILD_DEPLOY_TARGET` instead of `DRONE_DEPLOY_TO`
  - Use `CI_REPO_REMOTE` instead of `DRONE_REMOTE_URL`
  - Use `CI_AGENT_ARCH` instead of `DRONE_ARCH`
  - Use `CI_COMMIT_SHA` instead of `DRONE_COMMIT`

  For all available variables and their descriptions have a look at [built-in-environment-variables](/docs/usage/environment#built-in-environment-variables).

- Prometheus metrics have been changed from `drone_*` to `woodpecker_*`

- Base path has moved from `/var/lib/drone` to `/var/lib/woodpecker`

- Default SQLite database location has changed:
  - `/var/lib/drone/drone.sqlite` -> `/var/lib/woodpecker/woodpecker.sqlite`
  - `drone.sqlite` -> `woodpecker.sqlite`

- ...

## 0.14.0

No breaking changes

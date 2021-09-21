# Migrations

Some versions need some changes to the server configuration or the pipeline configuration files.

## 0.15.0

- Default value for custom pipeline path is now empty / un-set which results in following resolution:

  `.woodpecker/*.yml` -> `.woodpecker.yml` -> `.drone.yml`

  Only projects created after updating will have an empty value by default. Existing projects will stick to the current pipeline path which is `.drone.yml` in most cases.

  Read more about it at the [Project Settings](/docs/usage/project-settings#pipeline-path)

- Dropped support for `DRONE_*` environment variables. The according `WOODPECKER_*` variables must be used instead.
  Additionally some alternative namings have been removed to simplify maintenance:
  - `WOODPECKER_AGENT_SECRET` replaces `WOODPECKER_SECRET`, `DRONE_SECRET`, `WOODPECKER_PASSWORD`, `DRONE_PASSWORD` and `DRONE_AGENT_SECRET`.
  - `WOODPECKER_SERVER_HOST` replaces `WOODPECKER_HOST`, `DRONE_HOST` and `DRONE_SERVER_HOST`.
  - `WOODPECKER_DATABASE_DRIVER` replaces `DRONE_DATABASE_DRIVER` and `DATABASE_DRIVER`.
  - `WOODPECKER_DATABASE_DATASOURCE` replaces `DRONE_DATABASE_DATASOURCE` and `DATABASE_CONFIG`.

- ...

## 0.14.0

No breaking changes

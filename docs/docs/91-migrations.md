# Migrations

Some versions need some changes to the server configuration or the pipeline configuration files.

## next (1.1.0)

- Drop deprecated `CI_BUILD_*`, `CI_PREV_BUILD_*`, `CI_JOB_*`, `*_LINK`, `CI_SYSTEM_ARCH`, `CI_REPO_REMOTE` built-in environment variables
- Drop deprecated `pipeline:` keyword for steps in yaml config
- Drop deprecated `branches:` keyword for global branch filter
- Deprecate `platform:` filter in favor of `labels:`, [read more](./20-usage/20-pipeline-syntax.md#filter-by-platform)
- Ignore branches without a Woodpecker configuration

## 1.0.0

- The signature used to verify extensions calls (like those used for the [config-extension](./30-administration/100-external-configuration-api.md)) done by the Woodpecker server switched from using a shared-secret HMac to an ed25519 key-pair. Read more about it at the [config-extensions](./30-administration/100-external-configuration-api.md) documentation.
- Refactored support of old agent filter labels and expression. Learn how to use the new [filter](./20-usage/20-pipeline-syntax.md#labels)
- Renamed step environment variable `CI_SYSTEM_ARCH` to `CI_SYSTEM_PLATFORM`. Same applies for the cli exec variable.
- Renamed environment variables `CI_BUILD_*` and `CI_PREV_BUILD_*` to `CI_PIPELINE_*` and `CI_PREV_PIPELINE_*`, old ones are still available but deprecated
- Renamed environment variables `CI_JOB_*` to `CI_STEP_*`, old ones are still available but deprecated
- Renamed environment variable `CI_REPO_REMOTE` to `CI_REPO_CLONE_URL`, old is still available but deprecated
- Renamed environment variable `*_LINK` to `*_URL`, old ones are still available but deprecated
- Renamed API endpoints for pipelines (`<owner>/<repo>/builds/<buildId>` -> `<owner>/<repo>/pipelines/<pipelineId>`), old ones are still available but deprecated
- Updated Prometheus gauge `build_*` to `pipeline_*`
- Updated Prometheus gauge `*_job_*` to `*_step_*`
- Renamed config env `WOODPECKER_MAX_PROCS` to `WOODPECKER_MAX_WORKFLOWS` (still available as fallback)
- The pipelines are now also read from `.yaml` files, the new default order is `.woodpecker/*.yml` and `.woodpecker/*.yaml` (without any prioritization) -> `.woodpecker.yml` ->  `.woodpecker.yaml`
- Dropped support for [Coding](https://coding.net/), [Gogs](https://gogs.io) and Bitbucket Server (Stash).
- `/api/queue/resume` & `/api/queue/pause` endpoint methods were changed from `GET` to `POST`
- rename `pipeline:` key in your workflow config to `steps:`
- If you want to migrate old logs to the new format, watch the error messages on start. If there are none we are good to go, else you have to plan a migration that can take hours. Set `WOODPECKER_MIGRATIONS_ALLOW_LONG` to true and let it run.
- Using `repo-id` in favor of `owner/repo` combination
  - :warning: The api endpoints `/api/repos/{owner}/{repo}/...` were replaced by new endpoints using the repos id `/api/repos/{repo-id}`
  - To find the id of a repo use the `/api/repos/lookup/{repo-full-name-with-slashes}` endpoint.
  - The existing badge endpoint `/api/badges/{owner}/{repo}` will still work, but whenever possible try to use the new endpoint using the `repo-id`: `/api/badges/{repo-id}`.
  - The UI urls for a repository changed from `/repos/{owner}/{repo}/...` to `/repos/{repo-id}/...`. You will be redirected automatically when using the old url.
  - The woodpecker-go api-client is now using the `repo-id` instead of `owner/repo` for all functions
- Using `org-id` in favour of `owner` name
  - :warning: The api endpoints `/api/orgs/{owner}/...` were replaced by new endpoints using the orgs id `/api/repos/{org-id}`
  - To find the id of orgs use the `/api/orgs/lookup/{org_full_name}` endpoint.
  - The UI urls for a organization changed from `/org/{owner}/...` to `/orgs/{org-id}/...`. You will be redirected automatically when using the old url.
  - The woodpecker-go api-client is now using the `org-id` instead of `org name` for all functions
- The `command:` field has been removed from steps. If you were using it, please check if the entrypoint of the image you used is a shell.
  - If it is a shell, simply rename `command:` to `commands:`.
  - If it's not, you need to prepend the entrypoint before and also rename it (e.g., `commands: <entrypoint> <cmd>`).

## 0.15.0

- Default value for custom pipeline path is now empty / un-set which results in following resolution:

  `.woodpecker/*.yml` -> `.woodpecker.yml` -> `.drone.yml`

  Only projects created after updating will have an empty value by default. Existing projects will stick to the current pipeline path which is `.drone.yml` in most cases.

  Read more about it at the [Project Settings](./20-usage/71-project-settings.md#pipeline-path)

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

  For all available variables and their descriptions have a look at [built-in-environment-variables](./20-usage/50-environment.md#built-in-environment-variables).

- Prometheus metrics have been changed from `drone_*` to `woodpecker_*`

- Base path has moved from `/var/lib/drone` to `/var/lib/woodpecker`

- Default workspace base path has moved from `/drone` to `/woodpecker`

- Default SQLite database location has changed:
  - `/var/lib/drone/drone.sqlite` -> `/var/lib/woodpecker/woodpecker.sqlite`
  - `drone.sqlite` -> `woodpecker.sqlite`

- Plugin Settings moved into `settings` section:

  ```diff
   steps:
   something:
     image: my/plugin
  -  setting1: foo
  -  setting2: bar
  +  settings:
  +    setting1: foo
  +    setting2: bar
  ```

- `WOODPECKER_DEBUG` option for server and agent got removed in favor of `WOODPECKER_LOG_LEVEL=debug`

- Remove unused server flags which can safely be removed from your server config: `WOODPECKER_QUIC`,  `WOODPECKER_GITHUB_SCOPE`, `WOODPECKER_GITHUB_GIT_USERNAME`, `WOODPECKER_GITHUB_GIT_PASSWORD`, `WOODPECKER_GITHUB_PRIVATE_MODE`, `WOODPECKER_GITEA_GIT_USERNAME`, `WOODPECKER_GITEA_GIT_PASSWORD`, `WOODPECKER_GITEA_PRIVATE_MODE`, `WOODPECKER_GITLAB_GIT_USERNAME`, `WOODPECKER_GITLAB_GIT_PASSWORD`, `WOODPECKER_GITLAB_PRIVATE_MODE`

- Dropped support for manually setting the agents platform with `WOODPECKER_PLATFORM`. The platform is now automatically detected.

- Use `WOODPECKER_STATUS_CONTEXT` instead of the deprecated options `WOODPECKER_GITHUB_CONTEXT` and `WOODPECKER_GITEA_CONTEXT`.

## 0.14.0

No breaking changes

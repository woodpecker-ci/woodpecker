# Migrations

Some versions need some changes to the server configuration or the pipeline configuration files.

## `next`

- Deprecate `WOODPECKER_FILTER_LABELS` use `WOODPECKER_AGENT_LABELS`
- Removed built-in environment variables:
  - `CI_COMMIT_URL` use `CI_PIPELINE_FORGE_URL`
  - `CI_STEP_FINISHED` as empty during execution
  - `CI_PIPELINE_FINISHED` as empty during execution
  - `CI_PIPELINE_STATUS` was always `success`
  - `CI_STEP_STATUS` was always `success`
- Set `/woodpecker` as defautl workdir for the **woodpecker-cli** container
- Move docker resource limit settings from server into agent configuration
- Rename server environment variable `WOODPECKER_ESCALATE` to `WOODPECKER_PLUGINS_PRIVILEGED`
- All default privileged plugins (like `woodpeckerci/plugin-docker-buildx`) were removed. Please carefully [re-add those plugins](./30-administration/10-server-config.md#woodpecker_plugins_privileged) you trust and rely on.
- `WOODPECKER_DEFAULT_CLONE_IMAGE` got depricated use `WOODPECKER_DEFAULT_CLONE_PLUGIN`
- Check trusted-clone- and privileged-plugins by image name and tag (if tag is set)
- Secret filters for plugins now check against tag if specified
- Removed `WOODPECKER_DEV_OAUTH_HOST` and `WOODPECKER_DEV_GITEA_OAUTH_URL` use `WOODPECKER_EXPERT_FORGE_OAUTH_HOST`
- Compatibility mode of deprecated `pipeline:`, `platform:` and `branches:` pipeline config options are now removed and pipeline will now fail if still in use.
- Removed `steps.[name].group` in favor of `steps.[name].depends_on` (see [workflow syntax](./20-usage/20-workflow-syntax.md#depends_on) to learn how to set dependencies)
- Removed `WOODPECKER_ROOT_PATH` and `WOODPECKER_ROOT_URL` config variables. Use `WOODPECKER_HOST` with a path instead
- Pipelines without a config file will now be skipped instead of failing
- Removed implicitly defined `regcred` image pull secret name. Set it explicitly via `WOODPECKER_BACKEND_K8S_PULL_SECRET_NAMES`
- Removed `includes` and `excludes` support from **event** filter
- Removed uppercasing all secret env vars, instead, the value of the `secrets` property is used. [Read more](./20-usage/40-secrets.md#use-secrets-in-commands)
- Removed alternative names for secrets, use `environment` with `from_secret`
- Removed slice definition for env vars
- Removed `environment` filter, use `when.evaluate`
- Removed `WOODPECKER_WEBHOOK_HOST` in favor of `WOODPECKER_EXPERT_WEBHOOK_HOST`
- Migrated to rfc9421 for webhook signatures
- Renamed `start_time`, `end_time`, `created_at`, `started_at`, `finished_at` and `reviewed_at` JSON fields to `started`, `finished`, `created`, `started`, `finished`, `reviewed`
- Update all webhooks by pressing the "Repair all" button in the admin settings as the webhook token claims have changed
- Crons now use standard Linux syntax without seconds
- Replaced `configs` object by `netrc` in external configuration APIs
- Removed old API routes: `registry/` -> `registries`, `/authorize/token`
- Replaced `registry` command with `repo registry` in cli
- Disallow upgrades from 1.x, upgrade to 2.x first

## 2.0.0

- Dropped deprecated `CI_BUILD_*`, `CI_PREV_BUILD_*`, `CI_JOB_*`, `*_LINK`, `CI_SYSTEM_ARCH`, `CI_REPO_REMOTE` built-in environment variables
- Deprecated `platform:` filter in favor of `labels:`, [read more](./20-usage/20-workflow-syntax.md#filter-by-platform)
- Secrets `event` property was renamed to `events` and `image` to `images` as both are lists. The new property `events` / `images` has to be used in the api. The old properties `event` and `image` were removed.
- The secrets `plugin_only` option was removed. Secrets with images are now always only available for plugins using listed by the `images` property. Existing secrets with a list of `images` will now only be available to the listed images if they are used as a plugin.
- Removed `build` alias for `pipeline` command in CLI
- Removed `ssh` backend. Use an agent directly on the SSH machine using the `local` backend.
- Removed `/hook` and `/stream` API paths in favor of `/api/(hook|stream)`. You may need to use the "Repair repository" button in the repo settings or "Repair all" in the admin settings to recreate the forge hook.
- Removed `WOODPECKER_DOCS` config variable
- Renamed `link` to `url` (including all API fields)
- Deprecated `CI_COMMIT_URL` env var, use `CI_PIPELINE_FORGE_URL`

## 1.0.0

- The signature used to verify extension calls (like those used for the [config-extension](./30-administration/40-advanced/100-external-configuration-api.md)) done by the Woodpecker server switched from using a shared-secret HMac to an ed25519 key-pair. Read more about it at the [config-extensions](./30-administration/40-advanced/100-external-configuration-api.md) documentation.
- Refactored support for old agent filter labels and expressions. Learn how to use the new [filter](./20-usage/20-workflow-syntax.md#labels)
- Renamed step environment variable `CI_SYSTEM_ARCH` to `CI_SYSTEM_PLATFORM`. Same applies for the cli exec variable.
- Renamed environment variables `CI_BUILD_*` and `CI_PREV_BUILD_*` to `CI_PIPELINE_*` and `CI_PREV_PIPELINE_*`, old ones are still available but deprecated
- Renamed environment variables `CI_JOB_*` to `CI_STEP_*`, old ones are still available but deprecated
- Renamed environment variable `CI_REPO_REMOTE` to `CI_REPO_CLONE_URL`, old is still available but deprecated
- Renamed environment variable `*_LINK` to `*_URL`, old ones are still available but deprecated
- Renamed API endpoints for pipelines (`<owner>/<repo>/builds/<buildId>` -> `<owner>/<repo>/pipelines/<pipelineId>`), old ones are still available but deprecated
- Updated Prometheus gauge `build_*` to `pipeline_*`
- Updated Prometheus gauge `*_job_*` to `*_step_*`
- Renamed config env `WOODPECKER_MAX_PROCS` to `WOODPECKER_MAX_WORKFLOWS` (still available as fallback)
- The pipelines are now also read from `.yaml` files, the new default order is `.woodpecker/*.yml` and `.woodpecker/*.yaml` (without any prioritization) -> `.woodpecker.yml` -> `.woodpecker.yaml`
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

  Read more about it at the [Project Settings](./20-usage/75-project-settings.md#pipeline-path)

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

- Remove unused server flags which can safely be removed from your server config: `WOODPECKER_QUIC`, `WOODPECKER_GITHUB_SCOPE`, `WOODPECKER_GITHUB_GIT_USERNAME`, `WOODPECKER_GITHUB_GIT_PASSWORD`, `WOODPECKER_GITHUB_PRIVATE_MODE`, `WOODPECKER_GITEA_GIT_USERNAME`, `WOODPECKER_GITEA_GIT_PASSWORD`, `WOODPECKER_GITEA_PRIVATE_MODE`, `WOODPECKER_GITLAB_GIT_USERNAME`, `WOODPECKER_GITLAB_GIT_PASSWORD`, `WOODPECKER_GITLAB_PRIVATE_MODE`

- Dropped support for manually setting the agents platform with `WOODPECKER_PLATFORM`. The platform is now automatically detected.

- Use `WOODPECKER_STATUS_CONTEXT` instead of the deprecated options `WOODPECKER_GITHUB_CONTEXT` and `WOODPECKER_GITEA_CONTEXT`.

## 0.14.0

No breaking changes

## From Drone

:::warning
Migration from Drone is only possible if you were running Drone <= v0.8.
:::

1. Make sure you are already running Drone v0.8
2. Upgrade to Woodpecker v0.14.4, migration will be done during startup
3. Upgrade to the latest Woodpecker version. Pay attention to the breaking changes listed above.

# Migrations

## `next`

- No changes

## 3.0.0

## User-facing migrations

- The `gated` option, which restricted which pipelines can start right away without requiring approval, has been replaced by `require-approval` option. Even though ([#3348](https://github.com/woodpecker-ci/woodpecker/pull/3348)) was backported to 2.8, no default was set. Now, the default is to require approval only for forked repositories. This allows easier management of dependency bots and other trusted entities having write access to the repository.

- Environment variables must now be defined as maps. List definitions are disallowed. ([#4016](https://github.com/woodpecker-ci/woodpecker/pull/4016))

  2.x:

  ```yaml
  environment:
    - ENV1=value1
  ```

  3.x:

  ```yaml
  environment:
    ENV1: value1
  ```

- The following built-in environment variables have been removed/replaced:

  - `CI_COMMIT_URL` - `CI_PIPELINE_FORGE_URL` should be used instead
  - `CI_STEP_FINISHED` as it was empty during execution
  - `CI_PIPELINE_FINISHED` as it was empty during execution
  - `CI_PIPELINE_STATUS` due to always being set to `success`
  - `CI_STEP_STATUS` due to always being set to `success`
  - `WOODPECKER_WEBHOOK_HOST` in favor of `WOODPECKER_EXPERT_WEBHOOK_HOST`
  - Environment variables which are empty after workflow parsing are not being injected into the build but filtered out beforehand ([#4193](https://github.com/woodpecker-ci/woodpecker/pull/4193))

- For `woodpecker-cli` containers, `/woodpecker` has been set as the default `workdir`

- Secret plugin filters (in the "secrets" repo settings) can now validate against tags. Additionally, the description has been updated to reflect that these filters only apply to plugins ([#4069](https://github.com/woodpecker-ci/woodpecker/pull/4069)).
- Syntax deprecations which now result in an error:

  - `pipeline:` ([#3916](https://github.com/woodpecker-ci/woodpecker/pull/3916))
  - `platform:` ([#3916](https://github.com/woodpecker-ci/woodpecker/pull/3916))
  - `branches:` ([#3916](https://github.com/woodpecker-ci/woodpecker/pull/3916))

- Workflow syntax changes:

  - Grouping of steps via `steps.[name].group` should now be done using `steps.[name].depends_on`
  - The `includes` and `excludes` event filter options have been removed
  - Previously, env vars have been automatically sanitized to uppercase.
    As this has been confusing, the type-case of the secret definition is now respected ([#3375](https://github.com/woodpecker-ci/woodpecker/pull/3375)).
  - `secrets` have been entirely removed in favor of `environment` combined with the `from_secret` syntax.
    As `secrets` are just normal env vars which are masked, the goal was to allow them to be declared next to normal env vars and at the same time reduce the keyword syntax count.
    Additionally, the `from_secret` syntax gives more flexibility in naming.
    Whereas beforehand `secrets` where always named after their initial secret name, the `from_secret` reference can now be different.
    Last, one can inject multiple different env vars from the same secret reference.

    2.x:

    ```yaml
    secret: my_token
    ```

    3.x:

    ```yaml
    environment:
      my_token:
        from_secret: my_token
    ```

- The `environment` filter has been removed in favor of `when.evaluate`

- SDK changes:
  - The SDK fields `start_time`, `end_time`, `created_at`, `started_at`, `finished_at` and `reviewed_at` have been renamed to `started`, `finished`, `created`, `started`, `finished`, `reviewed` ([#3968](https://github.com/woodpecker-ci/woodpecker/pull/3968))
  - The `trusted` field of the repo model was changed from `boolean` to `object`
- CRON definitions now use standard Linux syntax without seconds.
  All custom CRON definitions which do not use keywords such as `@daily` or `@weekly` must be updated.

  Example definition for a CRON job running at 8 am daily:

  2.x:

  ```sh
  0 0 8 * * *
  ```

  3.x:

  ```sh
  0 8 * * *
  ```

- API changes:

  - Removed deprecated `registry/` endpoint. Use `registries`, `/authorize/token`

- CLI changes:
  The following restructuring was done to achieve a more consistent grouping:
  - `woodpecker-cli registry` → `woodpecker-cli repo registry`
  - `woodpecker-cli secret [add|rm|...] --global` → `woodpecker-cli admin secret [add|rm|...]`
  - `woodpecker-cli user` → `woodpecker-cli admin user`
  - `woodpecker-cli log-level` → `woodpecker-cli admin log-level`
  - `woodpecker-cli secret [add|rm|...] --organization` → `woodpecker-cli org secret [add|rm|...]`
  - `woodpecker-cli deploy` → `woodpecker-cli pipeline deploy`
  - `woodpecker-cli log` → `woodpecker-cli pipeline log`
  - `woodpecker-cli cron` → `woodpecker-cli repo cron`
  - `woodpecker-cli secret [add|rm|...] --repository` → `woodpecker-cli repo secret [add|rm|...]`
  - `woodpecker-cli pipeline logs` → `woodpecker-cli pipeline log show`
  - `woodpecker-cli [registry|secret|...] info` → `woodpecker-cli [registry|secret|...] show`
- Native Let's Encrypt certificate support has been dropped as it was almost unused and causing frequent issues.
  Let's Encrypt needs to be set up standalone now. The SSL key pair can still be used in `WOODPECKER_SERVER_CERT` and `WOODPECKER_SERVER_KEY` as an alternative to using a reverse proxy for TLS termination.

## Admin-facing migrations

- Previously, some (official) plugins were granted the `privileged` option by default to allow simplified usage.
  To streamline this process and enhance security transparency, no plugin is granted the `privileged` options by default anymore.
  To allow the use of these plugins in >= 3.0, they must be set explicitly through `WOODPECKER_PLUGINS_PRIVILEGED` on the admin side.
  This change mainly impacts the use of the `woodpeckerci/plugin-docker-buildx` plugin, which now will not work anymore unless explicitly whitelisted through this env var ([#4053](https://github.com/woodpecker-ci/woodpecker/pull/4053))

- Environment variable deprecations:

  - `WOODPECKER_LOG_XORM` and `WOODPECKER_LOG_XORM_SQL` got deprecated in favor of `WOODPECKER_DATABASE_LOG` and `WOODPECKER_DATABASE_LOG_SQL`
  - `WOODPECKER_FILTER_LABELS` got deprecated in favor of `WOODPECKER_AGENT_LABELS`
  - `WOODPECKER_ESCALATE` got deprecated in favor of `WOODPECKER_PLUGINS_PRIVILEGED`
  - `WOODPECKER_DEFAULT_CLONE_IMAGE` got deprecated in favor of `WOODPECKER_DEFAULT_CLONE_PLUGIN`
  - `WOODPECKER_DEV_OAUTH_HOST` and `WOODPECKER_DEV_GITEA_OAUTH_URL` got deprecated in favor of `WOODPECKER_EXPERT_FORGE_OAUTH_HOST`
  - `WOODPECKER_ROOT_PATH` and `WOODPECKER_ROOT_URL` got deprecated in favor of `WOODPECKER_HOST`

- The resource limit settings for the "docker" backend were moved from the server into agent configuration.
  This allows setting limits on an agent-level which allows greater resource definition granularity ([#3174](https://github.com/woodpecker-ci/woodpecker/pull/3174))

- "Kubernetes" backend: previously the image pull secret name was hard-coded to `regcred`.
  To allow more flexibility and specifying multiple pull secrets, the default has been removed.
  Image pull secrets must now be set explicitly via env var `WOODPECKER_BACKEND_K8S_PULL_SECRET_NAMES` ([#4005](https://github.com/woodpecker-ci/woodpecker/pull/4005))

- Webhook signatures now use the `rfc9421` protocol

- Replaced `configs` object by `netrc` in external configuration APIs

- Upgrading from 1.x versions to 3.x is disallowed, upgrading to 2.x first is required to ensure proper DB migrations

## 2.7.2

To secure your instance, set `WOODPECKER_PLUGINS_PRIVILEGED` to only allow specific versions of the `woodpeckerci/plugin-docker-buildx` plugin, use version 5.0.0 or above. This prevents older, potentially unstable versions from being privileged.

For example, to allow only version 5.0.0, use:

```bash
WOODPECKER_PLUGINS_PRIVILEGED=woodpeckerci/plugin-docker-buildx:5.0.0
```

To allow multiple versions, you can separate them with commas:

```bash
WOODPECKER_PLUGINS_PRIVILEGED=woodpeckerci/plugin-docker-buildx:5.0.0,woodpeckerci/plugin-docker-buildx:5.1.0
```

This setup ensures only specified, stable plugin versions are given privileged access.

Read more about it in [#4213](https://github.com/woodpecker-ci/woodpecker/pull/4213)

## 2.0.0

- Dropped deprecated `CI_BUILD_*`, `CI_PREV_BUILD_*`, `CI_JOB_*`, `*_LINK`, `CI_SYSTEM_ARCH`, `CI_REPO_REMOTE` built-in environment variables
- Deprecated `platform:` filter in favor of `labels:`, [read more](/docs/usage/workflow-syntax#filter-by-platform)
- Secrets `event` property was renamed to `events` and `image` to `images` as both are lists. The new property `events` / `images` has to be used in the api. The old properties `event` and `image` were removed.
- The secrets `plugin_only` option was removed. Secrets with images are now always only available for plugins using listed by the `images` property. Existing secrets with a list of `images` will now only be available to the listed images if they are used as a plugin.
- Removed `build` alias for `pipeline` command in CLI
- Removed `ssh` backend. Use an agent directly on the SSH machine using the `local` backend.
- Removed `/hook` and `/stream` API paths in favor of `/api/(hook|stream)`. You may need to use the "Repair repository" button in the repo settings or "Repair all" in the admin settings to recreate the forge hook.
- Removed `WOODPECKER_DOCS` config variable
- Renamed `link` to `url` (including all API fields)
- Deprecated `CI_COMMIT_URL` env var, use `CI_PIPELINE_FORGE_URL`

## 1.0.0

- The signature used to verify extension calls (like those used for the [config-extension](/docs/administration/advanced/external-configuration-api)) done by the Woodpecker server switched from using a shared-secret HMac to an ed25519 key-pair. Read more about it at the [config-extensions](/docs/administration/advanced/external-configuration-api) documentation.
- Refactored support for old agent filter labels and expressions. Learn how to use the new [filter](/docs/usage/workflow-syntax#labels)
- Renamed step environment variable `CI_SYSTEM_ARCH` to `CI_SYSTEM_PLATFORM`. Same applies for the cli exec variable.
- Renamed environment variables `CI_BUILD_*` and `CI_PREV_BUILD_*` to `CI_PIPELINE_*` and `CI_PREV_PIPELINE_*`, old ones are still available but deprecated
- Renamed environment variables `CI_JOB_*` to `CI_STEP_*`, old ones are still available but deprecated
- Renamed environment variable `CI_REPO_REMOTE` to `CI_REPO_CLONE_URL`, old is still available but deprecated
- Renamed environment variable `*_LINK` to `*_URL`, old ones are still available but deprecated
- Renamed API endpoints for pipelines (`<owner>/<repo>/builds/<buildId>` -> `<owner>/<repo>/pipelines/<pipelineId>`), old ones are still available but deprecated
- Updated Prometheus gauge `build_*` to `pipeline_*`
- Updated Prometheus gauge `*_job_*` to `*_step_*`
- Renamed config env `WOODPECKER_MAX_PROCS` to `WOODPECKER_MAX_WORKFLOWS` (still available as fallback) <!-- cspell:ignore PROCS -->
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

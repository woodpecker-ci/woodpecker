# Environment variables

Woodpecker provides the ability to pass environment variables to individual pipeline steps. Note that these can't overwrite any existing, built-in variables. Example pipeline step with custom environment variables:

```diff
 steps:
   - name: build
     image: golang
+    environment:
+      CGO: 0
+      GOOS: linux
+      GOARCH: amd64
     commands:
       - go build
       - go test
```

Please note that the environment section is not able to expand environment variables. If you need to expand variables they should be exported in the commands section.

```diff
 steps:
   - name: build
     image: golang
-    environment:
-      - PATH=$PATH:/go
     commands:
+      - export PATH=$PATH:/go
       - go build
       - go test
```

:::warning
`${variable}` expressions are subject to pre-processing. If you do not want the pre-processor to evaluate your expression it must be escaped:
:::

```diff
 steps:
   - name: build
     image: golang
     commands:
-      - export PATH=${PATH}:/go
+      - export PATH=$${PATH}:/go
       - go build
       - go test
```

## Built-in environment variables

This is the reference list of all environment variables available to your pipeline containers. These are injected into your pipeline step and plugins containers, at runtime.

The **Scope** column documents when each variable can be used:

- `config`: the variable is available at config-evaluation time, when the pipeline configuration is parsed. It can be referenced in [`when`](./20-workflow-syntax.md#when---conditional-execution) filters and expanded via [string substitution](#string-substitution) (e.g. `${CI_COMMIT_SHA}`).
- `runtime`: the variable is available as an environment variable inside the running step.

Most variables are available in both scopes. Variables scoped only to `runtime` (e.g. the step-specific variables and `CI_PIPELINE_STATUS`) are not populated while the configuration is evaluated, so they cannot be used in `when` filters or substitutions.

| NAME                               | Scope             | Description                                                                                                                                      | Example                                                                                                                         |
| ---------------------------------- | ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------------- |
| `CI`                               | `config, runtime` | CI environment name                                                                                                                              | `woodpecker`                                                                                                                    |
|                                    |                   | **Repository**                                                                                                                                   |                                                                                                                                 |
| `CI_REPO`                          | `config, runtime` | repository full name `<owner>/<name>`                                                                                                            | `john-doe/my-repo`                                                                                                              |
| `CI_REPO_OWNER`                    | `config, runtime` | repository owner                                                                                                                                 | `john-doe`                                                                                                                      |
| `CI_REPO_NAME`                     | `config, runtime` | repository name                                                                                                                                  | `my-repo`                                                                                                                       |
| `CI_REPO_REMOTE_ID`                | `config, runtime` | repository remote ID, is the UID it has in the forge                                                                                             | `82`                                                                                                                            |
| `CI_REPO_URL`                      | `config, runtime` | repository web URL                                                                                                                               | `https://git.example.com/john-doe/my-repo`                                                                                      |
| `CI_REPO_CLONE_URL`                | `config, runtime` | repository clone URL                                                                                                                             | `https://git.example.com/john-doe/my-repo.git`                                                                                  |
| `CI_REPO_CLONE_SSH_URL`            | `config, runtime` | repository SSH clone URL                                                                                                                         | `git@git.example.com:john-doe/my-repo.git`                                                                                      |
| `CI_REPO_DEFAULT_BRANCH`           | `config, runtime` | repository default branch                                                                                                                        | `main`                                                                                                                          |
| `CI_REPO_PRIVATE`                  | `config, runtime` | repository is private                                                                                                                            | `true`                                                                                                                          |
| `CI_REPO_TRUSTED_NETWORK`          | `config, runtime` | repository has trusted network access                                                                                                            | `false`                                                                                                                         |
| `CI_REPO_TRUSTED_VOLUMES`          | `config, runtime` | repository has trusted volumes access                                                                                                            | `false`                                                                                                                         |
| `CI_REPO_TRUSTED_SECURITY`         | `config, runtime` | repository has trusted security access                                                                                                           | `false`                                                                                                                         |
|                                    |                   | **Current Commit**                                                                                                                               |                                                                                                                                 |
| `CI_COMMIT_SHA`                    | `config, runtime` | commit SHA                                                                                                                                       | `eba09b4`                                                                                                                       |
| `CI_COMMIT_REF`                    | `config, runtime` | commit ref                                                                                                                                       | `refs/heads/main`                                                                                                               |
| `CI_COMMIT_REFSPEC`                | `config, runtime` | commit ref spec                                                                                                                                  | `issue-branch:main`                                                                                                             |
| `CI_COMMIT_BRANCH`                 | `config, runtime` | commit branch (equals target branch for pull requests)                                                                                           | `main`                                                                                                                          |
| `CI_COMMIT_SOURCE_BRANCH`          | `config, runtime` | commit source branch (set only for pull request events)                                                                                          | `issue-branch`                                                                                                                  |
| `CI_COMMIT_TARGET_BRANCH`          | `config, runtime` | commit target branch (set only for pull request events)                                                                                          | `main`                                                                                                                          |
| `CI_COMMIT_TAG`                    | `config, runtime` | commit tag name (empty if event is not `tag`)                                                                                                    | `v1.10.3`                                                                                                                       |
| `CI_COMMIT_PULL_REQUEST`           | `config, runtime` | commit pull request number (set only for pull request events)                                                                                    | `1`                                                                                                                             |
| `CI_COMMIT_PULL_REQUEST_LABELS`    | `config, runtime` | labels assigned to pull request (set only for pull request events)                                                                               | `server`                                                                                                                        |
| `CI_COMMIT_PULL_REQUEST_MILESTONE` | `config, runtime` | milestone assigned to pull request (set only for `pull_request` and `pull_request_closed` events)                                                | `summer-sprint`                                                                                                                 |
| `CI_COMMIT_PULL_REQUEST_DRAFT`     | `config, runtime` | whether the pull request is a draft (set only for pull request events; see [forge support](#ci_commit_pull_request_draft-forge-support))         | `true`, `false`                                                                                                                 |
| `CI_COMMIT_MESSAGE`                | `config, runtime` | commit message                                                                                                                                   | `Initial commit`                                                                                                                |
| `CI_COMMIT_TIMESTAMP`              | `config, runtime` | commit UNIX timestamp                                                                                                                            | `1722617519`                                                                                                                    |
| `CI_COMMIT_AUTHOR`                 | `config, runtime` | commit author username                                                                                                                           | `john-doe`                                                                                                                      |
| `CI_COMMIT_AUTHOR_EMAIL`           | `config, runtime` | commit author email address                                                                                                                      | `john-doe@example.com`                                                                                                          |
| `CI_COMMIT_PRERELEASE`             | `config, runtime` | release is a pre-release (empty if event is not `release`) — **deprecated**, use `CI_PIPELINE_RELEASE_PRE`                                       | `false`                                                                                                                         |
|                                    |                   | **Current pipeline**                                                                                                                             |                                                                                                                                 |
| `CI_PIPELINE_NUMBER`               | `config, runtime` | pipeline number                                                                                                                                  | `8`                                                                                                                             |
| `CI_PIPELINE_PARENT`               | `config, runtime` | number of parent pipeline                                                                                                                        | `0`                                                                                                                             |
| `CI_PIPELINE_STATUS`               | `runtime`         | state of the workflow right before the step was started                                                                                          | `success`, `failure`                                                                                                            |
| `CI_PIPELINE_EVENT`                | `config, runtime` | pipeline event (see [`event`](../20-usage/20-workflow-syntax.md#event))                                                                          | `push`<br/>`pull_request`<br/>`pull_request_closed`<br/>`pull_request_metadata`<br/>`tag`<br/>`release`<br/>`manual`<br/>`cron` |
| `CI_PIPELINE_EVENT_REASON`         | `config, runtime` | exact reason why `pull_request_metadata` event was send. it is forge instance specific and can change                                            | `label_updated`<br/>`milestoned`<br/>`demilestoned`<br/>`assigned`<br/>`edited`<br/>...                                         |
| `CI_PIPELINE_URL`                  | `config, runtime` | link to the web UI for the pipeline                                                                                                              | `https://ci.example.com/repos/7/pipeline/8`                                                                                     |
| `CI_PIPELINE_FORGE_URL`            | `config, runtime` | link to the forge's web UI for the commit(s) or tag that triggered the pipeline                                                                  | `https://git.example.com/john-doe/my-repo/commit/eba09b4`                                                                       |
| `CI_PIPELINE_DEPLOY_TARGET`        | `config, runtime` | pipeline deploy target for `deployment` events                                                                                                   | `production`                                                                                                                    |
| `CI_PIPELINE_DEPLOY_TASK`          | `config, runtime` | pipeline deploy task for `deployment` events                                                                                                     | `migration`                                                                                                                     |
| `CI_PIPELINE_RELEASE_TITLE`        | `config, runtime` | release title (empty if event is not `release`)                                                                                                  | `v1.10.3`                                                                                                                       |
| `CI_PIPELINE_RELEASE_PRE`          | `config, runtime` | release is a pre-release (empty if event is not `release`)                                                                                       | `false`                                                                                                                         |
| `CI_PIPELINE_CREATED`              | `config, runtime` | pipeline created UNIX timestamp                                                                                                                  | `1722617519`                                                                                                                    |
| `CI_PIPELINE_STARTED`              | `config, runtime` | pipeline started UNIX timestamp                                                                                                                  | `1722617519`                                                                                                                    |
| `CI_PIPELINE_FILES`                | `config, runtime` | changed files (empty if event is not `push` or `pull_request`), it is undefined if more than 500 files are touched                               | `[]`, `[".woodpecker.yml","README.md"]`                                                                                         |
| `CI_PIPELINE_AUTHOR`               | `config, runtime` | pipeline author username                                                                                                                         | `octocat`                                                                                                                       |
| `CI_PIPELINE_AVATAR`               | `config, runtime` | pipeline author avatar                                                                                                                           | `https://git.example.com/avatars/5dcbcadbce6f87f8abef`                                                                          |
| `CI_PIPELINE_RERUNS`               | `config, runtime` | number of times the pipeline has been restarted; not set on the initial run, `1` after the first restart, incremented on each subsequent restart | `1`                                                                                                                             |
|                                    |                   | **Current workflow**                                                                                                                             |                                                                                                                                 |
| `CI_WORKFLOW_NAME`                 | `config, runtime` | workflow name                                                                                                                                    | `release`                                                                                                                       |
|                                    |                   | **Current step**                                                                                                                                 |                                                                                                                                 |
| `CI_STEP_NAME`                     | `runtime`         | step name                                                                                                                                        | `build package`                                                                                                                 |
| `CI_STEP_TYPE`                     | `runtime`         | step type (`commands`, `plugin`, `service`, `clone` or `cache`)                                                                                  | `commands`                                                                                                                      |
| `CI_STEP_NUMBER`                   | `runtime`         | step number                                                                                                                                      | `0`                                                                                                                             |
| `CI_STEP_STARTED`                  | `runtime`         | step started UNIX timestamp                                                                                                                      | `1722617519`                                                                                                                    |
| `CI_STEP_URL`                      | `runtime`         | URL to step in UI                                                                                                                                | `https://ci.example.com/repos/7/pipeline/8`                                                                                     |
|                                    |                   | **Previous commit**                                                                                                                              |                                                                                                                                 |
| `CI_PREV_COMMIT_SHA`               | `config, runtime` | previous commit SHA                                                                                                                              | `1578411`                                                                                                                       |
| `CI_PREV_COMMIT_REF`               | `config, runtime` | previous commit ref                                                                                                                              | `refs/heads/main`                                                                                                               |
| `CI_PREV_COMMIT_REFSPEC`           | `config, runtime` | previous commit ref spec                                                                                                                         | `issue-branch:main`                                                                                                             |
| `CI_PREV_COMMIT_BRANCH`            | `config, runtime` | previous commit branch                                                                                                                           | `main`                                                                                                                          |
| `CI_PREV_COMMIT_SOURCE_BRANCH`     | `config, runtime` | previous commit source branch (set only for pull request events)                                                                                 | `issue-branch`                                                                                                                  |
| `CI_PREV_COMMIT_TARGET_BRANCH`     | `config, runtime` | previous commit target branch (set only for pull request events)                                                                                 | `main`                                                                                                                          |
| `CI_PREV_COMMIT_URL`               | `config, runtime` | previous commit link in forge                                                                                                                    | `https://git.example.com/john-doe/my-repo/commit/1578411`                                                                       |
| `CI_PREV_COMMIT_MESSAGE`           | `config, runtime` | previous commit message                                                                                                                          | `test`                                                                                                                          |
| `CI_PREV_COMMIT_TIMESTAMP`         | `config, runtime` | previous commit UNIX timestamp                                                                                                                   | `1722617519`                                                                                                                    |
| `CI_PREV_COMMIT_AUTHOR`            | `config, runtime` | previous commit author username                                                                                                                  | `john-doe`                                                                                                                      |
| `CI_PREV_COMMIT_AUTHOR_EMAIL`      | `config, runtime` | previous commit author email address                                                                                                             | `john-doe@example.com`                                                                                                          |
|                                    |                   | **Previous pipeline**                                                                                                                            |                                                                                                                                 |
| `CI_PREV_PIPELINE_NUMBER`          | `config, runtime` | previous pipeline number                                                                                                                         | `7`                                                                                                                             |
| `CI_PREV_PIPELINE_PARENT`          | `config, runtime` | previous pipeline number of parent pipeline                                                                                                      | `0`                                                                                                                             |
| `CI_PREV_PIPELINE_EVENT`           | `config, runtime` | previous pipeline event (see [`event`](../20-usage/20-workflow-syntax.md#event))                                                                 | `push`<br/>`pull_request`<br/>`pull_request_closed`<br/>`pull_request_metadata`<br/>`tag`<br/>`release`<br/>`manual`<br/>`cron` |
| `CI_PREV_PIPELINE_EVENT_REASON`    | `config, runtime` | previous exact reason `pull_request_metadata` event was send. it is forge instance specific and can change                                       | `label_updated`<br/>`milestoned`<br/>`demilestoned`<br/>`assigned`<br/>`edited`<br/>...                                         |
| `CI_PREV_PIPELINE_URL`             | `config, runtime` | previous pipeline link in CI                                                                                                                     | `https://ci.example.com/repos/7/pipeline/7`                                                                                     |
| `CI_PREV_PIPELINE_FORGE_URL`       | `config, runtime` | previous pipeline link to event in forge                                                                                                         | `https://git.example.com/john-doe/my-repo/commit/1578411`                                                                       |
| `CI_PREV_PIPELINE_DEPLOY_TARGET`   | `config, runtime` | previous pipeline deploy target for `deployment` events                                                                                          | `production`                                                                                                                    |
| `CI_PREV_PIPELINE_DEPLOY_TASK`     | `config, runtime` | previous pipeline deploy task for `deployment` events                                                                                            | `migration`                                                                                                                     |
| `CI_PREV_PIPELINE_STATUS`          | `config, runtime` | previous pipeline status                                                                                                                         | `success`, `failure`                                                                                                            |
| `CI_PREV_PIPELINE_CREATED`         | `config, runtime` | previous pipeline created UNIX timestamp                                                                                                         | `1722610173`                                                                                                                    |
| `CI_PREV_PIPELINE_STARTED`         | `config, runtime` | previous pipeline started UNIX timestamp                                                                                                         | `1722610173`                                                                                                                    |
| `CI_PREV_PIPELINE_FINISHED`        | `config, runtime` | previous pipeline finished UNIX timestamp                                                                                                        | `1722610383`                                                                                                                    |
| `CI_PREV_PIPELINE_AUTHOR`          | `config, runtime` | previous pipeline author username                                                                                                                | `octocat`                                                                                                                       |
| `CI_PREV_PIPELINE_AVATAR`          | `config, runtime` | previous pipeline author avatar                                                                                                                  | `https://git.example.com/avatars/5dcbcadbce6f87f8abef`                                                                          |
|                                    |                   | &emsp;                                                                                                                                           |                                                                                                                                 |
| `CI_WORKSPACE`                     | `runtime`         | Path of the workspace where source code gets cloned to                                                                                           | `/woodpecker/src/git.example.com/john-doe/my-repo`                                                                              |
|                                    |                   | **System**                                                                                                                                       |                                                                                                                                 |
| `CI_SYSTEM_NAME`                   | `config, runtime` | name of the CI system                                                                                                                            | `woodpecker`                                                                                                                    |
| `CI_SYSTEM_URL`                    | `config, runtime` | link to CI system                                                                                                                                | `https://ci.example.com`                                                                                                        |
| `CI_SYSTEM_HOST`                   | `config, runtime` | hostname of CI server                                                                                                                            | `ci.example.com`                                                                                                                |
| `CI_SYSTEM_VERSION`                | `config, runtime` | version of the server                                                                                                                            | `2.7.0`                                                                                                                         |
|                                    |                   | **Forge**                                                                                                                                        |                                                                                                                                 |
| `CI_FORGE_TYPE`                    | `config, runtime` | name of forge                                                                                                                                    | `bitbucket`<br/>`bitbucket_dc`<br/>`forgejo`<br/>`gitea`<br/>`github`<br/>`gitlab`                                              |
| `CI_FORGE_URL`                     | `config, runtime` | root URL of configured forge                                                                                                                     | `https://git.example.com`                                                                                                       |
|                                    |                   | **Internal** - Please don't use!                                                                                                                 |                                                                                                                                 |
| `CI_SCRIPT`                        | `runtime`         | Internal script path. Used to call pipeline step commands.                                                                                       |                                                                                                                                 |
| `CI_NETRC_USERNAME`                | `runtime`         | Credentials for private repos to be able to clone data. (Only available for specific images)                                                     |                                                                                                                                 |
| `CI_NETRC_PASSWORD`                | `runtime`         | Credentials for private repos to be able to clone data. (Only available for specific images)                                                     |                                                                                                                                 |
| `CI_NETRC_MACHINE`                 | `runtime`         | Credentials for private repos to be able to clone data. (Only available for specific images)                                                     |                                                                                                                                 |

## Global environment variables

If you want specific environment variables to be available in all of your pipelines use the `WOODPECKER_ENVIRONMENT` setting on the Woodpecker server. Note that these can't overwrite any existing, built-in variables.

```ini
WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
```

These can be used, for example, to manage the image tag used by multiple projects.

```ini
WOODPECKER_ENVIRONMENT=GOLANG_VERSION:1.18
```

```diff
 steps:
   - name: build
-    image: golang:1.18
+    image: golang:${GOLANG_VERSION}
     commands:
       - [...]
```

## String Substitution

Woodpecker provides the ability to substitute environment variables at runtime. This gives us the ability to use dynamic settings, commands and filters in our pipeline configuration.

Example commit substitution:

```diff
 steps:
   - name: s3
     image: woodpeckerci/plugin-s3
     settings:
+      target: /target/${CI_COMMIT_SHA}
```

Example tag substitution:

```diff
 steps:
   - name: s3
     image: woodpeckerci/plugin-s3
     settings:
+      target: /target/${CI_COMMIT_TAG}
```

## String Operations

Woodpecker also emulates bash string operations. This gives us the ability to manipulate the strings prior to substitution. Example use cases might include substring and stripping prefix or suffix values.

| OPERATION          | DESCRIPTION                                      |
| ------------------ | ------------------------------------------------ |
| `${param}`         | parameter substitution                           |
| `${param,}`        | parameter substitution with lowercase first char |
| `${param,,}`       | parameter substitution with lowercase            |
| `${param^}`        | parameter substitution with uppercase first char |
| `${param^^}`       | parameter substitution with uppercase            |
| `${param:pos}`     | parameter substitution with substring            |
| `${param:pos:len}` | parameter substitution with substring and length |
| `${param=default}` | parameter substitution with default              |
| `${param##prefix}` | parameter substitution with prefix removal       |
| `${param%%suffix}` | parameter substitution with suffix removal       |
| `${param/old/new}` | parameter substitution with find and replace     |

Example variable substitution with substring:

```diff
 steps:
   - name: s3
     image: woodpeckerci/plugin-s3
     settings:
+      target: /target/${CI_COMMIT_SHA:0:8}
```

Example variable substitution strips `v` prefix from `v.1.0.0`:

```diff
 steps:
   - name: s3
     image: woodpeckerci/plugin-s3
     settings:
+      target: /target/${CI_COMMIT_TAG##v}
```

## `CI_COMMIT_PULL_REQUEST_DRAFT` forge support

For pull request events, `CI_COMMIT_PULL_REQUEST_DRAFT` is set to `true` or `false` depending on whether the pull request is a draft.

| Forge                | Supported          | Notes                                                             |
| -------------------- | ------------------ | ----------------------------------------------------------------- |
| GitHub               | :white_check_mark: |                                                                   |
| Gitea                | :white_check_mark: |                                                                   |
| GitLab               | :white_check_mark: | Uses `draft`; falls back to legacy `work_in_progress` when needed |
| Forgejo              | :x:                | Webhook payloads include draft status, but it is not exposed yet  |
| Bitbucket            | :x:                | Webhook payloads include draft status, but it is not exposed yet  |
| Bitbucket Datacenter | :x:                | Webhook payloads include draft status, but it is not exposed yet  |

On unsupported forges the variable is still set to `false`.

## `pull_request_metadata` specific event reason values

For the `pull_request_metadata` event, the exact reason a metadata change was detected is passe through in `CI_PIPELINE_EVENT_REASON`.

**GitLab** merges metadata updates into one webhook. Event reasons are separated by `,` as a list.

:::note
Event reason values are forge-specific and may change between versions.
:::

| Event                | GitHub             | Gitea              | Forgejo            | GitLab             | Bitbucket | Bitbucket Datacenter | Description                                                                    |
| -------------------- | ------------------ | ------------------ | ------------------ | ------------------ | --------- | -------------------- | ------------------------------------------------------------------------------ |
| `assigned`           | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:       | :x:                  | Pull request was assigned to a user                                            |
| `converted_to_draft` | :white_check_mark: | :x:                | :x:                | :x:                | :x:       | :x:                  | Pull request was converted to a draft                                          |
| `demilestoned`       | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:       | :x:                  | Pull request was removed from a milestone                                      |
| `description_edited` | :x:                | :x:                | :x:                | :white_check_mark: | :x:       | :x:                  | Description edited                                                             |
| `edited`             | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:                | :x:       | :x:                  | The title or body of a pull request was edited, or the base branch was changed |
| `label_added`        | :x:                | :x:                | :x:                | :white_check_mark: | :x:       | :x:                  | Pull had no labels and now got label(s) added                                  |
| `label_cleared`      | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:       | :x:                  | All labels removed                                                             |
| `label_updated`      | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:       | :x:                  | New label(s) added / label(s) changed                                          |
| `locked`             | :white_check_mark: | :x:                | :x:                | :x:                | :x:       | :x:                  | Conversation on a pull request was locked                                      |
| `milestoned`         | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:       | :x:                  | Pull request was added to a milestone                                          |
| `ready_for_review`   | :white_check_mark: | :x:                | :x:                | :x:                | :x:       | :x:                  | Draft pull request was marked as ready for review                              |
| `review_requested`   | :x:                | :x:                | :x:                | :white_check_mark: | :x:       | :x:                  | New review was requested                                                       |
| `title_edited`       | :x:                | :x:                | :x:                | :white_check_mark: | :x:       | :x:                  | Title edited                                                                   |
| `unassigned`         | :white_check_mark: | :white_check_mark: | :white_check_mark: | :white_check_mark: | :x:       | :x:                  | User was unassigned from a pull request                                        |
| `unlabeled`          | :white_check_mark: | :x:                | :x:                | :x:                | :x:       | :x:                  | Label was removed from a pull request                                          |
| `unlocked`           | :white_check_mark: | :x:                | :x:                | :x:                | :x:       | :x:                  | Conversation on a pull request was unlocked                                    |

**Bitbucket** and **Bitbucket Datacenter** [are not supported at the moment](https://github.com/woodpecker-ci/woodpecker/pull/5214).

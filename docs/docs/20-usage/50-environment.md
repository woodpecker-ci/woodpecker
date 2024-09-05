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

| NAME                             | Description                                                                                                        | Example                                                                                    |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ |
| `CI`                             | CI environment name                                                                                                | `woodpecker`                                                                               |
|                                  | **Repository**                                                                                                     |                                                                                            |
| `CI_REPO`                        | repository full name `<owner>/<name>`                                                                              | `john-doe/my-repo`                                                                         |
| `CI_REPO_OWNER`                  | repository owner                                                                                                   | `john-doe`                                                                                 |
| `CI_REPO_NAME`                   | repository name                                                                                                    | `my-repo`                                                                                  |
| `CI_REPO_REMOTE_ID`              | repository remote ID, is the UID it has in the forge                                                               | `82`                                                                                       |
| `CI_REPO_SCM`                    | repository SCM                                                                                                     | `git`                                                                                      |
| `CI_REPO_URL`                    | repository web URL                                                                                                 | `https://git.example.com/john-doe/my-repo`                                                 |
| `CI_REPO_CLONE_URL`              | repository clone URL                                                                                               | `https://git.example.com/john-doe/my-repo.git`                                             |
| `CI_REPO_CLONE_SSH_URL`          | repository SSH clone URL                                                                                           | `git@git.example.com:john-doe/my-repo.git`                                                 |
| `CI_REPO_DEFAULT_BRANCH`         | repository default branch                                                                                          | `main`                                                                                     |
| `CI_REPO_PRIVATE`                | repository is private                                                                                              | `true`                                                                                     |
| `CI_REPO_TRUSTED`                | repository is trusted                                                                                              | `false`                                                                                    |
|                                  | **Current Commit**                                                                                                 |                                                                                            |
| `CI_COMMIT_SHA`                  | commit SHA                                                                                                         | `eba09b46064473a1d345da7abf28b477468e8dbd`                                                 |
| `CI_COMMIT_REF`                  | commit ref                                                                                                         | `refs/heads/main`                                                                          |
| `CI_COMMIT_REFSPEC`              | commit ref spec                                                                                                    | `issue-branch:main`                                                                        |
| `CI_COMMIT_BRANCH`               | commit branch (equals target branch for pull requests)                                                             | `main`                                                                                     |
| `CI_COMMIT_SOURCE_BRANCH`        | commit source branch (empty if event is not `pull_request` or `pull_request_closed`)                               | `issue-branch`                                                                             |
| `CI_COMMIT_TARGET_BRANCH`        | commit target branch (empty if event is not `pull_request` or `pull_request_closed`)                               | `main`                                                                                     |
| `CI_COMMIT_TAG`                  | commit tag name (empty if event is not `tag`)                                                                      | `v1.10.3`                                                                                  |
| `CI_COMMIT_PULL_REQUEST`         | commit pull request number (empty if event is not `pull_request` or `pull_request_closed`)                         | `1`                                                                                        |
| `CI_COMMIT_PULL_REQUEST_LABELS`  | labels assigned to pull request (empty if event is not `pull_request` or `pull_request_closed`)                    | `server`                                                                                   |
| `CI_COMMIT_MESSAGE`              | commit message                                                                                                     | `Initial commit`                                                                           |
| `CI_COMMIT_AUTHOR`               | commit author username                                                                                             | `john-doe`                                                                                 |
| `CI_COMMIT_AUTHOR_EMAIL`         | commit author email address                                                                                        | `john-doe@example.com`                                                                     |
| `CI_COMMIT_AUTHOR_AVATAR`        | commit author avatar                                                                                               | `https://git.example.com/avatars/5dcbcadbce6f87f8abef`                                     |
| `CI_COMMIT_PRERELEASE`           | release is a pre-release (empty if event is not `release`)                                                         | `false`                                                                                    |
|                                  | **Current pipeline**                                                                                               |                                                                                            |
| `CI_PIPELINE_NUMBER`             | pipeline number                                                                                                    | `8`                                                                                        |
| `CI_PIPELINE_PARENT`             | number of parent pipeline                                                                                          | `0`                                                                                        |
| `CI_PIPELINE_EVENT`              | pipeline event (see [`event`](../20-usage/20-workflow-syntax.md#event))                                            | `push`, `pull_request`, `pull_request_closed`, `tag`, `release`, `manual`, `cron`          |
| `CI_PIPELINE_URL`                | link to the web UI for the pipeline                                                                                | `https://ci.example.com/repos/7/pipeline/8`                                                |
| `CI_PIPELINE_FORGE_URL`          | link to the forge's web UI for the commit(s) or tag that triggered the pipeline                                    | `https://git.example.com/john-doe/my-repo/commit/eba09b46064473a1d345da7abf28b477468e8dbd` |
| `CI_PIPELINE_DEPLOY_TARGET`      | pipeline deploy target for `deployment` events                                                                     | `production`                                                                               |
| `CI_PIPELINE_DEPLOY_TASK`        | pipeline deploy task for `deployment` events                                                                       | `migration`                                                                                |
| `CI_PIPELINE_STATUS`             | pipeline status                                                                                                    | `success`, `failure`                                                                       |
| `CI_PIPELINE_CREATED`            | pipeline created UNIX timestamp                                                                                    | `1722617519`                                                                               |
| `CI_PIPELINE_STARTED`            | pipeline started UNIX timestamp                                                                                    | `1722617519`                                                                               |
| `CI_PIPELINE_FINISHED`           | pipeline finished UNIX timestamp                                                                                   | `1722617522`                                                                               |
| `CI_PIPELINE_FILES`              | changed files (empty if event is not `push` or `pull_request`), it is undefined if more than 500 files are touched | `[]`, `[".woodpecker.yml","README.md"]`                                                    |
|                                  | **Current workflow**                                                                                               |                                                                                            |
| `CI_WORKFLOW_NAME`               | workflow name                                                                                                      | `release`                                                                                  |
|                                  | **Current step**                                                                                                   |                                                                                            |
| `CI_STEP_NAME`                   | step name                                                                                                          | `build package`                                                                            |
| `CI_STEP_NUMBER`                 | step number                                                                                                        | `0`                                                                                        |
| `CI_STEP_STATUS`                 | step status                                                                                                        | `success`, `failure`                                                                       |
| `CI_STEP_STARTED`                | step started UNIX timestamp                                                                                        | `1722617519`                                                                               |
| `CI_STEP_FINISHED`               | step finished UNIX timestamp                                                                                       | `1722617522`                                                                               |
| `CI_STEP_URL`                    | URL to step in UI                                                                                                  | `https://ci.example.com/repos/7/pipeline/8`                                                |
|                                  | **Previous commit**                                                                                                |                                                                                            |
| `CI_PREV_COMMIT_SHA`             | previous commit SHA                                                                                                | `15784117e4e103f36cba75a9e29da48046eb82c4`                                                 |
| `CI_PREV_COMMIT_REF`             | previous commit ref                                                                                                | `refs/heads/main`                                                                          |
| `CI_PREV_COMMIT_REFSPEC`         | previous commit ref spec                                                                                           | `issue-branch:main`                                                                        |
| `CI_PREV_COMMIT_BRANCH`          | previous commit branch                                                                                             | `main`                                                                                     |
| `CI_PREV_COMMIT_SOURCE_BRANCH`   | previous commit source branch                                                                                      | `issue-branch`                                                                             |
| `CI_PREV_COMMIT_TARGET_BRANCH`   | previous commit target branch                                                                                      | `main`                                                                                     |
| `CI_PREV_COMMIT_URL`             | previous commit link in forge                                                                                      | `https://git.example.com/john-doe/my-repo/commit/15784117e4e103f36cba75a9e29da48046eb82c4` |
| `CI_PREV_COMMIT_MESSAGE`         | previous commit message                                                                                            | `test`                                                                                     |
| `CI_PREV_COMMIT_AUTHOR`          | previous commit author username                                                                                    | `john-doe`                                                                                 |
| `CI_PREV_COMMIT_AUTHOR_EMAIL`    | previous commit author email address                                                                               | `john-doe@example.com`                                                                     |
| `CI_PREV_COMMIT_AUTHOR_AVATAR`   | previous commit author avatar                                                                                      | `https://git.example.com/avatars/12`                                                       |
|                                  | **Previous pipeline**                                                                                              |                                                                                            |
| `CI_PREV_PIPELINE_NUMBER`        | previous pipeline number                                                                                           | `7`                                                                                        |
| `CI_PREV_PIPELINE_PARENT`        | previous pipeline number of parent pipeline                                                                        | `0`                                                                                        |
| `CI_PREV_PIPELINE_EVENT`         | previous pipeline event (see [`event`](../20-usage/20-workflow-syntax.md#event))                                   | `push`, `pull_request`, `pull_request_closed`, `tag`, `release`, `manual`, `cron`          |
| `CI_PREV_PIPELINE_URL`           | previous pipeline link in CI                                                                                       | `https://ci.example.com/repos/7/pipeline/7`                                                |
| `CI_PREV_PIPELINE_FORGE_URL`     | previous pipeline link to event in forge                                                                           | `https://git.example.com/john-doe/my-repo/commit/15784117e4e103f36cba75a9e29da48046eb82c4` |
| `CI_PREV_PIPELINE_DEPLOY_TARGET` | previous pipeline deploy target for `deployment` events                                                            | `production`                                                                               |
| `CI_PREV_PIPELINE_DEPLOY_TASK`   | previous pipeline deploy task for `deployment` events                                                              | `migration`                                                                                |
| `CI_PREV_PIPELINE_STATUS`        | previous pipeline status                                                                                           | `success`, `failure`                                                                       |
| `CI_PREV_PIPELINE_CREATED`       | previous pipeline created UNIX timestamp                                                                           | `1722610173`                                                                               |
| `CI_PREV_PIPELINE_STARTED`       | previous pipeline started UNIX timestamp                                                                           | `1722610173`                                                                               |
| `CI_PREV_PIPELINE_FINISHED`      | previous pipeline finished UNIX timestamp                                                                          | `1722610383`                                                                               |
|                                  | &emsp;                                                                                                             |                                                                                            |
| `CI_WORKSPACE`                   | Path of the workspace where source code gets cloned to                                                             | `/woodpecker/src/git.example.com/john-doe/my-repo`                                         |
|                                  | **System**                                                                                                         |                                                                                            |
| `CI_SYSTEM_NAME`                 | name of the CI system                                                                                              | `woodpecker`                                                                               |
| `CI_SYSTEM_URL`                  | link to CI system                                                                                                  | `https://ci.example.com`                                                                   |
| `CI_SYSTEM_HOST`                 | hostname of CI server                                                                                              | `ci.example.com`                                                                           |
| `CI_SYSTEM_VERSION`              | version of the server                                                                                              | `2.7.0`                                                                                    |
|                                  | **Forge**                                                                                                          |                                                                                            |
| `CI_FORGE_TYPE`                  | name of forge                                                                                                      | `bitbucket` , `bitbucket_dc` , `forgejo` , `gitea` , `github` , `gitlab`                   |
| `CI_FORGE_URL`                   | root URL of configured forge                                                                                       | `https://git.example.com`                                                                  |
|                                  | **Internal** - Please don't use!                                                                                   |                                                                                            |
| `CI_SCRIPT`                      | Internal script path. Used to call pipeline step commands.                                                         |                                                                                            |
| `CI_NETRC_USERNAME`              | Credentials for private repos to be able to clone data. (Only available for specific images)                       |                                                                                            |
| `CI_NETRC_PASSWORD`              | Credentials for private repos to be able to clone data. (Only available for specific images)                       |                                                                                            |
| `CI_NETRC_MACHINE`               | Credentials for private repos to be able to clone data. (Only available for specific images)                       |                                                                                            |

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
   - name: docker
     image: woodpeckerci/plugin-kaniko
     settings:
+      tags: ${CI_COMMIT_SHA}
```

Example tag substitution:

```diff
 steps:
   - name: docker
     image: woodpeckerci/plugin-kaniko
     settings:
+      tags: ${CI_COMMIT_TAG}
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
   - name: docker
     image: woodpeckerci/plugin-kaniko
     settings:
+      tags: ${CI_COMMIT_SHA:0:8}
```

Example variable substitution strips `v` prefix from `v.1.0.0`:

```diff
 steps:
   - name: docker
     image: woodpeckerci/plugin-kaniko
     settings:
+      tags: ${CI_COMMIT_TAG##v}
```

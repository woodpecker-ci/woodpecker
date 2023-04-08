# Environment variables

Woodpecker provides the ability to pass environment variables to individual pipeline steps. Example pipeline step with custom environment variables:

```diff
pipeline:
  build:
    image: golang
+   environment:
+     - CGO=0
+     - GOOS=linux
+     - GOARCH=amd64
    commands:
      - go build
      - go test
```

Please note that the environment section is not able to expand environment variables. If you need to expand variables they should be exported in the commands section.

```diff
pipeline:
  build:
    image: golang
-   environment:
-     - PATH=$PATH:/go
    commands:
+     - export PATH=$PATH:/go
      - go build
      - go test
```

> Please be warned that `${variable}` expressions are subject to pre-processing. If you do not want the pre-processor to evaluate your expression it must be escaped:

```diff
pipeline:
  build:
    image: golang
    commands:
-     - export PATH=${PATH}:/go
+     - export PATH=$${PATH}:/go
      - go build
      - go test
```

## Built-in environment variables

This is the reference list of all environment variables available to your pipeline containers. These are injected into your pipeline step and plugins containers, at runtime.

| NAME                             | Description                                                                                  |
| -------------------------------- | -------------------------------------------------------------------------------------------- |
| `CI=woodpecker`                  | environment is woodpecker                                                                    |
|                                  | **Repository**                                                                               |
| `CI_REPO`                        | repository full name `<owner>/<name>`                                                        |
| `CI_REPO_OWNER`                  | repository owner                                                                             |
| `CI_REPO_NAME`                   | repository name                                                                              |
| `CI_REPO_SCM`                    | repository SCM (git)                                                                         |
| `CI_REPO_LINK`                   | repository link                                                                              |
| `CI_REPO_CLONE_URL`              | repository clone URL                                                                         |
| `CI_REPO_DEFAULT_BRANCH`         | repository default branch (master)                                                           |
| `CI_REPO_PRIVATE`                | repository is private                                                                        |
| `CI_REPO_TRUSTED`                | repository is trusted                                                                        |
|                                  | **Current Commit**                                                                           |
| `CI_COMMIT_SHA`                  | commit SHA                                                                                   |
| `CI_COMMIT_REF`                  | commit ref                                                                                   |
| `CI_COMMIT_REFSPEC`              | commit ref spec                                                                              |
| `CI_COMMIT_BRANCH`               | commit branch (equals target branch for pull requests)                                       |
| `CI_COMMIT_SOURCE_BRANCH`        | commit source branch                                                                         |
| `CI_COMMIT_TARGET_BRANCH`        | commit target branch                                                                         |
| `CI_COMMIT_TAG`                  | commit tag name (empty if event is not `tag`)                                                |
| `CI_COMMIT_PULL_REQUEST`         | commit pull request number (empty if event is not `pull_request`)                            |
| `CI_COMMIT_PULL_REQUEST_LABELS`  | labels assigned to pull request (empty if event is not `pull_request`)                       |
| `CI_COMMIT_LINK`                 | commit link in forge                                                                         |
| `CI_COMMIT_MESSAGE`              | commit message                                                                               |
| `CI_COMMIT_AUTHOR`               | commit author username                                                                       |
| `CI_COMMIT_AUTHOR_EMAIL`         | commit author email address                                                                  |
| `CI_COMMIT_AUTHOR_AVATAR`        | commit author avatar                                                                         |
|                                  | **Current pipeline**                                                                         |
| `CI_PIPELINE_NUMBER`             | pipeline number                                                                              |
| `CI_PIPELINE_PARENT`             | number of parent pipeline                                                                    |
| `CI_PIPELINE_EVENT`              | pipeline event (push, pull_request, tag, deployment)                                         |
| `CI_PIPELINE_LINK`               | pipeline link in CI                                                                          |
| `CI_PIPELINE_DEPLOY_TARGET`      | pipeline deploy target for `deployment` events (ie production)                               |
| `CI_PIPELINE_STATUS`             | pipeline status (success, failure)                                                           |
| `CI_PIPELINE_CREATED`            | pipeline created UNIX timestamp                                                              |
| `CI_PIPELINE_STARTED`            | pipeline started UNIX timestamp                                                              |
| `CI_PIPELINE_FINISHED`           | pipeline finished UNIX timestamp                                                             |
|                                  | **Current workflow**                                                                         |
| `CI_WORKFLOW_NAME`               | workflow name                                                                                |
|                                  | **Current step**                                                                             |
| `CI_STEP_NAME`                   | step name                                                                                    |
| `CI_STEP_STATUS`                 | step status (success, failure)                                                               |
| `CI_STEP_STARTED`                | step started UNIX timestamp                                                                  |
| `CI_STEP_FINISHED`               | step finished UNIX timestamp                                                                 |
|                                  | **Previous commit**                                                                          |
| `CI_PREV_COMMIT_SHA`             | previous commit SHA                                                                          |
| `CI_PREV_COMMIT_REF`             | previous commit ref                                                                          |
| `CI_PREV_COMMIT_REFSPEC`         | previous commit ref spec                                                                     |
| `CI_PREV_COMMIT_BRANCH`          | previous commit branch                                                                       |
| `CI_PREV_COMMIT_SOURCE_BRANCH`   | previous commit source branch                                                                |
| `CI_PREV_COMMIT_TARGET_BRANCH`   | previous commit target branch                                                                |
| `CI_PREV_COMMIT_LINK`            | previous commit link in forge                                                                |
| `CI_PREV_COMMIT_MESSAGE`         | previous commit message                                                                      |
| `CI_PREV_COMMIT_AUTHOR`          | previous commit author username                                                              |
| `CI_PREV_COMMIT_AUTHOR_EMAIL`    | previous commit author email address                                                         |
| `CI_PREV_COMMIT_AUTHOR_AVATAR`   | previous commit author avatar                                                                |
|                                  | **Previous pipeline**                                                                        |
| `CI_PREV_PIPELINE_NUMBER`        | previous pipeline number                                                                     |
| `CI_PREV_PIPELINE_PARENT`        | previous pipeline number of parent pipeline                                                  |
| `CI_PREV_PIPELINE_EVENT`         | previous pipeline event (push, pull_request, tag, deployment)                                |
| `CI_PREV_PIPELINE_LINK`          | previous pipeline link in CI                                                                 |
| `CI_PREV_PIPELINE_DEPLOY_TARGET` | previous pipeline deploy target for `deployment` events (ie production)                      |
| `CI_PREV_PIPELINE_STATUS`        | previous pipeline status (success, failure)                                                  |
| `CI_PREV_PIPELINE_CREATED`       | previous pipeline created UNIX timestamp                                                     |
| `CI_PREV_PIPELINE_STARTED`       | previous pipeline started UNIX timestamp                                                     |
| `CI_PREV_PIPELINE_FINISHED`      | previous pipeline finished UNIX timestamp                                                    |
|                                  | &emsp;                                                                                       |
| `CI_WORKSPACE`                   | Path of the workspace where source code gets cloned to                                       |
|                                  | **System**                                                                                   |
| `CI_SYSTEM_NAME`                 | name of the CI system: `woodpecker`                                                          |
| `CI_SYSTEM_LINK`                 | link to CI system                                                                            |
| `CI_SYSTEM_HOST`                 | hostname of CI server                                                                        |
| `CI_SYSTEM_VERSION`              | version of the server                                                                        |
|                                  | **Internal** - Please don't use!                                                             |
| `CI_SCRIPT`                      | Internal script path. Used to call pipeline step commands.                                   |
| `CI_NETRC_USERNAME`              | Credentials for private repos to be able to clone data. (Only available for specific images) |
| `CI_NETRC_PASSWORD`              | Credentials for private repos to be able to clone data. (Only available for specific images) |
| `CI_NETRC_MACHINE`               | Credentials for private repos to be able to clone data. (Only available for specific images) |

## Global environment variables

If you want specific environment variables to be available in all of your pipelines use the `WOODPECKER_ENVIRONMENT` setting on the Woodpecker server. Note that these can't overwrite any existing, built-in variables.

```diff
services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
```

These can be used, for example, to manage the image tag used by multiple projects.

```diff
pipeline:
  build:
-   image: golang:1.18
+   image: golang:${GOLANG_VERSION}
    commands:
      - [...]
    environment:
      - [...]
+     - WOODPECKER_ENVIRONMENT=GOLANG_VERSION:1.18
```

## String Substitution

Woodpecker provides the ability to substitute environment variables at runtime. This gives us the ability to use dynamic settings, commands and filters in our pipeline configuration.

Example commit substitution:

```diff
pipeline:
  docker:
    image: plugins/docker
    settings:
+     tags: ${CI_COMMIT_SHA}
```

Example tag substitution:

```diff
pipeline:
  docker:
    image: plugins/docker
    settings:
+     tags: ${CI_COMMIT_TAG}
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
pipeline:
  docker:
    image: plugins/docker
    settings:
+     tags: ${CI_COMMIT_SHA:0:8}
```

Example variable substitution strips `v` prefix from `v.1.0.0`:

```diff
pipeline:
  docker:
    image: plugins/docker
    settings:
+     tags: ${CI_COMMIT_TAG##v}
```

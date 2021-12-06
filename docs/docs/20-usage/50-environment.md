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

| NAME                           | Description                                                                                  |
| ------------------------------ | -------------------------------------------------------------------------------------------- |
| `CI=woodpecker`                | environment is woodpecker                                                                    |
|                                | **Repository**                                                                               |
| `CI_REPO`                      | repository full name `<owner>/<name>`                                                        |
| `CI_REPO_OWNER`                | repository owner                                                                             |
| `CI_REPO_NAME`                 | repository name                                                                              |
| `CI_REPO_SCM`                  | repository scm (git)                                                                         |
| `CI_REPO_LINK`                 | repository link                                                                              |
| `CI_REPO_REMOTE`               | repository clone url                                                                         |
| `CI_REPO_DEFAULT_BRANCH`       | repository default branch (master)                                                           |
| `CI_REPO_PRIVATE`              | repository is private                                                                        |
| `CI_REPO_TRUSTED`              | repository is trusted                                                                        |
|                                | **Current Commit**                                                                           |
| `CI_COMMIT_SHA`                | commit sha                                                                                   |
| `CI_COMMIT_REF`                | commit ref                                                                                   |
| `CI_COMMIT_REFSPEC`            | commit ref spec                                                                              |
| `CI_COMMIT_BRANCH`             | commit branch                                                                                |
| `CI_COMMIT_SOURCE_BRANCH`      | commit source branch                                                                         |
| `CI_COMMIT_TARGET_BRANCH`      | commit target branch                                                                         |
| `CI_COMMIT_TAG`                | commit tag name (empty if event is not `tag`)                                                |
| `CI_COMMIT_PULL_REQUEST`       | commit pull request number (empty if event is not `pull_request`)                            |
| `CI_COMMIT_LINK`               | commit link in remote                                                                        |
| `CI_COMMIT_MESSAGE`            | commit message                                                                               |
| `CI_COMMIT_AUTHOR`             | commit author username                                                                       |
| `CI_COMMIT_AUTHOR_EMAIL`       | commit author email address                                                                  |
| `CI_COMMIT_AUTHOR_AVATAR`      | commit author avatar                                                                         |
|                                | **Current build**                                                                            |
| `CI_BUILD_NUMBER`              | build number                                                                                 |
| `CI_BUILD_PARENT`              | build number of parent build                                                                 |
| `CI_BUILD_EVENT`               | build event (push, pull_request, tag, deployment)                                            |
| `CI_BUILD_LINK`                | build link in ci                                                                             |
| `CI_BUILD_DEPLOY_TARGET`       | build deploy target for `deployment` events (ie production)                                  |
| `CI_BUILD_STATUS`              | build status (success, failure)                                                              |
| `CI_BUILD_CREATED`             | build created unix timestamp                                                                 |
| `CI_BUILD_STARTED`             | build started unix timestamp                                                                 |
| `CI_BUILD_FINISHED`            | build finished unix timestamp                                                                |
|                                | **Current job**                                                                              |
| `CI_JOB_NUMBER`                | job number                                                                                   |
| `CI_JOB_STATUS`                | job status (success, failure)                                                                |
| `CI_JOB_STARTED`               | job started unix timestamp                                                                   |
| `CI_JOB_FINISHED`              | job finished unix timestamp                                                                  |
|                                | **Previous commit**                                                                          |
| `CI_PREV_COMMIT_SHA`           | previous commit sha                                                                          |
| `CI_PREV_COMMIT_REF`           | previous commit ref                                                                          |
| `CI_PREV_COMMIT_REFSPEC`       | previous commit ref spec                                                                     |
| `CI_PREV_COMMIT_BRANCH`        | previous commit branch                                                                       |
| `CI_PREV_COMMIT_SOURCE_BRANCH` | previous commit source branch                                                                |
| `CI_PREV_COMMIT_TARGET_BRANCH` | previous commit target branch                                                                |
| `CI_PREV_COMMIT_LINK`          | previous commit link in remote                                                               |
| `CI_PREV_COMMIT_MESSAGE`       | previous commit message                                                                      |
| `CI_PREV_COMMIT_AUTHOR`        | previous commit author username                                                              |
| `CI_PREV_COMMIT_AUTHOR_EMAIL`  | previous commit author email address                                                         |
| `CI_PREV_COMMIT_AUTHOR_AVATAR` | previous commit author avatar                                                                |
|                                | **Previous build**                                                                           |
| `CI_PREV_BUILD_NUMBER`         | previous build number                                                                        |
| `CI_PREV_BUILD_PARENT`         | previous build number of parent build                                                        |
| `CI_PREV_BUILD_EVENT`          | previous build event (push, pull_request, tag, deployment)                                   |
| `CI_PREV_BUILD_LINK`           | previous build link in ci                                                                    |
| `CI_PREV_BUILD_DEPLOY_TARGET`  | previous build deploy target for `deployment` events (ie production)                         |
| `CI_PREV_BUILD_STATUS`         | previous build status (success, failure)                                                     |
| `CI_PREV_BUILD_CREATED`        | previous build created unix timestamp                                                        |
| `CI_PREV_BUILD_STARTED`        | previous build started unix timestamp                                                        |
| `CI_PREV_BUILD_FINISHED`       | previous build finished unix timestamp                                                       |
|                                | &emsp;                                                                                             |
| `CI_WORKSPACE`                 | Path of the workspace where source code gets cloned to                                       |
|                                | **System**                                                                                   |
| `CI_SYSTEM_NAME`               | name of the ci system: `woodpecker`                                                          |
| `CI_SYSTEM_LINK`               | link to ci system                                                                            |
| `CI_SYSTEM_HOST`               | hostname of ci server                                                                        |
| `CI_SYSTEM_VERSION`            | version of the server                                                                        |
|                                | **Internal** - Please don't use!                                                               |
| `CI_SCRIPT`                    | Internal script path. Used to call pipeline step commands.                                   |
| `CI_NETRC_USERNAME`            | Credentials for private repos to be able to clone data. (Only available for specific images) |
| `CI_NETRC_PASSWORD`            | Credentials for private repos to be able to clone data. (Only available for specific images) |
| `CI_NETRC_MACHINE`             | Credentials for private repos to be able to clone data. (Only available for specific images) |

## Global environment variables

If you want specific environment variables to be available in all of your builds use the `WOODPECKER_ENVIRONMENT` setting on the Woodpecker server.

```.diff
services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
```

## String Substitution

Woodpecker provides the ability to substitute environment variables at runtime. This gives us the ability to use dynamic build or commit details in our pipeline configuration.

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

| OPERATION          | DESC                                             |
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

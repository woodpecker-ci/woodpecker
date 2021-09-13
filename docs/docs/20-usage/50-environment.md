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

This is the reference list of all environment variables available to your build environment. These are injected into your build and plugins containers, at runtime.

| NAME                         | DESC                                   |
| ---------------------------- | -------------------------------------- |
| `CI=drone`                   | environment is drone                   |
| `DRONE=true`                 | environment is drone                   |
| `DRONE_ARCH`                 | environment architecture (linux/amd64) |
| `DRONE_REPO`                 | repository full name                   |
| `DRONE_REPO_OWNER`           | repository owner                       |
| `DRONE_REPO_NAME`            | repository name                        |
| `DRONE_REPO_SCM`             | repository scm (git)                   |
| `DRONE_REPO_LINK`            | repository link                        |
| `DRONE_REPO_AVATAR`          | repository avatar                      |
| `DRONE_REPO_BRANCH`          | repository default branch (master)     |
| `DRONE_REPO_PRIVATE`         | repository is private                  |
| `DRONE_REPO_TRUSTED`         | repository is trusted                  |
| `DRONE_REMOTE_URL`           | repository clone url                   |
| `DRONE_COMMIT_SHA`           | commit sha                             |
| `DRONE_COMMIT_REF`           | commit ref                             |
| `DRONE_COMMIT_BRANCH`        | commit branch                          |
| `DRONE_COMMIT_LINK`          | commit link in remote                  |
| `DRONE_COMMIT_MESSAGE`       | commit message                         |
| `DRONE_COMMIT_AUTHOR`        | commit author username                 |
| `DRONE_COMMIT_AUTHOR_EMAIL`  | commit author email address            |
| `DRONE_COMMIT_AUTHOR_AVATAR` | commit author avatar                   |
| `DRONE_BUILD_NUMBER`         | build number                           |
| `DRONE_BUILD_EVENT`          | build event (push, pull_request, tag)  |
| `DRONE_BUILD_STATUS`         | build status (success, failure)        |
| `DRONE_BUILD_LINK`           | build result link                      |
| `DRONE_BUILD_CREATED`        | build created unix timestamp           |
| `DRONE_BUILD_STARTED`        | build started unix timestamp           |
| `DRONE_BUILD_FINISHED`       | build finished unix timestamp          |
| `DRONE_PREV_BUILD_STATUS`    | prior build status                     |
| `DRONE_PREV_BUILD_NUMBER`    | prior build number                     |
| `DRONE_PREV_COMMIT_SHA`      | prior build commit sha                 |
| `DRONE_JOB_NUMBER`           | job number                             |
| `DRONE_JOB_STATUS`           | job status                             |
| `DRONE_JOB_STARTED`          | job started                            |
| `DRONE_JOB_FINISHED`         | job finished                           |
| `DRONE_BRANCH`               | commit branch                          |
| `DRONE_TARGET_BRANCH`        | The target branch of a Pull Request    |
| `DRONE_SOURCE_BRANCH`        | The source branch of a Pull Request    |
| `DRONE_COMMIT`               | commit sha                             |
| `DRONE_TAG`                  | commit tag                             |
| `DRONE_PULL_REQUEST`         | pull request number                    |
| `DRONE_DEPLOY_TO`            | deployment target (ie production)      |

## Global environment variables

If you want specific environment variables to be available in all of your builds use the `WOODPECKER_ENVIRONMENT` setting on the Woodpecker server.

```.env
WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
```

```.diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_ORGS=dolores,dogpatch
      - WOODPECKER_ADMIN=johnsmith,janedoe
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
+     - WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
```

## String Substitution

Woodpecker provides the ability to substitute environment variables at runtime. This gives us the ability to use dynamic build or commit details in our pipeline configuration.

Example commit substitution:

```diff
pipeline:
  docker:
    image: plugins/docker
+   tags: ${DRONE_COMMIT_SHA}
```

Example tag substitution:

```diff
pipeline:
  docker:
    image: plugins/docker
+   tags: ${DRONE_TAG}
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
+   tags: ${DRONE_COMMIT_SHA:0:8}
```

Example variable substitution strips `v` prefix from `v.1.0.0`:

```diff
pipeline:
  docker:
    image: plugins/docker
+   tags: ${DRONE_TAG##v}
```

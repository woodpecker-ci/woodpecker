<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->

**Table of Contents** _generated with [DocToc](https://github.com/thlorenz/doctoc)_

- [Pipeline basics](#pipeline-basics)
  - [Activation](#activation)
  - [Configuration](#configuration)
  - [Execution](#execution)
- [Pipelines](#pipelines)
  - [Build Steps](#build-steps)
  - [Images](#images)
    - [Images from private registries](#images-from-private-registries)
    - [GCR Registry Support](#gcr-registry-support)
  - [Parallel Execution](#parallel-execution)
  - [Conditional Pipeline Execution](#conditional-pipeline-execution)
  - [Conditional Step Execution](#conditional-step-execution)
    - [Failure Execution](#failure-execution)
- [Services](#services)
  - [Configuration](#configuration-1)
  - [Detachment](#detachment)
  - [Initialization](#initialization)
- [Plugins](#plugins)
  - [Plugin Isolation](#plugin-isolation)
  - [Plugin Marketplace](#plugin-marketplace)
- [Environment variables](#environment-variables)
  - [Built-in environment variables](#built-in-environment-variables)
  - [String Substitution](#string-substitution)
  - [String Operations](#string-operations)
- [Secrets](#secrets)
  - [Adding Secrets](#adding-secrets)
  - [Alternate Names](#alternate-names)
  - [Pull Requests](#pull-requests)
  - [Examples](#examples)
- [Volumes](#volumes)
- [Webhooks](#webhooks)
  - [Required Permissions](#required-permissions)
  - [Skip Commits](#skip-commits)
  - [Skip Branches](#skip-branches)
- [Workspace](#workspace)
- [Cloning](#cloning)
  - [Git Submodules](#git-submodules)
- [Privileged mode](#privileged-mode)
- [Promoting](#promoting)
  - [Triggering Deployments](#triggering-deployments)
- [Matrix builds](#matrix-builds)
  - [Interpolation](#interpolation)
  - [Examples](#examples-1)
- [Multi-pipeline builds](#multi-pipeline-builds)
  - [Example multi-pipeline definition](#example-multi-pipeline-definition)
  - [Flow control](#flow-control)
  - [Status lines](#status-lines)
  - [Rational](#rational)
- [Badges](#badges)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

This document explains the process for activating and configuring a continuous delivery pipeline.

# Pipeline basics

## Activation

To activate your project navigate to your account settings. You will see a list of repositories which can be activated with a simple toggle. When you activate your repository, Drone automatically adds webhooks to your version control system (e.g. GitHub).

Webhooks are used to trigger pipeline executions. When you push code to your repository, open a pull request, or create a tag, your version control system will automatically send a webhook to Drone which will in turn trigger pipeline execution.

![repository list](`repo_list.png)

## Configuration

To configure you pipeline you should place a `.drone.yml` file in the root of your repository. The .drone.yml file is used to define your pipeline steps. It is a superset of the widely used docker-compose file format.

Example pipeline configuration:

```yaml
pipeline:
  build:
    image: golang
    commands:
      - go get
      - go build
      - go test

services:
  postgres:
    image: postgres:9.4.5
    environment:
      - POSTGRES_USER=myapp
```

Example pipeline configuration with multiple, serial steps:

```yaml
pipeline:
  backend:
    image: golang
    commands:
      - go get
      - go build
      - go test

  frontend:
    image: node:6
    commands:
      - npm install
      - npm test

  notify:
    image: plugins/slack
    channel: developers
    username: drone
```

## Execution

To trigger your first pipeline execution you can push code to your repository, open a pull request, or push a tag. Any of these events triggers a webhook from your version control system and execute your pipeline.

# Pipelines

The pipeline section defines a list of steps to build, test and deploy your code. Pipeline steps are executed serially, in the order in which they are defined. If a step returns a non-zero exit code, the pipeline immediately aborts and returns a failure status.

Example pipeline:

```yaml
pipeline:
  backend:
    image: golang
    commands:
      - go build
      - go test
  frontend:
    image: node
    commands:
      - npm install
      - npm run test
      - npm run build
```

In the above example we define two pipeline steps, `frontend` and `backend`. The names of these steps are completely arbitrary.

## Build Steps

Build steps are steps in your pipeline that execute arbitrary commands inside the specified docker container. The commands are executed using the workspace as the working directory.

```diff
pipeline:
  backend:
    image: golang
    commands:
+     - go build
+     - go test
```

There is no magic here. The above commands are converted to a simple shell script. The commands in the above example are roughly converted to the below script:

```diff
#!/bin/sh
set -e

go build
go test
```

The above shell script is then executed as the docker entrypoint. The below docker command is an (incomplete) example of how the script is executed:

```
docker run --entrypoint=build.sh golang
```

> Please note that only build steps can define commands. You cannot use commands with plugins or services.

## Images

Drone uses Docker images for the build environment, for plugins and for service containers. The image field is exposed in the container blocks in the Yaml:

```diff
pipeline:
  build:
+   image: golang:1.6
    commands:
      - go build
      - go test

  publish:
+   image: plugins/docker
    repo: foo/bar

services:
  database:
+   image: mysql
```

Drone supports any valid Docker image from any Docker registry:

```text
image: golang
image: golang:1.7
image: library/golang:1.7
image: index.docker.io/library/golang
image: index.docker.io/library/golang:1.7
```

Drone does not automatically upgrade docker images. Example configuration to always pull the latest image when updates are available:

```diff
pipeline:
  build:
    image: golang:latest
+   pull: true
```

#### Images from private registries

You must provide registry credentials on the UI in order to pull private pipeline images defined in your Yaml configuration file.

These credentials are never exposed to your pipeline, which means they cannot be used to push, and are safe to use with pull requests, for example. Pushing to a registry still require setting credentials for the appropriate plugin.

Example configuration using a private image:

```diff
pipeline:
  build:
+   image: gcr.io/custom/golang
    commands:
      - go build
      - go test
```

Drone matches the registry hostname to each image in your yaml. If the hostnames match, the registry credentials are used to authenticate to your registry and pull the image. Note that registry credentials are used by the Drone agent and are never exposed to your build containers.

Example registry hostnames:

- Image `gcr.io/foo/bar` has hostname `gcr.io`
- Image `foo/bar` has hostname `docker.io`
- Image `qux.com:8000/foo/bar` has hostname `qux.com:8000`

Example registry hostname matching logic:

- Hostname `gcr.io` matches image `gcr.io/foo/bar`
- Hostname `docker.io` matches `golang`
- Hostname `docker.io` matches `library/golang`
- Hostname `docker.io` matches `bradyrydzewski/golang`
- Hostname `docker.io` matches `bradyrydzewski/golang:latest`

#### GCR Registry Support

For specific details on configuring access to Google Container Registry, please view the docs [here](https://cloud.google.com/container-registry/docs/advanced-authentication#using_a_json_key_file).

## Parallel Execution

Drone supports parallel step execution for same-machine fan-in and fan-out. Parallel steps are configured using the `group` attribute. This instructs the pipeline runner to execute the named group in parallel.

Example parallel configuration:

```diff
pipeline:
  backend:
+   group: build
    image: golang
    commands:
      - go build
      - go test
  frontend:
+   group: build
    image: node
    commands:
      - npm install
      - npm run test
      - npm run build
  publish:
    image: plugins/docker
    repo: octocat/hello-world
```

In the above example, the `frontend` and `backend` steps are executed in parallel. The pipeline runner will not execute the `publish` step until the group completes.

## Conditional Pipeline Execution

Drone supports defining conditional pipelines to skip commits based on the target branch. If the branch matches the `branches:` block the pipeline is executed, otherwise it is skipped.

Example skipping a commit when the target branch is not master:

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

+branches: master
```

Example matching multiple target branches:

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

+branches: [ master, develop ]
```

Example uses glob matching:

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

+branches: [ master, feature/* ]
```

Example includes branches:

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

+branches:
+  include: [ master, feature/* ]
```

Example excludes branches:

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

+branches:
+  exclude: [ develop, feature/* ]
```

## Conditional Step Execution

Drone supports defining conditional pipeline steps in the `when` block. If all conditions in the `when` block evaluate to true the step is executed, otherwise it is skipped.

Example conditional execution by branch:

```diff
pipeline:
  slack:
    image: plugins/slack
    channel: dev
+   when:
+     branch: master
```

> The step now triggers on master, but also if the target branch of a pull request is `master`. Add an event condition to limit it further to pushes on master only.

Execute a step if the branch is `master` or `develop`:

```diff
when:
  branch: [master, develop]
```

Execute a step if the branch starts with `prefix/*`:

```diff
when:
  branch: prefix/*
```

Execute a step using custom include and exclude logic:

```diff
when:
  branch:
    include: [ master, release/* ]
    exclude: [ release/1.0.0, release/1.1.* ]
```

Execute a step if the build event is a `tag`:

```diff
when:
  event: tag
```

Execute a step if the build event is a `tag` created from the specified branch:

```diff
when:
  event: tag
+ branch: master
```

Execute a step for all non-pull request events:

```diff
when:
  event: [push, tag, deployment]
```

Execute a step for all build events:

```diff
when:
  event: [push, pull_request, tag, deployment]
```

Execute a step if the tag name starts with `release`:

```diff
when:
  tag: release*
```

Execute a step when the build status changes:

```diff
when:
  status: changed
```

Execute a step when the build is passing or failing:

```diff
when:
  status:  [ failure, success ]
```

Execute a step for a specific platform:

```diff
when:
  platform: linux/amd64
```

Execute a step for a specific platform using wildcards:

```diff
when:
  platform:  [ linux/*, windows/amd64 ]
```

Execute a step for deployment events matching the target deployment environment:

```diff
when:
  environment: production
  event: deployment
```

Execute a step for a single matrix permutation:

```diff
when:
  matrix:
    GO_VERSION: 1.5
    REDIS_VERSION: 2.8
```

Execute a step only on a certain Drone instance:

```diff
when:
  instance: stage.drone.company.com
```

#### Failure Execution

Drone uses the container exit code to determine the success or failure status of a build. Non-zero exit codes fail the build and cause the pipeline to immediately exit.

There are use cases for executing pipeline steps on failure, such as sending notifications for failed builds. Use the status constraint to override the default behavior and execute steps even when the build status is failure:

```diff
pipeline:
  slack:
    image: plugins/slack
    channel: dev
+   when:
+     status: [ success, failure ]
```

# Services

Drone provides a services section in the Yaml file used for defining service containers. The below configuration composes database and cache containers.

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

services:
  database:
    image: mysql

  cache:
    image: redis
```

Services are accessed using custom hostnames. In the above example the mysql service is assigned the hostname `database` and is available at `database:3306`.

## Configuration

Service containers generally expose environment variables to customize service startup such as default usernames, passwords and ports. Please see the official image documentation to learn more.

```diff
services:
  database:
    image: mysql
+   environment:
+     - MYSQL_DATABASE=test
+     - MYSQL_ALLOW_EMPTY_PASSWORD=yes

  cache:
    image: redis
```

## Detachment

Service and long running containers can also be included in the pipeline section of the configuration using the detach parameter without blocking other steps. This should be used when explicit control over startup order is required.

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

  database:
    image: redis
+   detach: true

  test:
    image: golang
    commands:
      - go test
```

Containers from detached steps will terminate when the pipeline ends.

## Initialization

Service containers require time to initialize and begin to accept connections. If you are unable to connect to a service you may need to wait a few seconds or implement a backoff.

```diff
pipeline:
  test:
    image: golang
    commands:
+     - sleep 15
      - go get
      - go test

services:
  database:
    image: mysql
```

# Plugins

Plugins are Docker containers that perform pre-defined tasks and are configured as steps in your pipeline. Plugins can be used to deploy code, publish artifacts, send notification, and more.

Example pipeline using the Docker and Slack plugins:

```yaml
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

  publish:
    image: plugins/docker
    repo: foo/bar
    tags: latest

  notify:
    image: plugins/slack
    channel: dev
```

## Plugin Isolation

Plugins are executed in Docker containers and are isolated from the other steps in your build pipeline. Plugins do share the build workspace, mounted as a volume, and therefore have access to your source tree.

## Plugin Marketplace

Plugins are packaged and distributed as Docker containers. They are conceptually similar to software libraries (think npm) and can be published and shared with the community. You can find a list of available plugins at [http://plugins.drone.io](http://plugins.drone.io).

# Environment variables

Drone provides the ability to define environment variables scoped to individual build steps. Example pipeline step with custom environment variables:

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

Please be warned that `${variable}` expressions are subject to pre-processing. If you do not want the pre-processor to evaluate your expression it must be escaped:

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
| `DRONE_COMMIT`               | commit sha                             |
| `DRONE_TAG`                  | commit tag                             |
| `DRONE_PULL_REQUEST`         | pull request number                    |
| `DRONE_DEPLOY_TO`            | deployment target (ie production)      |

## String Substitution

Drone provides the ability to substitute environment variables at runtime. This gives us the ability to use dynamic build or commit details in our pipeline configuration.

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

Drone also emulates bash string operations. This gives us the ability to manipulate the strings prior to substitution. Example use cases might include substring and stripping prefix or suffix values.

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

# Secrets

Drone provides the ability to store named parameters external to the Yaml configuration file, in a central secret store. Individual steps in the yaml can request access to these named parameters at runtime.

Secrets are exposed to your pipeline steps and plugins as uppercase environment variables and can therefore be referenced in the commands section of your pipeline.

```diff
pipeline:
  docker:
    image: docker
    commands:
+     - echo $DOCKER_USERNAME
+     - echo $DOCKER_PASSWORD
    secrets: [ docker_username, docker_password ]
```

Please note parameter expressions are subject to pre-processing. When using secrets in parameter expressions they should be escaped.

```diff
pipeline:
  docker:
    image: docker
    commands:
-     - echo ${DOCKER_USERNAME}
-     - echo ${DOCKER_PASSWORD}
+     - echo $${DOCKER_USERNAME}
+     - echo $${DOCKER_PASSWORD}
    secrets: [ docker_username, docker_password ]
```

## Adding Secrets

Secrets are added to the Drone secret store on the UI or with the CLI.

## Alternate Names

There may be scenarios where you are required to store secrets using alternate names. You can map the alternate secret name to the expected name using the below syntax:

```diff
pipeline:
  docker:
    image: plugins/docker
    repo: octocat/hello-world
    tags: latest
+   secrets:
+     - source: docker_prod_password
+       target: docker_password
```

## Pull Requests

Secrets are not exposed to pull requests by default. You can override this behavior by creating the secret and enabling the `pull_request` event type.

```diff
drone secret add \
  -repository octocat/hello-world \
  -image plugins/docker \
+ -event pull_request \
+ -event push \
+ -event tag \
  -name docker_username \
  -value <value>
```

Please be careful when exposing secrets to pull requests. If your repository is open source and accepts pull requests your secrets are not safe. A bad actor can submit a malicious pull request that exposes your secrets.

## Examples

Create the secret using default settings. The secret will be available to all images in your pipeline, and will be available to all push, tag, and deployment events (not pull request events).

```diff
drone secret add \
  -repository octocat/hello-world \
  -name aws_access_key_id \
  -value <value>
```

Create the secret and limit to a single image:

```diff
drone secret add \
  -repository octocat/hello-world \
+ -image plugins/s3 \
  -name aws_access_key_id \
  -value <value>
```

Create the secrets and limit to a set of images:

```diff
drone secret add \
  -repository octocat/hello-world \
+ -image plugins/s3 \
+ -image peloton/drone-ecs \
  -name aws_access_key_id \
  -value <value>
```

Create the secret and enable for multiple hook events:

```diff
drone secret add \
  -repository octocat/hello-world \
  -image plugins/s3 \
+ -event pull_request \
+ -event push \
+ -event tag \
  -name aws_access_key_id \
  -value <value>
```

Loading secrets from file using curl `@` syntax. This is the recommended approach for loading secrets from file to preserve newlines:

```diff
drone secret add \
  -repository octocat/hello-world \
  -name ssh_key \
+ -value @/root/ssh/id_rsa
```

# Volumes

Drone gives the ability to define Docker volumes in the Yaml. You can use this parameter to mount files or folders on the host machine into your containers.

> Volumes are only available to trusted repositories and for security reasons should only be used in private environments.

```diff
pipeline:
  build:
    image: docker
    commands:
      - docker build --rm -t octocat/hello-world .
      - docker run --rm octocat/hello-world --test
      - docker push octocat/hello-world
      - docker rmi octocat/hello-world
    volumes:
+     - /var/run/docker.sock:/var/run/docker.sock
```

Please note that Drone mounts volumes on the host machine. This means you must use absolute paths when you configure volumes. Attempting to use relative paths will result in an error.

```diff
- volumes: [ ./certs:/etc/ssl/certs ]
+ volumes: [ /etc/ssl/certs:/etc/ssl/certs ]
```

# Webhooks

When you activate your repository Drone automatically add webhooks to your version control system (e.g. GitHub). There is no manual configuration required.

Webhooks are used to trigger pipeline executions. When you push code to your repository, open a pull request, or create a tag, your version control system will automatically send a webhook to Drone which will in turn trigger pipeline execution.

## Required Permissions

The user who enables a repo in Drone must have `Admin` rights on that repo, so that Drone can add the webhook.

Note that manually creating webhooks yourself is not possible. This is because webhooks are signed using a per-repository secret key which is not exposed to end users.

## Skip Commits

Drone gives the ability to skip individual commits by adding `[CI SKIP]` to the commit message. Note this is case-insensitive.

```diff
git commit -m "updated README [CI SKIP]"
```

## Skip Branches

Drone gives the ability to skip commits based on the target branch. The below example will skip a commit when the target branch is not master.

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

+branches: master
```

Please see the pipeline conditions [documentation]({{< ref "usage/config/pipeline-conditions.md" >}}) for more options and details.

# Workspace

The workspace defines the shared volume and working directory shared by all pipeline steps. The default workspace matches the below pattern, based on your repository url.

```
/drone/src/github.com/octocat/hello-world
```

The workspace can be customized using the workspace block in the Yaml file:

```diff
+workspace:
+  base: /go
+  path: src/github.com/octocat/hello-world

pipeline:
  build:
    image: golang:latest
    commands:
      - go get
      - go test
```

The base attribute defines a shared base volume available to all pipeline steps. This ensures your source code, dependencies and compiled binaries are persisted and shared between steps.

```diff
workspace:
+ base: /go
  path: src/github.com/octocat/hello-world

pipeline:
  deps:
    image: golang:latest
    commands:
      - go get
      - go test
  build:
    image: node:latest
    commands:
      - go build
```

This would be equivalent to the following docker commands:

```
docker volume create my-named-volume

docker run --volume=my-named-volume:/go golang:latest
docker run --volume=my-named-volume:/go node:latest
```

The path attribute defines the working directory of your build. This is where your code is cloned and will be the default working directory of every step in your build process. The path must be relative and is combined with your base path.

```diff
workspace:
  base: /go
+ path: src/github.com/octocat/hello-world
```

```text
git clone https://github.com/octocat/hello-world \
  /go/src/github.com/octocat/hello-world
```

# Cloning

Drone automatically configures a default clone step if not explicitly defined. You can manually configure the clone step in your pipeline for customization:

```diff
+clone:
+  git:
+    image: plugins/git

pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test
```

Example configuration to override depth:

```diff
clone:
  git:
    image: plugins/git
+   depth: 50
```

Example configuration to use a custom clone plugin:

```diff
clone:
  git:
+   image: octocat/custom-git-plugin
```

Example configuration to clone Mercurial repository:

```diff
clone:
  hg:
+   image: plugins/hg
+   path: bitbucket.org/foo/bar
```

## Git Submodules

To use the credentials that cloned the repository to clone it's submodules, update `.gitmodules` to use `https` instead of `git`:

```diff
[submodule "my-module"]
	path = my-module
-	url = git@github.com:octocat/my-module.git
+	url = https://github.com/octocat/my-module.git
```

To use the ssh git url in `.gitmodules` for users cloning with ssh, and also use the https url in drone, add `submodule_override`:

```diff
clone:
  git:
    image: plugins/git
    recursive: true
+   submodule_override:
+     my-module: https://github.com/octocat/my-module.git

pipeline:
  ...
```

# Privileged mode

Drone gives the ability to configure privileged mode in the Yaml. You can use this parameter to launch containers with escalated capabilities.

> Privileged mode is only available to trusted repositories and for security reasons should only be used in private environments.

```diff
pipeline:
  build:
    image: docker
    environment:
      - DOCKER_HOST=tcp://docker:2375
    commands:
      - docker --tls=false ps

services:
  docker:
    image: docker:dind
    command: [ "--storage-driver=vfs", "--tls=false" ]
+   privileged: true
```

# Promoting

Drone provides the ability to promote individual commits or tags (e.g. promote to production). When you promote a commit or tag it triggers a new pipeline execution with event type `deployment`. You can use the event type and target environment to limit step execution.

```diff
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

  publish:
    image: plugins/docker
    registry: registry.heroku.com
    repo: registry.heroku.com/my-staging-app/web
    when:
+     event: deployment
+     environment: staging

  publish_to_prod:
    image: plugins/docker
    registry: registry.heroku.com
    repo: registry.heroku.com/my-production-app/web
    when:
+     event: deployment
+     environment: production
```

The above example demonstrates how we can configure pipeline steps to only execute when the deployment matches a specific target environment.

## Triggering Deployments

Deployments are triggered from the command line utility. They are triggered from an existing build. This is conceptually similar to promoting builds.

```text
drone deploy <repo> <build> <environment>
```

Promote the specified build number to your staging environment:

```text
drone deploy octocat/hello-world 24 staging
```

Promote the specified build number to your production environment:

```text
drone deploy octocat/hello-world 24 production
```

# Matrix builds

Drone has integrated support for matrix builds. Drone executes a separate build task for each combination in the matrix, allowing you to build and test a single commit against multiple configurations.

Example matrix definition:

```yaml
matrix:
  GO_VERSION:
    - 1.4
    - 1.3
  REDIS_VERSION:
    - 2.6
    - 2.8
    - 3.0
```

Example matrix definition containing only specific combinations:

```yaml
matrix:
  include:
    - GO_VERSION: 1.4
      REDIS_VERSION: 2.8
    - GO_VERSION: 1.5
      REDIS_VERSION: 2.8
    - GO_VERSION: 1.6
      REDIS_VERSION: 3.0
```

## Interpolation

Matrix variables are interpolated in the yaml using the `${VARIABLE}` syntax, before the yaml is parsed. This is an example yaml file before interpolating matrix parameters:

```yaml
pipeline:
  build:
    image: golang:${GO_VERSION}
    commands:
      - go get
      - go build
      - go test

services:
  database:
    image: ${DATABASE}

matrix:
  GO_VERSION:
    - 1.4
    - 1.3
  DATABASE:
    - mysql:5.5
    - mysql:6.5
    - mariadb:10.1
```

Example Yaml file after injecting the matrix parameters:

```diff
pipeline:
  build:
-   image: golang:${GO_VERSION}
+   image: golang:1.4
    commands:
      - go get
      - go build
      - go test
+   environment:
+     - GO_VERSION=1.4
+     - DATABASE=mysql:5.5

services:
  database:
-   image: ${DATABASE}
+   image: mysql:5.5
```

## Examples

Example matrix build based on Docker image tag:

```yaml
pipeline:
  build:
    image: golang:${TAG}
    commands:
      - go build
      - go test

matrix:
  TAG:
    - 1.7
    - 1.8
    - latest
```

Example matrix build based on Docker image:

```yaml
pipeline:
  build:
    image: ${IMAGE}
    commands:
      - go build
      - go test

matrix:
  IMAGE:
    - golang:1.7
    - golang:1.8
    - golang:latest
```

# Multi-pipeline builds

By default, Drone looks for the pipeline definition in `.drone.yml` in the project root.

The Multi-Pipeline feature allows the pipeline to be splitted to several files and placed in the `.drone/` folder

## Example multi-pipeline definition

```bash
.drone
├── .build.yml
├── .deploy.yml
├── .lint.yml
└── .test.yml
```

.drone/.build.yml

```yaml
pipeline:
  build:
    image: debian:stable-slim
    commands:
      - echo building
      - sleep 5
```

.drone/.deploy.yml

```yaml
pipeline:
  deploy:
    image: debian:stable-slim
    commands:
      - echo deploying

depends_on:
  - lint
  - build
  - test
```

.drone/.test.yml

```yaml
pipeline:
  test:
    image: debian:stable-slim
    commands:
      - echo testing
      - sleep 5

depends_on:
  - build
```

.drone/.lint.yml

```yaml
pipeline:
  lint:
    image: debian:stable-slim
    commands:
      - echo linting
      - sleep 5
```

## Flow control

The pipelines run in parallel on a separate agents and share nothing.

Dependencies between pipelines can be set with the `depends_on` element. A pipeline doesn't execute until its dependencies did not complete succesfully.

```diff
pipeline:
  deploy:
    image: debian:stable-slim
    commands:
      - echo deploying

+depends_on:
+  - lint
+  - build
+  - test
```

Pipelines that need to run even on failures should set the `run_on` tag.

```diff
pipeline:
  notify:
    image: debian:stable-slim
    commands:
      - echo notifying

depends_on:
  - deploy

+run_on: [ success, failure ]
```

Some pipelines don't need the source code, set the `skip_clone` tag to skip cloning:

```diff

pipeline:
  notify:
    image: debian:stable-slim
    commands:
      - echo notifying

depends_on:
  - deploy

run_on: [ success, failure ]
+skip_clone: true
```

## Status lines

Each pipeline has its own status line on Github.

## Rational

- faster lint/test feedback, the pipeline doesn't have to run fully to have a lint status pushed to the the remote
- better organization of the pipeline along various concerns: testing, linting, feature apps
- utilizaing more agents to speed up build

# Badges

Drone has integrated support for repository status badges. These badges can be added to your website or project readme file to display the status of your code.

Badge endpoint:

```text
<scheme>://<hostname>/api/badges/<owner>/<repo>/status.svg
```

The status badge displays the status for the latest build to your default branch (e.g. master). You can customize the branch by adding the `branch` query parameter.

```diff
-<scheme>://<hostname>/api/badges/<owner>/<repo>/status.svg
+<scheme>://<hostname>/api/badges/<owner>/<repo>/status.svg?branch=<branch>
```

Please note status badges do not include pull request results, since the status of a pull request does not provide an accurate representation of your repository state.

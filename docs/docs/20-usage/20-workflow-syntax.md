# Workflow syntax

The Workflow section defines a list of steps to build, test and deploy your code. The steps are executed serially in the order in which they are defined. If a step returns a non-zero exit code, the workflow and therefore the entire pipeline terminates immediately and returns an error status.

:::note
An exception to this rule are steps with a [`status: [failure]`](#status) condition, which ensures that they are executed in the case of a failed run.
:::

:::note
We support most of YAML 1.2, but preserve some behavior from 1.1 for backward compatibility.
Read more at: [https://github.com/go-yaml/yaml](https://github.com/go-yaml/yaml/tree/v3)
:::

Example steps:

```yaml
steps:
  - name: backend
    image: golang
    commands:
      - go build
      - go test
  - name: frontend
    image: node
    commands:
      - npm install
      - npm run test
      - npm run build
```

In the above example we define two steps, `frontend` and `backend`. The names of these steps are completely arbitrary.

The name is optional, if not added the steps will be numerated.

Another way to name a step is by using dictionaries:

```yaml
steps:
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

## Skip Commits

Woodpecker gives the ability to skip individual commits by adding `[SKIP CI]` or `[CI SKIP]` to the commit message. Note this is case-insensitive.

```bash
git commit -m "updated README [CI SKIP]"
```

## Steps

Every step of your workflow executes commands inside a specified container.<br>
The defined steps are executed in sequence by default, if they should run in parallel you can use [`depends_on`](./20-workflow-syntax.md#depends_on).<br>
The associated commit is checked out with git to a workspace which is mounted to every step of the workflow as the working directory.

```diff
 steps:
   - name: backend
     image: golang
     commands:
+      - go build
+      - go test
```

### File changes are incremental

- Woodpecker clones the source code in the beginning of the workflow
- Changes to files are persisted through steps as the same volume is mounted to all steps

```yaml title=".woodpecker.yaml"
steps:
  - name: build
    image: debian
    commands:
      - echo "test content" > myfile
  - name: a-test-step
    image: debian
    commands:
      - cat myfile
```

### `image`

Woodpecker pulls the defined image and uses it as environment to execute the workflow step commands, for plugins and for service containers.

When using the `local` backend, the `image` entry is used to specify the shell, such as Bash or Fish, that is used to run the commands.

```diff
 steps:
   - name: build
+    image: golang:1.6
     commands:
       - go build
       - go test

   - name: publish
+    image: woodpeckerci/plugin-kaniko
     repo: foo/bar

 services:
   - name: database
+    image: mysql
```

Woodpecker supports any valid Docker image from any Docker registry:

```yaml
image: golang
image: golang:1.7
image: library/golang:1.7
image: index.docker.io/library/golang
image: index.docker.io/library/golang:1.7
```

Woodpecker does not automatically upgrade container images. Example configuration to always pull the latest image when updates are available:

```diff
 steps:
   - name: build
     image: golang:latest
+    pull: true
```

Learn more how you can use images from [different registries](./41-registries.md).

### `commands`

Commands of every step are executed serially as if you would enter them into your local shell.

```diff
 steps:
   - name: backend
     image: golang
     commands:
+      - go build
+      - go test
```

There is no magic here. The above commands are converted to a simple shell script. The commands in the above example are roughly converted to the below script:

```bash
#!/bin/sh
set -e

go build
go test
```

The above shell script is then executed as the container entrypoint. The below docker command is an (incomplete) example of how the script is executed:

```bash
docker run --entrypoint=build.sh golang
```

:::note
Only build steps can define commands. You cannot use commands with plugins or services.
:::

### `entrypoint`

Allows you to specify the entrypoint for containers. Note that this must be a list of the command and its arguments (e.g. `["/bin/sh", "-c"]`).

If you define [`commands`](#commands), the default entrypoint will be `["/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"]`.
You can also use a custom shell with `CI_SCRIPT` (Base64-encoded) if you set `commands`.

### `environment`

Woodpecker provides the ability to pass environment variables to individual steps.

For more details, check the [environment docs](./50-environment.md).

### `secrets`

Woodpecker provides the ability to store named parameters external to the YAML configuration file, in a central secret store. These secrets can be passed to individual steps of the workflow at runtime.

For more details, check the [secrets docs](./40-secrets.md).

### `failure`

Some of the steps may be allowed to fail without causing the whole workflow and therefore pipeline to report a failure (e.g., a step executing a linting check). To enable this, add `failure: ignore` to your step. If Woodpecker encounters an error while executing the step, it will report it as failed but still executes the next steps of the workflow, if any, without affecting the status of the workflow.

```diff
 steps:
   - name: backend
     image: golang
     commands:
       - go build
       - go test
+    failure: ignore
```

### `when` - Conditional Execution

Woodpecker supports defining a list of conditions for a step by using a `when` block. If at least one of the conditions in the `when` block evaluate to true the step is executed, otherwise it is skipped. A condition is evaluated to true if _all_ sub-conditions are true.
A condition can be a check like:

```diff
 steps:
   - name: slack
     image: plugins/slack
     settings:
       channel: dev
+    when:
+      - event: pull_request
+        repo: test/test
+      - event: push
+        branch: main
```

The `slack` step is executed if one of these conditions is met:

1. The pipeline is executed from a pull request in the repo `test/test`
2. The pipeline is executed from a push to `main`

#### `repo`

Example conditional execution by repository:

```diff
 steps:
   - name: slack
     image: plugins/slack
     settings:
       channel: dev
+    when:
+      - repo: test/test
```

#### `branch`

:::note
Branch conditions are not applied to tags.
:::

Example conditional execution by branch:

```diff
 steps:
   - name: slack
     image: plugins/slack
     settings:
       channel: dev
+    when:
+      - branch: main
```

> The step now triggers on main branch, but also if the target branch of a pull request is `main`. Add an event condition to limit it further to pushes on main only.

Execute a step if the branch is `main` or `develop`:

```yaml
when:
  - branch: [main, develop]
```

Execute a step if the branch starts with `prefix/*`:

```yaml
when:
  - branch: prefix/*
```

The branch matching is done using [doublestar](https://github.com/bmatcuk/doublestar/#usage), note that a pattern starting with `*` should be put between quotes and a literal `/` needs to be escaped. A few examples:

- `*\\/*` to match patterns with exactly 1 `/`
- `*\\/**` to match patters with at least 1 `/`
- `*` to match patterns without `/`
- `**` to match everything

Execute a step using custom include and exclude logic:

```yaml
when:
  - branch:
      include: [main, release/*]
      exclude: [release/1.0.0, release/1.1.*]
```

#### `event`

The available events are:

- `push`: triggered when a commit is pushed to a branch.
- `pull_request`: triggered when a pull request is opened or a new commit is pushed to it.
- `pull_request_closed`: triggered when a pull request is closed or merged.
- `tag`: triggered when a tag is pushed.
- `release`: triggered when a release, pre-release or draft is created. (You can apply further filters using [evaluate](#evaluate) with [environment variables](./50-environment.md#built-in-environment-variables).)
- `deployment` (only available for GitHub): triggered when a deployment is created in the repository.
- `cron`: triggered when a cron job is executed.
- `manual`: triggered when a user manually triggers a pipeline.

Execute a step if the build event is a `tag`:

```yaml
when:
  - event: tag
```

Execute a step if the pipeline event is a `push` to a specified branch:

```diff
when:
  - event: push
+   branch: main
```

Execute a step for multiple events:

```yaml
when:
  - event: [push, tag, deployment]
```

#### `cron`

This filter **only** applies to cron events and filters based on the name of a cron job.

Make sure to have a `event: cron` condition in the `when`-filters as well.

```yaml
when:
  - event: cron
    cron: sync_* # name of your cron job
```

[Read more about cron](./45-cron.md)

#### `ref`

The `ref` filter compares the git reference against which the workflow is executed.
This allows you to filter, for example, tags that must start with **v**:

```yaml
when:
  - event: tag
    ref: refs/tags/v*
```

#### `status`

There are use cases for executing steps on failure, such as sending notifications for failed workflow / pipeline. Use the status constraint to execute steps even when the workflow fails:

```diff
 steps:
   - name: slack
     image: plugins/slack
     settings:
       channel: dev
+    when:
+      - status: [ success, failure ]
```

#### `platform`

:::note
This condition should be used in conjunction with a [matrix](./30-matrix-workflows.md#example-matrix-pipeline-using-multiple-platforms) workflow as a regular workflow will only be executed by a single agent which only has one arch.
:::

Execute a step for a specific platform:

```yaml
when:
  - platform: linux/amd64
```

Execute a step for a specific platform using wildcards:

```yaml
when:
  - platform: [linux/*, windows/amd64]
```

#### `matrix`

Execute a step for a single matrix permutation:

```yaml
when:
  - matrix:
      GO_VERSION: 1.5
      REDIS_VERSION: 2.8
```

#### `instance`

Execute a step only on a certain Woodpecker instance matching the specified hostname:

```yaml
when:
  - instance: stage.woodpecker.company.com
```

#### `path`

:::info
Path conditions are applied only to **push** and **pull_request** events.
It is currently **only available** for GitHub, GitLab and Gitea (version 1.18.0 and newer)
:::

Execute a step only on a pipeline with certain files being changed:

```yaml
when:
  - path: 'src/*'
```

You can use [glob patterns](https://github.com/bmatcuk/doublestar#patterns) to match the changed files and specify if the step should run if a file matching that pattern has been changed `include` or if some files have **not** been changed `exclude`.

For pipelines without file changes (empty commits or on events without file changes like `tag`), you can use `on_empty` to set whether this condition should be **true** _(default)_ or **false** in these cases.

```yaml
when:
  - path:
      include: ['.woodpecker/*.yaml', '*.ini']
      exclude: ['*.md', 'docs/**']
      ignore_message: '[ALL]'
      on_empty: true
```

:::info
Passing a defined ignore-message like `[ALL]` inside the commit message will ignore all path conditions and the `on_empty` setting.
:::

#### `evaluate`

Execute a step only if the provided evaluate expression is equal to true. Both built-in [`CI_`](./50-environment.md#built-in-environment-variables) and custom variables can be used inside the expression.

The expression syntax can be found in [the docs](https://github.com/expr-lang/expr/blob/master/docs/language-definition.md) of the underlying library.

Run on pushes to the default branch for the repository `owner/repo`:

```yaml
when:
  - evaluate: 'CI_PIPELINE_EVENT == "push" && CI_REPO == "owner/repo" && CI_COMMIT_BRANCH == CI_REPO_DEFAULT_BRANCH'
```

Run on commits created by user `woodpecker-ci`:

```yaml
when:
  - evaluate: 'CI_COMMIT_AUTHOR == "woodpecker-ci"'
```

Skip all commits containing `please ignore me` in the commit message:

```yaml
when:
  - evaluate: 'not (CI_COMMIT_MESSAGE contains "please ignore me")'
```

Run on pull requests with the label `deploy`:

```yaml
when:
  - evaluate: 'CI_COMMIT_PULL_REQUEST_LABELS contains "deploy"'
```

Skip step only if `SKIP=true`, run otherwise or if undefined:

```yaml
when:
  - evaluate: 'SKIP != "true"'
```

### `depends_on`

Normally steps of a workflow are executed serially in the order in which they are defined. As soon as you set `depends_on` for a step a [directed acyclic graph](https://en.wikipedia.org/wiki/Directed_acyclic_graph) will be used and all steps of the workflow will be executed in parallel besides the steps that have a dependency set to another step using `depends_on`:

```diff
 steps:
   - name: build # build will be executed immediately
     image: golang
     commands:
       - go build

   - name: deploy
     image: woodpeckerci/plugin-kaniko
     settings:
       repo: foo/bar
+    depends_on: [build, test] # deploy will be executed after build and test finished

   - name: test # test will be executed immediately as no dependencies are set
     image: golang
     commands:
       - go test
```

:::note
You can define a step to start immediately without dependencies by adding an empty `depends_on: []`. By setting `depends_on` on a single step all other steps will be immediately executed as well if no further dependencies are specified.

```yaml
steps:
  - name: check code format
    image: mstruebing/editorconfig-checker
    depends_on: [] # enable parallel steps
  ...
```

:::

### `volumes`

Woodpecker gives the ability to define Docker volumes in the YAML. You can use this parameter to mount files or folders on the host machine into your containers.

For more details check the [volumes docs](./70-volumes.md).

### `detach`

Woodpecker gives the ability to detach steps to run them in background until the workflow finishes.

For more details check the [service docs](./60-services.md#detachment).

### `directory`

Using `directory`, you can set a subdirectory of your repository or an absolute path inside the Docker container in which your commands will run.

## `services`

Woodpecker can provide service containers. They can for example be used to run databases or cache containers during the execution of workflow.

For more details check the [services docs](./60-services.md).

## `workspace`

The workspace defines the shared volume and working directory shared by all workflow steps.
The default workspace base is `/woodpecker` and the path is extended with the repository URL (`src/{url-without-schema}`).
So an example would be `/woodpecker/src/github.com/octocat/hello-world`.

The workspace can be customized using the workspace block in the YAML file:

```diff
+workspace:
+  base: /go
+  path: src/github.com/octocat/hello-world

 steps:
   - name: build
     image: golang:latest
     commands:
       - go get
       - go test
```

:::note
Plugins will always have the workspace base at `/woodpecker`
:::

The base attribute defines a shared base volume available to all steps. This ensures your source code, dependencies and compiled binaries are persisted and shared between steps.

```diff
 workspace:
+  base: /go
   path: src/github.com/octocat/hello-world

 steps:
   - name: deps
     image: golang:latest
     commands:
       - go get
       - go test
   - name: build
     image: node:latest
     commands:
       - go build
```

This would be equivalent to the following docker commands:

```bash
docker volume create my-named-volume

docker run --volume=my-named-volume:/go golang:latest
docker run --volume=my-named-volume:/go node:latest
```

The path attribute defines the working directory of your build. This is where your code is cloned and will be the default working directory of every step in your build process. The path must be relative and is combined with your base path.

```diff
 workspace:
   base: /go
+  path: src/github.com/octocat/hello-world
```

```bash
git clone https://github.com/octocat/hello-world \
  /go/src/github.com/octocat/hello-world
```

<!-- markdownlint-disable no-duplicate-heading -->

## `matrix`

<!-- markdownlint-enable no-duplicate-heading -->

Woodpecker has integrated support for matrix builds. Woodpecker executes a separate build task for each combination in the matrix, allowing you to build and test a single commit against multiple configurations.

For more details check the [matrix build docs](./30-matrix-workflows.md).

## `labels`

You can set labels for your workflow to select an agent to execute the workflow on. An agent will pick up and run a workflow when **every** label assigned to it matches the agents labels.

To set additional agent labels, check the [agent configuration options](../30-administration/15-agent-config.md#woodpecker_agent_labels). Agents will have at least four default labels: `platform=agent-os/agent-arch`, `hostname=my-agent`, `backend=docker` (type of the agent backend) and `repo=*`. Agents can use a `*` as a wildcard for a label. For example `repo=*` will match every repo.

Workflow labels with an empty value will be ignored.
By default, each workflow has at least the `repo=your-user/your-repo-name` label. If you have set the [platform attribute](#platform) for your workflow it will have a label like `platform=your-os/your-arch` as well.

You can add additional labels as a key value map:

```diff
+labels:
+  location: europe # only agents with `location=europe` or `location=*` will be used
+  weather: sun
+  hostname: "" # this label will be ignored as it is empty

 steps:
   - name: build
     image: golang
     commands:
       - go build
       - go test
```

### Filter by platform

To configure your workflow to only be executed on an agent with a specific platform, you can use the `platform` key.
Have a look at the official [go docs](https://go.dev/doc/install/source) for the available platforms. The syntax of the platform is `GOOS/GOARCH` like `linux/arm64` or `linux/amd64`.

Example:

Assuming we have two agents, one `linux/arm` and one `linux/amd64`. Previously this workflow would have executed on **either agent**, as Woodpecker is not fussy about where it runs the workflows. By setting the following option it will only be executed on an agent with the platform `linux/arm64`.

```diff
+labels:
+  platform: linux/arm64

 steps:
   [...]
```

## `variables`

Woodpecker supports using [YAML anchors & aliases](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases) as variables in the workflow configuration.

For more details and examples check the [Advanced usage docs](./90-advanced-usage.md)

## `clone`

Woodpecker automatically configures a default clone step if not explicitly defined. When using the `local` backend, the [plugin-git](https://github.com/woodpecker-ci/plugin-git) binary must be on your `$PATH` for the default clone step to work. If not, you can still write a manual clone step.

You can manually configure the clone step in your workflow for customization:

```diff
+clone:
+  git:
+    image: woodpeckerci/plugin-git

 steps:
   - name: build
     image: golang
     commands:
       - go build
       - go test
```

Example configuration to override depth:

```diff
 clone:
   - name: git
     image: woodpeckerci/plugin-git
+    settings:
+      partial: false
+      depth: 50
```

Example configuration to use a custom clone plugin:

```diff
 clone:
   - name: git
+    image: octocat/custom-git-plugin
```

Example configuration to clone Mercurial repository:

```diff
 clone:
   - name: hg
+    image: plugins/hg
+    settings:
+      path: bitbucket.org/foo/bar
```

### Git Submodules

To use the credentials that cloned the repository to clone it's submodules, update `.gitmodules` to use `https` instead of `git`:

```diff
 [submodule "my-module"]
 path = my-module
-url = git@github.com:octocat/my-module.git
+url = https://github.com/octocat/my-module.git
```

To use the ssh git url in `.gitmodules` for users cloning with ssh, and also use the https url in Woodpecker, add `submodule_override`:

```diff
 clone:
   - name: git
     image: woodpeckerci/plugin-git
     settings:
       recursive: true
+      submodule_override:
+        my-module: https://github.com/octocat/my-module.git

steps:
  ...
```

## `skip_clone`

By default Woodpecker is automatically adding a clone step. This clone step can be configured by the [clone](#clone) property. If you do not need a `clone` step at all you can skip it using:

```yaml
skip_clone: true
```

## `when` - Global workflow conditions

Woodpecker gives the ability to skip whole workflows ([not just steps](#when---conditional-execution)) based on certain conditions by a `when` block. If all conditions in the `when` block evaluate to true the workflow is executed, otherwise it is skipped, but treated as successful and other workflows depending on it will still continue.

For more information about the specific filters, take a look at the [step-specific `when` filters](#when---conditional-execution).

Example conditional execution by branch:

```diff
+when:
+  branch: main
+
 steps:
   - name: slack
     image: plugins/slack
     settings:
       channel: dev
```

The workflow now triggers on `main`, but also if the target branch of a pull request is `main`.

<!-- markdownlint-disable no-duplicate-heading -->

## `depends_on`

<!-- markdownlint-enable no-duplicate-heading -->

Woodpecker supports to define multiple workflows for a repository. Those workflows will run independent from each other. To depend them on each other you can use the [`depends_on`](./25-workflows.md#flow-control) keyword.

## `runs_on`

Workflows that should run even on failure should set the `runs_on` tag. See [here](./25-workflows.md#flow-control) for an example.

## Privileged mode

Woodpecker gives the ability to configure privileged mode in the YAML. You can use this parameter to launch containers with escalated capabilities.

:::info
Privileged mode is only available to trusted repositories and for security reasons should only be used in private environments. See [project settings](./75-project-settings.md#trusted) to enable trusted mode.
:::

```diff
 steps:
   - name: build
     image: docker
     environment:
       - DOCKER_HOST=tcp://docker:2375
     commands:
       - docker --tls=false ps

 - name: services
   docker:
     image: docker:dind
     commands: dockerd-entrypoint.sh --storage-driver=vfs --tls=false
+    privileged: true
```

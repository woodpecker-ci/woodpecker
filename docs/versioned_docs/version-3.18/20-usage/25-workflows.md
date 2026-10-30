# Workflows

A pipeline has at least one workflow. A workflow is a set of steps that are executed in sequence using the same workspace which is a shared folder containing the repository and all the generated data from previous steps.

In case there is a single configuration in `.woodpecker.yaml` Woodpecker will create a pipeline with a single workflow.

By placing the configurations in a folder which is by default named `.woodpecker/` Woodpecker will create a pipeline with multiple workflows each named by the file they are defined in. Only `.yml` and `.yaml` files will be used and files in any subfolders like `.woodpecker/sub-folder/test.yaml` will be ignored.

You can also set some custom path like `.my-ci/pipelines/` instead of `.woodpecker/` in the [project settings](./75-project-settings.md).

## Benefits of using workflows

- faster lint/test feedback, the workflow doesn't have to run fully to have a lint status pushed to the remote
- better organization of a pipeline along various concerns using one workflow for: testing, linting, building and deploying
- utilizing more agents to speed up the execution of the whole pipeline

## Example workflow definition

:::warning
Please note that files are only shared between steps of the same workflow (see [File changes are incremental](./20-workflow-syntax.md#file-changes-are-incremental)). That means you cannot access artifacts e.g. from the `build` workflow in the `deploy` workflow.
If you still need to pass artifacts between the workflows you need use some storage [plugin](./51-plugins/51-overview.md) (e.g. one which stores files in an Amazon S3 bucket).
:::

```bash
.woodpecker/
├── build.yaml
├── deploy.yaml
├── lint.yaml
└── test.yaml
```

```yaml title=".woodpecker/build.yaml"
steps:
  - name: build
    image: debian:stable-slim
    commands:
      - echo building
      - sleep 5
```

```yaml title=".woodpecker/deploy.yaml"
steps:
  - name: deploy
    image: debian:stable-slim
    commands:
      - echo deploying

depends_on:
  - lint
  - build
  - test
```

```yaml title=".woodpecker/test.yaml"
steps:
  - name: test
    image: debian:stable-slim
    commands:
      - echo testing
      - sleep 5

depends_on:
  - build
```

```yaml title=".woodpecker/lint.yaml"
steps:
  - name: lint
    image: debian:stable-slim
    commands:
      - echo linting
      - sleep 5
```

## Status lines

Each workflow will report its own status back to your forge.

## Flow control

The workflows run in parallel on separate agents and share nothing.

Dependencies between workflows can be set with the `depends_on` element. A workflow doesn't execute until all of its dependencies finished successfully.

The name for a `depends_on` entry is the filename without the path, leading dots and without the file extension `.yml` or `.yaml`. If the project config for example uses `.woodpecker/` as path for CI files with a file named `.woodpecker/.lint.yaml` the corresponding `depends_on` entry would be `lint`.

```diff
 steps:
   - name: deploy
     image: debian:stable-slim
     commands:
       - echo deploying

+depends_on:
+  - lint
+  - build
+  - test
```

Workflows that need to run even on failures should set the `status` filter.

```diff
 steps:
   - name: notify
     image: debian:stable-slim
     commands:
       - echo notifying

 depends_on:
   - deploy

+when:
+  - status: [ success, failure ]
```

This works just like the [`status` filter for steps](./20-workflow-syntax.md#status).

### Optional dependencies

In a monorepo, workflows often use `when: path` to only run when relevant files change. A deploy workflow may need to wait for all check workflows, but some of them might not run because their path filter didn't match. With `depends_on`, this would block the deploy workflow entirely.

Mark a dependency as `optional: true` so it is only enforced when the referenced workflow is part of the pipeline. If the dependency is not built (e.g. its `when` conditions don't match), it is silently ignored.

```diff
 steps:
   - name: deploy
     image: debian:stable-slim
     commands:
       - echo deploying app a

 depends_on:
   - check-a
+  - name: check-b
+    optional: true
+  - name: check-c
+    optional: true
```

In this example, `deploy` always waits for `check-a`. It also waits for `check-b` and `check-c` if they are part of the pipeline, but runs without them if they were filtered out.

The same syntax works at the step level within a workflow: if a step uses `depends_on` with `optional: true` on another step that was filtered out by a `when` condition, the dependency is silently dropped.

:::info
Some workflows don't need the source code, like creating a notification on failure.
Read more about `skip_clone` at [pipeline syntax](./20-workflow-syntax.md#skip_clone)
:::

## Concurrency

By default workflows run with no concurrency limit. Some workflows, however, must not run more than a given number of times at once. A typical example is a deployment workflow: running two deployments at the same time can cause race conditions or corrupt state. Cancelling the previous pipeline is often not an option either, since it could interrupt an ongoing deployment.

The `concurrency` setting limits how many instances of a workflow may run at the same time. When the limit is reached, additional instances stay queued and start only once a running one has finished. Nothing is cancelled.

```yaml title=".woodpecker/deploy.yaml"
steps:
  - name: deploy
    image: debian:stable-slim
    commands:
      - echo deploying

depends_on:
  - test

concurrency:
  limit: 1
```

You can also use the shorthand form to only set the limit:

```yaml
concurrency: 1
```

### Ordering

Queued workflows of the same group start in the order their pipelines were created, **not** in the order they become ready to run. This matters when a workflow depends on other workflows (via `depends_on`) whose duration varies: even if a later pipeline's checks finish first, its limited workflow will not overtake an earlier pipeline that is still waiting. This guarantees that, for example, deployments happen in commit order.

### Groups

By default, the limit applies per workflow within a repository. Different runs of the same workflow are limited against each other, while different workflows (and other repositories) are unaffected.

Setting a `group` is optional. You can set a custom `group` to share a limit across workflows or to make the limit more specific. The group supports [environment variable substitution](./50-environment.md), so you can, for example, limit concurrency per branch or per deployment target:

```yaml
concurrency:
  limit: 1
  group: deploy-${CI_COMMIT_BRANCH}
```

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

Workflows that need to run even on failures should set the `runs_on` tag.

```diff
 steps:
   - name: notify
     image: debian:stable-slim
     commands:
       - echo notifying

 depends_on:
   - deploy

+runs_on: [ success, failure ]
```

:::info
Some workflows don't need the source code, like creating a notification on failure.
Read more about `skip_clone` at [pipeline syntax](./20-workflow-syntax.md#skip_clone)
:::

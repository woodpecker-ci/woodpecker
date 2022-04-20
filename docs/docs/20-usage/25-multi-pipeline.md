# Multi pipelines

:::info
This Feature is only available for GitHub, Gitea & GitLab repositories. Follow [this](https://github.com/woodpecker-ci/woodpecker/issues/131) issue to support further development.
:::

By default, Woodpecker looks for the pipeline definition in `.woodpecker.yml` in the project root.

The Multi-Pipeline feature allows the pipeline to be split into several files and placed in the `.woodpecker/` folder. Only `.yml` files will be used and files in any subfolders like `.woodpecker/sub-folder/test.yml` will be ignored. You can set some custom path like `.my-ci/pipelines/` instead of `.woodpecker/` in the [project settings](/docs/usage/project-settings).

## Rational

- faster lint/test feedback, the pipeline doesn't have to run fully to have a lint status pushed to the remote
- better organization of the pipeline along various concerns: testing, linting, feature apps
- utilizing more agents to speed up build

## Example multi-pipeline definition
:::warning
Please note that files are only shared between steps of the same pipeline (see [File changes are incremental](/docs/usage/pipeline-syntax#file-changes-are-incremental)). That means you cannot access artifacts e.g. from the `build` pipeline below in the `deploy` pipeline.
If you still need to pass artifacts between the pipelines you need use storage [plugins](/docs/usage/plugins/plugins) (e.g. one which stores files in an Amazon S3 bucket).
:::

```bash
.woodpecker/
├── .build.yml
├── .deploy.yml
├── .lint.yml
└── .test.yml
```

.woodpecker/.build.yml

```yaml
pipeline:
  build:
    image: debian:stable-slim
    commands:
      - echo building
      - sleep 5
```

.woodpecker/.deploy.yml

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

.woodpecker/.test.yml

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

.woodpecker/.lint.yml

```yaml
pipeline:
  lint:
    image: debian:stable-slim
    commands:
      - echo linting
      - sleep 5
```

## Status lines

Each pipeline will report its own status back to your forge.

## Flow control

The pipelines run in parallel on separate agents and share nothing.

Dependencies between pipelines can be set with the `depends_on` element. A pipeline doesn't execute until its dependencies finish successfully.

The name for a `depends_on` entry is the filename without the path, leading dots and without the file extension `.yml`. If the project config for example uses `.woodpecker/` as path for ci files with a file named `.woodpecker/.lint.yml` the corresponding `depends_on` entry would be `lint`.

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


# Multi-pipeline builds

By default, Woodpecker looks for the pipeline definition in `.drone.yml` in the project root.

The Multi-Pipeline feature allows the pipeline to be splitted to several files and placed in the `.drone/` folder

## Rational

- faster lint/test feedback, the pipeline doesn't have to run fully to have a lint status pushed to the the remote
- better organization of the pipeline along various concerns: testing, linting, feature apps
- utilizaing more agents to speed up build

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

## Status lines

Each pipeline has its own status line on Github.


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

# Matrix workflows

Woodpecker has integrated support for matrix workflows. Woodpecker executes a separate workflow for each combination in the matrix, allowing you to build and test a single commit against multiple configurations.

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

Matrix variables are interpolated in the YAML using the `${VARIABLE}` syntax, before the YAML is parsed. This is an example YAML file before interpolating matrix parameters:

```yaml
matrix:
  GO_VERSION:
    - 1.4
    - 1.3
  DATABASE:
    - mysql:5.5
    - mysql:6.5
    - mariadb:10.1

steps:
  build:
    image: golang:${GO_VERSION}
    commands:
      - go get
      - go build
      - go test

services:
  database:
    image: ${DATABASE}
```

Example YAML file after injecting the matrix parameters:

```diff
steps:
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

### Example matrix pipeline based on Docker image tag

```yaml
matrix:
  TAG:
    - 1.7
    - 1.8
    - latest

steps:
  build:
    image: golang:${TAG}
    commands:
      - go build
      - go test
```

### Example matrix pipeline based on container image

```yaml
matrix:
  IMAGE:
    - golang:1.7
    - golang:1.8
    - golang:latest

steps:
  build:
    image: ${IMAGE}
    commands:
      - go build
      - go test
```

### Example matrix pipeline using multiple platforms

```yaml
matrix:
  platform:
    - linux/amd64
    - linux/arm64

labels:
  platform: ${platform}

steps:
  test:
    image: alpine
    commands:
      - echo "I am running on ${platform}"

  test-arm-only:
    image: alpine
    commands:
      - echo "I am running on ${platform}"
      - echo "Arm is cool!"
    when:
      platform: linux/arm*
```

:::note
If you want to control the architecture of a pipeline on a Kubernetes runner, see [the nodeSelector documentation of the Kubernetes backend](../30-administration/22-backends/40-kubernetes.md#nodeSelector).
:::

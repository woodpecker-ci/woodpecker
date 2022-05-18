# Matrix builds

Woodpecker has integrated support for matrix builds. Woodpecker executes a separate build task for each combination in the matrix, allowing you to build and test a single commit against multiple configurations.

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

Example YAML file after injecting the matrix parameters:

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

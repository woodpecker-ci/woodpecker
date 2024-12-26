# Advanced usage

## Advanced YAML syntax

YAML has some advanced syntax features that can be used like variables to reduce duplication in your pipeline config:

### Anchors & aliases

You can use [YAML anchors & aliases](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases) as variables in your pipeline config.

To convert this:

```yaml
steps:
  - name: test
    image: golang:1.18
    commands: go test ./...
  - name: build
    image: golang:1.18
    commands: build
```

Just add a new section called **variables** like this:

```diff
+variables:
+  - &golang_image 'golang:1.18'

 steps:
   - name: test
-    image: golang:1.18
+    image: *golang_image
     commands: go test ./...
   - name: build
-    image: golang:1.18
+    image: *golang_image
     commands: build
```

### Map merges and overwrites

```yaml
variables:
  - &base-plugin-settings
    target: dist
    recursive: false
    try: true
  - &special-setting
    special: true
  - &some-plugin codeberg.org/6543/docker-images/print_env

steps:
  - name: develop
    image: *some-plugin
    settings:
      <<: [*base-plugin-settings, *special-setting] # merge two maps into an empty map
    when:
      branch: develop

  - name: main
    image: *some-plugin
    settings:
      <<: *base-plugin-settings # merge one map and ...
      try: false # ... overwrite original value
      ongoing: false # ... adding a new value
    when:
      branch: main
```

### Sequence merges

```yaml
variables:
  pre_cmds: &pre_cmds
    - echo start
    - whoami
  post_cmds: &post_cmds
    - echo stop
  hello_cmd: &hello_cmd
    - echo hello

steps:
  - name: step1
    image: debian
    commands:
      - <<: *pre_cmds # prepend a sequence
      - echo exec step now do dedicated things
      - <<: *post_cmds # append a sequence
  - name: step2
    image: debian
    commands:
      - <<: [*pre_cmds, *hello_cmd] # prepend two sequences
      - echo echo from second step
      - <<: *post_cmds
```

### References

- [Official YAML specification](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases)
- [YAML cheat sheet](https://learnxinyminutes.com/docs/yaml)

## Persisting environment data between steps

One can create a file containing environment variables, and then source it in each step that needs them.

```yaml
steps:
  - name: init
    image: bash
    commands:
      - echo "FOO=hello" >> envvars
      - echo "BAR=world" >> envvars

  - name: debug
    image: bash
    commands:
      - source envvars
      - echo $FOO
```

## Declaring global variables

As described in [Global environment variables](./50-environment.md#global-environment-variables), you can define global variables:

```ini
WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
```

Note that this tightly couples the server and app configurations (where the app is a completely separate application). But this is a good option for truly global variables which should apply to all steps in all pipelines for all apps.

## Docker in docker (dind) setup

:::warning
This set up will only work on trusted repositories and for security reasons should only be used in private environments.
See [project settings](./75-project-settings.md#trusted) to enable trusted mode.
:::

The snippet below shows how a step can communicate with the docker daemon via a `docker:dind` service.

:::note
If your aim ist to build/publish OCI images, consider using the [Docker Buildx Plugin](https://woodpecker-ci.org/plugins/Docker%20Buildx) instead.
:::

First we need to define a servie running a docker with the `dind` tag. This service must run in privileged mode:

```yaml
services:
  - name: docker
    image: docker:27.4-dind
    privileged: true
    ports:
      - 2376
```

Next we need to set up TLS communication between the `dind` service and the step that wants to communicate with the docker daemon (since Unauthenticated TCP connections have been deprecated [as of docker v26](https://github.com/docker/cli/blob/v27.4.0/docs/deprecated.md#unauthenticated-tcp-connections) and will ve removed in release v28).

We can achieve this by letting the daemon generate TLS certificates for us and share them with the client via a volume mount in the agent (`/opt/woodpeckerci/dind-certs` in the example below).

```diff
services:
  - name: docker
    image: docker:27.4-dind
    privileged: true
+    environment:
+      DOCKER_TLS_CERTDIR: /dind-certs
+    volumes:
+      - /opt/woodpeckerci/dind-certs:/dind-certs
     ports:
       - 2376
```
In the step that needs access to the daemon we need to:

1. Set the `DOCKER_*` environment variables shown below, setting up the connection with the daemon. These are standardized environment variables that should work with the docker client used by your framework of choice (e.g. [TestContainers](https://testcontainers.com/), [Spring Boot Docker Compose](https://mvnrepository.com/artifact/org.springframework.boot/spring-boot-docker-compose) or similar).
2. Mount the volume where the daemon has created the certificates (`/opt/woodpeckerci/dind-certs`)

In this example we test the connection with the vanilla docker client:

```diff
steps:
  - name: test
    image: docker:27.4-cli
+    environment:
+      DOCKER_HOST: "tcp://docker:2376"
+      DOCKER_CERT_PATH: "/dind-certs/client"27.4-cli
+      DOCKER_TLS_VERIFY: "1"
+    volumes:
+      - /opt/woodpeckerci/dind-certs:/dind-certs
    commands:
      - docker version
```

This step should output version information of the client and the server if everything has been set correctly.

Complete example:

```yaml
steps:
  - name: test
    image: docker:27.4-cli
    environment:
      DOCKER_HOST: "tcp://docker:2376"
      DOCKER_CERT_PATH: "/dind-certs/client"27.4-cli
      DOCKER_TLS_VERIFY: "1"
    volumes:
      - /opt/woodpeckerci/dind-certs:/dind-certs
    commands:
      - docker version

services:
  - name: docker
    image: docker:27.4-dind
    privileged: true
    environment:
      DOCKER_TLS_CERTDIR: /dind-certs
    volumes:
      - /opt/woodpeckerci/dind-certs:/dind-certs
    ports:
      - 2376
```

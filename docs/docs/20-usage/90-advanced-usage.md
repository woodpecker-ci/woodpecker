# Advanced usage

## Advanced YAML syntax

YAML has some advanced syntax features that can be used like variables to reduce duplication in your pipeline config:

### Anchors & aliases

You can use [YAML anchors & aliases](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases) as variables in your pipeline config.

To convert this:

```yaml title=".woodpecker.yml"
steps:
  test:
    image: golang:1.18
    commands: go test ./...
  build:
    image: golang:1.18
    commands: build
```

Just add a new section called **variables** like this:

```diff title=".woodpecker.yml"
+variables:
+  - &golang_image 'golang:1.18'

 steps:
   test:
-    image: golang:1.18
+    image: *golang_image
     commands: go test ./...
   build:
-    image: golang:1.18
+    image: *golang_image
     commands: build
```

### Map merges and overwrites

```yaml title=".woodpecker.yml"
variables:
  - &base-plugin-settings
    target: dist
    recursive: false
    try: true
  - &special-setting
    special: true
  - &some-plugin codeberg.org/6543/docker-images/print_env

steps:
  develop:
    image: *some-plugin
    settings:
      <<: [*base-plugin-settings, *special-setting] # merge two maps into an empty map
    when:
      branch: develop

  main:
    image: *some-plugin
    settings:
      <<: *base-plugin-settings # merge one map and ...
      try: false # ... overwrite original value
      ongoing: false # ... adding a new value
    when:
      branch: main
```

### Sequence merges

```yaml title=".woodpecker.yml"
variables:
  pre_cmds: &pre_cmds
    - echo start
    - whoami
  post_cmds: &post_cmds
    - echo stop
  hello_cmd: &hello_cmd
    - echo hello

steps:
  step1:
    image: debian
    commands:
      - <<: *pre_cmds # prepend a sequence
      - echo exec step now do dedicated things
      - <<: *post_cmds # append a sequence
  step2:
    image: debian
    commands:
      - <<: [*pre_cmds, *hello_cmd] # prepend two sequences
      - echo echo from second step
      - <<: *post_cmds
```

### References

- [Official YAML specification](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases)
- [YAML Cheatsheet](https://learnxinyminutes.com/docs/yaml)

## Persisting environment data between steps

One can create a file containing environment variables, and then source it in each step that needs them.

```yaml title=".woodpecker.yml"
steps:
  init:
    image: bash
    commands:
      - echo "FOO=hello" >> envvars
      - echo "BAR=world" >> envvars

  debug:
    image: bash
    commands:
      - source envvars
      - echo $FOO
```

## Declaring global variables in `docker-compose.yml`

As described in [Global environment variables](./50-environment.md#global-environment-variables), one can define global variables:

```yaml title="docker-compose.yml"
services:
  woodpecker-server:
    # ...
    environment:
      - WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
      # ...
```

Note that this tightly couples the server and app configurations (where the app is a completely separate application). But this is a good option for truly global variables which should apply to all steps in all pipelines for all apps.

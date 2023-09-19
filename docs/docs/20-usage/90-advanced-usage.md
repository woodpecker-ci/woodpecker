# Advanced usage

## Advanced YAML syntax

### Anchors & aliases

You can use [YAML anchors & aliases](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases) as variables in your pipeline config.

To convert this:
```yml
steps:
  test:
    image: golang:1.18
    commands: go test ./...
  build:
    image: golang:1.18
    commands: build
```

Just add a new section called **variables** like this:

```diff
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

## Using variables

Once your pipeline starts to grow in size, it will become important to keep it DRY ("Don't Repeat Yourself") by using variables and environment variables. Depending on your specific need, there are a number of options.

### YAML extensions

As described in [Advanced YAML syntax](./35-advanced-yaml-syntax.md).

```yml
variables:
  - &golang_image 'golang:1.18'

 steps:
   build:
     image: *golang_image
     commands: build
```

Note that the `golang_image` alias cannot be used with string interpolation. But this is otherwise a good option for most cases.

### YAML extensions (alternate form)

Another approach using YAML extensions:

```yml
variables:
  - global_env: &global_env
    - BASH_VERSION=1.2.3
    - PATH_SRC=src/
    - PATH_TEST=test/
    - FOO=something

steps:
  build:
    image: bash:${BASH_VERSION}
    directory: ${PATH_SRC}
    commands:
      - make ${FOO} -o ${PATH_TEST}
    environment: *global_env

  test:
    image: bash:${BASH_VERSION}
    commands:
      - test ${PATH_TEST}
    environment:
      - <<:*global_env
      - ADDITIONAL_LOCAL="var value"
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

- [Official specification](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases)
- [Cheatsheet](https://learnxinyminutes.com/docs/yaml)

## Persisting environment data between steps

One can create a file containing environment variables, and then source it in each step that needs them.

```yml
steps:
  init:
    image: bash
    commands:
      echo "FOO=hello" >> envvars
      echo "BAR=world" >> envvars

  debug:
    image: bash
    commands:
      - source envvars
      - echo $FOO
```

### Declaring global variables in `docker-compose.yml`

As described in [Global environment variables](./50-environment.md#global-environment-variables), one can define global variables:

```yml
services:
  woodpecker-server:
    # ...
    environment:
      - WOODPECKER_ENVIRONMENT=first_var:value1,second_var:value2
      # ...
```

Note that this tightly couples the server and app configurations (where the app is a completely separate application). But this is a good option for truly global variables which should apply to all steps in all pipelines for all apps.

# Advanced YAML syntax

## Anchors & aliases

You can use [YAML anchors & aliases](https://yaml.org/spec/1.2.2/#3222-anchors-and-aliases) as variables in your pipeline config.

To convert this:
```yml
pipeline:
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

 pipeline:
   test:
-    image: golang:1.18
+    image: *golang_image
     commands: go test ./...
   build:
-    image: golang:1.18
+    image: *golang_image
     commands: build
```

## Map merges and overwrites

```yaml
variables:
  &base-plugin-settings
    target: dist
    recursive: false
    try: true
  &special-setting
    special: true

pipeline:
  develop:
    image: some-plugin
    settings:
      <<: [*base-plugin-settings, *special-setting] # merge two maps into an empty map
    when:
      branch: develop

  main:
    image: some-plugin
    settings:
      <<: *base-plugin-settings # merge one map and ...
      try: false # ... overwrite original value
      ongoing: false # ... adding a new value
    when:
      branch: main
```

## Sequence merges

```yaml
variables:
  &pre_cmds
   - echo start
   - whoami
  &post_cmds
   - echo stop
  &hello_cmd
   - echo hello

pipeline:
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

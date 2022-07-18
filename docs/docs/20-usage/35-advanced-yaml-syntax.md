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

## Example of YAML override and extension

```yml
variables: 
  &some-plugin-settings
      settings:
        target: dist
        recursive: false
        try: true

pipelines:
  develop:
    name: Build and test
    image: some-plugin
    settings: *some-plugin-settings
    when:
      branch: develop

  main
    name: Build and test
    image: some-plugin
    settings:
      <<: *some-plugin-settings
      try: false # replacing original value from `some-plugin-settings`
      ongoing: false # adding a new value to `some-plugin-settings`
    when:
      branch: main
```

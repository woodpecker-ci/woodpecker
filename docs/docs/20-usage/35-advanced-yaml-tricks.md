# Advanced YAML tricks

## Aliases

with aliases you can have variables in your pipeline config.

to convert this:
```yml
pipeline:
  test:
    image: golang:1.18
    command: go test ./...
  build:
    image: golang:1.18
    command: build
```

just use a new section called **variables**:

```diff
+variables:
+  - &golang_image 'golang:1.18'
+
 pipeline:
   test:
-    image: golang:1.18
+    image: *golang_image
     command: go test ./...
   build:
-    image: golang:1.18
+    image: *golang_image
     command: build
```

## Overrides and Extensions

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

  master:
    name: Build and test
    image: some-plugin
    settings:
      <<: *some-plugin-settings
      try: false #override
      ongoing: false #extension
    when:
      branch: master
```

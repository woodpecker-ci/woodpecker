# Advanced YAML tricks

## Aliases

With aliases you can have variables in your pipeline config.

To convert this:
```yml
pipeline:
  test:
    image: golang:1.18
    command: go test ./...
  build:
    image: golang:1.18
    command: build
```

Just use a new section called **variables**:

```diff
+variables:
+  - &golang_image 'golang:1.18'
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

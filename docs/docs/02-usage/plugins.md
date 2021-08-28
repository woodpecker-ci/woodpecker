# Plugins

Plugins are Docker containers that perform pre-defined tasks and are configured as steps in your pipeline. Plugins can be used to deploy code, publish artifacts, send notification, and more.

Example pipeline using the Docker and Slack plugins:

```yaml
pipeline:
  build:
    image: golang
    commands:
      - go build
      - go test

  publish:
    image: plugins/docker
    repo: foo/bar
    tags: latest

  notify:
    image: plugins/slack
    channel: dev
```

## Plugin Isolation

Plugins are executed in Docker containers and are isolated from the other steps in your build pipeline. Plugins do share the build workspace, mounted as a volume, and therefore have access to your source tree.

## Creating a plugin

See a [detailed plugin example](/usage/bash-plugin).

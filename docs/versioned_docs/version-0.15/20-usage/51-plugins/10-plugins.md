# Plugins

Plugins are pipeline steps that perform pre-defined tasks and are configured as steps in your pipeline. Plugins can be used to deploy code, publish artifacts, send notification, and more.

They are automatically pulled from [plugins.drone.io](http://plugins.drone.io).

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
    settings:
      repo: foo/bar
      tags: latest

  notify:
    image: plugins/slack
    settings:
      channel: dev
```

## Plugin Isolation

Plugins are just pipeline steps. They share the build workspace, mounted as a volume, and therefore have access to your source tree.

## Creating a plugin

See a [detailed plugin example](./20-sample-plugin.md).

# Plugins

Plugins are pipeline steps that perform pre-defined tasks and are configured as steps in your pipeline. Plugins can be used to deploy code, publish artifacts, send notification, and more.

They are automatically pulled from the default container registry the agent's have configured.

Example pipeline using the Docker and Slack plugins:

```yaml
steps:
  - name: build
    image: golang
    commands:
      - go build
      - go test

  - name: publish
    image: plugins/docker
    settings:
      repo: foo/bar
      tags: latest

  - name: notify
    image: plugins/slack
    settings:
      channel: dev
```

## Plugin Isolation

Plugins are just pipeline steps. They share the build workspace, mounted as a volume, and therefore have access to your source tree.

## Finding Plugins

For official plugins, you can use the Woodpecker plugin index:

- [Official Woodpecker Plugins](https://woodpecker-ci.org/plugins)

:::tip
There are also other plugin lists with additional plugins. Keep in mind that [Drone](https://www.drone.io/) plugins are generally supported, but could need some adjustments and tweaking.

- [Drone Plugins](http://plugins.drone.io)
- [Geeklab Woodpecker Plugins](https://woodpecker-plugins.geekdocs.de/)

:::

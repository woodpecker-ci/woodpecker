# Plugins

Plugins are pipeline steps that perform pre-defined tasks and are configured as steps in your pipeline.
Plugins can be used to deploy code, publish artifacts, send notification, and more.

They are automatically pulled from the default container registry the agent's have configured.

```dockerfile title="Dockerfile"
FROM cloud/kubectl
COPY deploy /usr/local/deploy
ENTRYPOINT ["/usr/local/deploy"]
```

```bash title="deploy"
kubectl apply -f $PLUGIN_TEMPLATE
```

```yaml title=".woodpecker.yaml"
steps:
  - name: deploy-to-k8s
    image: cloud/my-k8s-plugin
    settings:
      template: config/k8s/service.yaml
```

Example pipeline using the Docker and Slack plugins:

```yaml
steps:
  - name: build
    image: golang
    commands:
      - go build
      - go test

  - name: publish
    image: woodpeckerci/plugin-kaniko
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
While normal steps are all about arbitrary code execution, plugins should only allow the functions intended by the plugin author.

That's why there are a few limitations. The workspace base is always mounted at `/woodpecker`, but the working directory is dynamically
adjusted accordingly, as user of a plugin you should not have to care about this. Also, you cannot use the plugin together with `commands`
or `entrypoint` which will fail. Using `secrets` or `environment` is possible, but in this case, the plugin is internally not treated as plugin
anymore. The container then cannot access secrets with plugin filter anymore and the containers won't be privileged without explicit definition.

## Finding Plugins

For official plugins, you can use the Woodpecker plugin index:

- [Official Woodpecker Plugins](https://woodpecker-ci.org/plugins)

:::tip
There are also other plugin lists with additional plugins. Keep in mind that [Drone](https://www.drone.io/) plugins are generally supported, but could need some adjustments and tweaking.

- [Drone Plugins](http://plugins.drone.io)
- [Geeklab Woodpecker Plugins](https://woodpecker-plugins.geekdocs.de/)

:::

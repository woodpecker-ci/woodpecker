# Welcome to Woodpecker

Woodpecker is a simple CI engine with great extensibility. It focuses on executing pipelines inside [containers](https://opencontainers.org/).
If you are already using containers in your daily workflow, you'll for sure love Woodpecker.

![woodpecker](woodpecker.png)

## .woodpecker.yml

- Place your pipeline in a file named `.woodpecker.yml` in your repository
- Pipeline steps can be named as you like
- Run any command in the commands section

```yaml
# .woodpecker.yml
steps:
  build:
    image: debian
    commands:
      - echo "This is the build step"
  a-test-step:
    image: debian
    commands:
      - echo "Testing.."
```

### Steps are containers

- Define any container image as context
  - either use your own and install the needed tools in a custom image
  - or search for available images that are already tailored for your needs in image registries like [Docker Hub](https://hub.docker.com/search?type=image)
- List the commands that should be executed in the container

```diff
steps:
  build:
-   image: debian
+   image: mycompany/image-with-awscli
    commands:
      - aws help
```

### File changes are incremental

- Woodpecker clones the source code in the beginning
- File changes are persisted throughout individual steps as the same volume is being mounted in all steps

```yaml
# .woodpecker.yml
steps:
  build:
    image: debian
    commands:
      - touch myfile
  a-test-step:
    image: debian
    commands:
      - cat myfile
```

## Plugins are straightforward

- If you copy the same shell script from project to project
- Pack it into a plugin instead
- And make the yaml declarative
- Plugins are Docker images with your script as an entrypoint

```Dockerfile
# Dockerfile
FROM laszlocloud/kubectl
COPY deploy /usr/local/deploy
ENTRYPOINT ["/usr/local/deploy"]
```

```bash
# deploy
kubectl apply -f $PLUGIN_TEMPLATE
```

```yaml
# .woodpecker.yml
steps:
  deploy-to-k8s:
    image: laszlocloud/my-k8s-plugin
    settings:
      template: config/k8s/service.yml
```

See [plugin docs](./20-usage/51-plugins/10-plugins.md).

## Continue reading

- [Create a Woodpecker pipeline for your repository](./20-usage/10-intro.md)
- [Setup your own Woodpecker instance](./30-administration/00-deployment/00-overview.md)

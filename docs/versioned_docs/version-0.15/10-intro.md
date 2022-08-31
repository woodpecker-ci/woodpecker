# Welcome to Woodpecker

Woodpecker is a simple CI engine with great extensibility. It runs your pipelines inside [Docker](https://www.docker.com/) containers, so if you are already using them in your daily workflow, you'll love Woodpecker for sure.

![woodpecker](woodpecker.png)

## .woodpecker.yml

- Place your pipeline in a file named `.woodpecker.yml` in your repository
- Pipeline steps can be named as you like
- Run any command in the commands section

```yaml
# .woodpecker.yml
pipeline:
  build:
    image: debian
    commands:
      - echo "This is the build step"
  a-test-step:
    image: debian
    commands:
      - echo "Testing.."
```

### Build steps are containers

- Define any Docker image as context
  - either use your own and install the needed tools in custom Docker images, or
  - search [Docker Hub](https://hub.docker.com/search?type=image) for images that are already tailored for your needs) 
- List the commands that should be executed in your container, in order to build or test your application

```diff
pipeline:
  build:
-   image: debian
+   image: mycompany/image-with-awscli
    commands:
      - aws help
```

### File changes are incremental

- Woodpecker clones the source code in the beginning pipeline
- Changes to files are persisted through steps as the same volume is mounted to all steps

```yaml
# .woodpecker.yml
pipeline:
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
pipeline:
  deploy-to-k8s:
    image: laszlocloud/my-k8s-plugin
    template: config/k8s/service.yml
```

See [plugin docs](./20-usage/51-plugins/10-plugins.md).

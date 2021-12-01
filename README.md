# test 1

<p align="center">
  <a href="https://github.com/woodpecker-ci/woodpecker/">
    <img alt="Woodpecker" src="https://raw.githubusercontent.com/woodpecker-ci/woodpecker/master/docs/static/img/logo.svg" width="220"/>
  </a>
</p>
<br/>
<p align="center">
  <a href="https://wp.laszlo.cloud/woodpecker-ci/woodpecker" title="Build Status">
    <img src="https://wp.laszlo.cloud/api/badges/woodpecker-ci/woodpecker/status.svg">
  </a>
  <a href="https://discord.gg/fcMQqSMXJy" title="Join the Discord chat at https://discord.gg/fcMQqSMXJy">
    <img src="https://img.shields.io/discord/838698813463724034.svg">
  </a>
  <a href="https://goreportcard.com/badge/github.com/woodpecker-ci/woodpecker" title="Go Report Card">
    <img src="https://goreportcard.com/badge/github.com/woodpecker-ci/woodpecker">
  </a>
  <a href="https://godoc.org/github.com/woodpecker-ci/woodpecker" title="GoDoc">
    <img src="https://godoc.org/github.com/woodpecker-ci/woodpecker?status.svg">
  </a>
  <a href="https://github.com/woodpecker-ci/woodpecker/releases/latest" title="GitHub release">
    <img src="https://img.shields.io/github/v/release/woodpecker-ci/woodpecker?sort=semver">
  </a>
  <a href="https://hub.docker.com/r/woodpeckerci/woodpecker-server" title="Docker pulls">
    <img src="https://img.shields.io/docker/pulls/woodpeckerci/woodpecker-server">
  </a>
  <a href="https://opensource.org/licenses/Apache-2.0" title="License: Apache-2.0">
    <img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg">
  </a>
</p>
<br/>

# Woodpecker

> Woodpecker is a community fork of the Drone CI system.

![woodpecker](docs/docs/woodpecker.png)

## Support

Please consider to donate and become a backer. 🙏 [[Become a backer](https://opencollective.com/woodpecker-ci#category-CONTRIBUTE)]
<a href="https://opencollective.com/woodpecker-ci" target="_blank"><img src="https://opencollective.com/woodpecker-ci/backers.svg?width=890"></a>

## Usage

### .woodpecker.yml

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
- Install the needed tools in custom Docker images, use them as context

```diff
 pipeline:
   build:
-    image: debian
+    image: mycompany/image-with-awscli
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

### Plugins are straightforward

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

## Documentation

https://woodpecker-ci.org/

## Who uses Woodpecker?

Currently, I know of one organization using Woodpecker. With 50+ users, 130+ repos and more than 1100 builds a week.

Leave a [comment](https://github.com/woodpecker-ci/woodpecker/issues/122) if you're using it.

## Contribution

See [Contributing Guide](CONTRIBUTING.md)

## License

Woodpecker is Apache 2.0 licensed with the source files in this repository having a header indicating which license they are under and what copyrights apply.

Files under the `docs/` folder are licensed under Creative Commons Attribution-ShareAlike 4.0 International Public License.

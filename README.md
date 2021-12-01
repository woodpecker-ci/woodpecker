<p align="center">
  <a href="https://github.com/woodpecker-ci/woodpecker/">
    <img alt="Woodpecker" src="https://raw.githubusercontent.com/woodpecker-ci/woodpecker/master/docs/static/img/logo.svg" width="220"/>
  </a>
</p>
<br/>
<p align="center">
  <a href="https://ci.woodpecker-ci.org/woodpecker-ci/woodpecker" title="Build Status">
    <img src="https://ci.woodpecker-ci.org/api/badges/woodpecker-ci/woodpecker/status.svg">
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

[Read More](https://woodpecker-ci.org/docs/usage/intro)

### Build steps are containers

- Define any Docker image as context
- Install the needed tools in custom Docker images, use them as context

[Read More](https://woodpecker-ci.org/docs/usage/pipeline-syntax#steps)

### Plugins

Woodpecker has official plugins https://woodpecker-ci.org/plugins,
but you can also use your own.

[Read More](https://woodpecker-ci.org/docs/usage/plugins/plugins)

## Documentation

https://woodpecker-ci.org/

## Who uses Woodpecker?

[Codeberg](https://codeberg.org), the woodpecker project itself, and many others not listed.

Leave a [comment](https://github.com/woodpecker-ci/woodpecker/issues/122) if you're using it.

## Contribution

See [Contributing Guide](CONTRIBUTING.md)

## Stars over time
[![Stargazers over time](https://starchart.cc/woodpecker-ci/woodpecker.svg)](https://starchart.cc/woodpecker-ci/woodpecker)

## License

Woodpecker is Apache 2.0 licensed with the source files in this repository having a header indicating which license they are under and what copyrights apply.

Files under the `docs/` folder are licensed under Creative Commons Attribution-ShareAlike 4.0 International Public License.

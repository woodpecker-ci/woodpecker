# Woodpecker

An opinionated fork of the Drone CI system.

- Based on the v0.8 code tree
- Focused on developer experience.

[![Go Report Card](https://goreportcard.com/badge/github.com/laszlocph/woodpecker)](https://goreportcard.com/report/github.com/laszlocph/woodpecker) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

![woodpecker](docs/drone.png)

## Table of contents

- [About this fork](#about-this-fork)
  - [Motivation](#motivation)
  - [The focus of this fork](#the-focus-of-this-fork)
  - [Who uses this fork](#who-uses-this-fork)
- [Pipelines](#pipelines)
  - [Getting started](#getting-started)
  - [Pipeline documentation](#pipeline-documentation)
- [Plugins](#plugins)
  - [Custom plugins](#custom-plugins)
- [Contributing](#contributing)
- [License](#license)

## About this fork

#### Motivation

Why fork? See my [motivation](docs/motivation.md)

#### The focus of this fork

This fork is not meant to compete with Drone or reimplement its enterprise features in the open.

Instead, I'm taking a proven CI system - that Drone 0.8 is - and applying a distinct set of product ideas focusing on:

- UI experience
- the developer feedback loop
- documentation and best practices
- tighter Github integration
- Kubernetes backend

with less focus on:

- niche git systems like gitea, gogs
- computing architectures like arm64
- new pipeline formats like jsonnet

#### Who uses this fork

Currently I know of one organization using this fork. With 50+ users, 130+ repos and more than 300 builds a week.

## Pipelines

#### Getting started

Place this snippet into a file called `.drone.yml`

```yaml
pipeline:
  build:
    image: debian:stable-slim
    commands:
      - echo "This is the build step"
  a-test-step:
    image: debian:stable-slim
    commands:
      - echo "Testing.."
```

The pipeline runs on the Drone CI server and typically triggered by webhooks. One benefit of the container architecture is that it runs on your laptop too:

```sh
$ drone exec --local
stable-slim: Pulling from library/debian
a94641239323: Pull complete
Digest: sha256:d846d80f98c8aca7d3db0fadd14a0a4c51a2ce1eb2e9e14a550b3bd0c45ba941
Status: Downloaded newer image for debian:stable-slim
[build:L0:0s] + echo "This is the build step"
[build:L1:0s] This is the build step
[a-test-step:L0:0s] + echo "Testing.."
[a-test-step:L1:0s] Testing..
```

Pipeline steps are commands running in container images.
These containers are wired together and they share a volume with the source code on it.

#### Pipeline documentation

See all [pipeline features](docs/usage/pipeline.md).

## Plugins

Plugins are Docker containers that perform pre-defined tasks and are configured as steps in your pipeline. Plugins can be used to deploy code, publish artifacts, send notification, and more.

Example pipeline using the Docker and Slack plugins:

```yaml
pipeline:
  backend:
    image: golang
    commands:
      - go get
      - go build
      - go test

  docker:
    image: plugins/docker
    username: kevinbacon
    password: pa55word
    repo: foo/bar
    tags: latest

  notify:
    image: plugins/slack
    channel: developers
    username: drone
```

#### Custom plugins

Plugins are Docker containers with their entrypoint set to a predefined script.

[See how an example plugin can be implemented in a bash script](docs/usage/bash_plugin.md).

## Server setup


## Contributing

woodpecker is Apache 2.0 licensed and accepts contributions via GitHub pull requests.

[How to build the project]()

## License

woodpecker is Apache 2.0 licensed with the source files in this repository having a header indicating which license they are under and what copyrights apply.

Files under the `docs/` folder is licensed under Creative Commons Attribution-ShareAlike 4.0 International Public License. It is a derivative work of the https://github.com/drone/docs git repository.

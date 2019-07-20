# Drone-OSS-08 !

An opinionated fork of the Drone CI system.

- Based on the v0.8 code tree
- Focused on developer experience.

[![Build Status](https://cloud.drone.io/api/badges/laszlocph/drone-oss-08/status.svg)](https://cloud.drone.io/laszlocph/drone-oss-08) [![Go Report Card](https://goreportcard.com/badge/github.com/laszlocph/drone-oss-08)](https://goreportcard.com/report/github.com/laszlocph/drone-oss-08) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

![Drone-OSS-08](docs/drone.png)

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
- [Server setup](#server-setup)
  - [Quickstart](#quickstart)
  - [Authentication](#authentication)
  - [Database](#database)
  - [SSL](#ssl)
  - [Metrics](#metrics)
  - [Behind a proxy](#behind-a-proxy)
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

#### Quickstart

The below [docker-compose](https://docs.docker.com/compose/) configuration can be used to start the Drone server with a single agent. It relies on a number of environment variables that you must set before running `docker-compose up`. The variables are described below.

Each agent is able to process one build by default. If you have 4 agents installed and connected to the Drone server, your system will process 4 builds in parallel. You can add more agents to increase the number of parallel builds or set the agent's `DRONE_MAX_PROCS=1` environment variable to increase the number of parallel builds for that agent.

```yaml
version: '2'

services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
      - 80:8000
      - 9000
    volumes:
      - drone-server-data:/var/lib/drone/
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}

  drone-agent:
    image: drone/agent:{{% version %}}
    command: agent
    restart: always
    depends_on:
      - drone-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=drone-server:9000
      - DRONE_SECRET=${DRONE_SECRET}

volumes:
  drone-server-data:
```

Drone needs to know its own address. You must therefore provide the address in `<scheme>://<hostname>` format. Please omit trailing slashes.

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    environment:
      - DRONE_OPEN=true
+     - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

Drone agents require access to the host machine Docker daemon.

```diff
services:
  drone-agent:
    image: drone/agent:{{% version %}}
    command: agent
    restart: always
    depends_on: [ drone-server ]
+   volumes:
+     - /var/run/docker.sock:/var/run/docker.sock
```

Drone agents require the server address for agent-to-server communication.

```diff
services:
  drone-agent:
    image: drone/agent:{{% version %}}
    command: agent
    restart: always
    depends_on: [ drone-server ]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
+     - DRONE_SERVER=drone-server:9000
      - DRONE_SECRET=${DRONE_SECRET}
```

Drone server and agents use a shared secret to authenticate communication. This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`.

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
+     - DRONE_SECRET=${DRONE_SECRET}
  drone-agent:
    image: drone/agent:{{% version %}}
    environment:
      - DRONE_SERVER=drone-server:9000
      - DRONE_DEBUG=true
+     - DRONE_SECRET=${DRONE_SECRET}
```

Drone registration is closed by default. This example enables open registration for users that are members of approved GitHub organizations.

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    environment:
+     - DRONE_OPEN=true
+     - DRONE_ORGS=dolores,dogpatch
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

Drone administrators should also be enumerated in your configuration.

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    environment:
      - DRONE_OPEN=true
      - DRONE_ORGS=dolores,dogpatch
+     - DRONE_ADMIN=johnsmith,janedoe
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

#### Authentication

Authentication is done using OAuth and is delegated to one of multiple version control providers, configured using environment variables. The example above demonstrates basic GitHub integration.

See the complete reference for [Github](docs/administration/github.md), [Bitbucket Cloud](docs/administration/bitbucket.md), [Bitbucket Server](docs/administration/bitbucket_server.md) and [Gitlab](docs/administration/gitlab.md).

#### Database

Drone mounts a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the sqlite database.

See the [database settings](docs/administration/database.md) page to configure Postgresql or MySQL as database.

```diff
services:
  drone-server:
    image: drone/drone:{{% version %}}
    ports:
      - 80:8000
      - 9000
+   volumes:
+     - drone-server-data:/var/lib/drone/
    restart: always
```

#### SSL

Drone supports ssl configuration by mounting certificates into your container.

See the [SSL guide](docs/administration/ssl.md).

Automated [Lets Encrypt](docs/administration/lets_encrypt.md) is also supported.

#### Metrics

A [Prometheus endpoint](docs/administration/lets_encrypt.md) is exposed.

#### Behind a proxy

See the [proxy guide](docs/administration/proxy.md) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.

## Contributing

Drone-OSS-08 is Apache 2.0 licensed and accepts contributions via GitHub pull requests.

[How to build the project]()

## License

Drone-OSS-08 is Apache 2.0 licensed with the source files in this repository having a header indicating which license they are under and what copyrights apply.

Files under the `docs/` folder is licensed under Creative Commons Attribution-ShareAlike 4.0 International Public License. It is a derivative work of the https://github.com/drone/docs git repository.

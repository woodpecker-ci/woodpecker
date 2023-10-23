# Overview

A Woodpecker deployment consists of two parts:

- A server which is the heart of Woodpecker and ships the web interface.
- Next to one server you can deploy any number of agents which will run the pipelines.

> Each agent is able to process one pipeline step by default.
>
> If you have 4 agents installed and connected to the Woodpecker server, your system will process 4 builds in parallel.
>
> You can add more agents to increase the number of parallel steps or set the agent's `WOODPECKER_MAX_WORKFLOWS=1` environment variable to increase the number of parallel workflows for that agent.

## Which version of Woodpecker should I use?

Woodpecker is having two different kinds of releases: **stable** and **next**.

To find out more about the differences between the two releases, please read the [FAQ](/faq#which-version-of-woodpecker-should-i-use).

## Hardware Requirements

Below are resources requirements for Woodpecker components itself:

| Component | Memory | CPU |
| --------- | ------ | --- |
| Server    | 200 MB | 1   |
| Agent     |  32 MB | 1   |

Note, that those values do not include the operating system or workload (pipelines execution) resources consumption.

In addition you need at least some kind of database which requires additional resources depending on the selected database system.

## Installation

You can install Woodpecker on multiple ways:

- Using [docker-compose](./10-docker-compose.md) with the official [container images](./10-docker-compose.md#docker-images)
- Using [Kubernetes](./20-kubernetes.md) via the Woodpecker Helm chart
- Using binaries, DEBs or RPMs you can download from [latest release](https://github.com/woodpecker-ci/woodpecker/releases/latest)

## Authentication

Authentication is done using OAuth and is delegated to your forge which is configured by using environment variables. The example above demonstrates basic GitHub integration.

See the complete reference for all supported forges [here](../11-forges/10-overview.md).

## Database

By default Woodpecker uses a SQLite database which requires zero installation or configuration. See the [database settings](../30-database.md) page to further configure it or use MySQL or Postgres.

## SSL

Woodpecker supports SSL configuration by using Let's encrypt or by using own certificates. See the [SSL guide](../60-ssl.md).

## Metrics

A [Prometheus endpoint](../90-prometheus.md) is exposed.

## Behind a proxy

See the [proxy guide](../70-proxy.md) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.

In the case you need to use Woodpecker with a URL path prefix (like: https://example.org/woodpecker/), you can use the option [`WOODPECKER_ROOT_PATH`](../10-server-config.md#woodpecker_root_path).

## Third-party installation methods

:::info
These installation methods are not officially supported. If you experience issues with them, please open issues in the specific repositories.
:::

- Using [NixOS](./30-nixos.md) via the [NixOS module](https://search.nixos.org/options?channel=unstable&size=200&sort=relevance&query=woodpecker)
- [Using YunoHost](https://apps.yunohost.org/app/woodpecker)
- [On Cloudron](https://www.cloudron.io/store/org.woodpecker_ci.cloudronapp.html)

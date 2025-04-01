# Getting started

A Woodpecker deployment consists of two parts:

- A server which is the heart of Woodpecker and ships the web interface.
- Next to one server, you can deploy any number of agents which will run the pipelines.

Each agent is able to process one [workflow](../20-usage/15-terminology/index.md) by default. If you have 4 agents installed and connected to the Woodpecker server, your system will process four workflows (not pipelines) in parallel.

:::tip
You can add more agents to increase the number of parallel workflows or set the agent's `WOODPECKER_MAX_WORKFLOWS=1` environment variable to increase the number of parallel workflows per agent.
:::

## Which version of Woodpecker should I use?

Woodpecker is having two different kinds of releases: **stable** and **next**.

Find more information about the different versions [here](/versions).

## Hardware Requirements

Below are minimal resources requirements for Woodpecker components itself:

| Component | Memory | CPU |
| --------- | ------ | --- |
| Server    | 200 MB | 1   |
| Agent     | 32 MB  | 1   |

Note, that those values do not include the operating system or workload (pipelines execution) resource consumption.

In addition you need at least some kind of database which requires additional resources depending on the selected database system.

## Installation

You can install Woodpecker on multiple ways. If you are not sure which one to choose, we recommend using the [docker compose](./05-deployment-methods/10-docker-compose.md) method for the beginning:

- Using [docker compose](./05-deployment-methods/10-docker-compose.md) with the official [container images](./05-deployment-methods/10-docker-compose.md#docker-images)
- Using [Kubernetes](./05-deployment-methods/20-kubernetes.md) via the Woodpecker Helm chart
- Using binaries, DEBs or RPMs you can download from [latest release](https://github.com/woodpecker-ci/woodpecker/releases/latest)
- Or using a [third-party installation method](./05-deployment-methods/30-third-party.md)

## Database

By default Woodpecker uses a SQLite database which requires zero installation or configuration. See the [database settings](./10-database.md) page if you want to use a different database system like MySQL or PostgreSQL.

## Forge

What would be a CI/CD system without any code? By connecting Woodpecker to your [forge](../20-usage/15-terminology/index.md) like GitHub or Gitea you can start running pipelines on events like pushes or pull requests. Woodpecker will also use your forge for authentication and to report back the status of your pipelines. See the [forge settings](./11-forges/11-overview.md) to connect it to Woodpecker.

## Configuration

Check the [server configuration](./10-server-config.md) and [agent configuration](./15-agent-config.md) pages to see if you need to adjust any additional parts and after that you should be ready to start with [your first pipeline](../20-usage/10-intro.md).

## Agent

The agent is the worker which executes the [workflows](../20-usage/15-terminology/index.md).
Woodpecker agents can execute work using a [backend](../20-usage/15-terminology/index.md) like [docker](./22-backends/10-docker.md) or [kubernetes](./22-backends/40-kubernetes.md).
By default if you choose to deploy an agent using [docker compose](./05-deployment-methods/10-docker-compose.md) the agent simply use docker for the backend as well.
So nothing to worry about here. If you still prefer to adjust the agent to your needs, check the [agent configuration](./15-agent-config.md) page.

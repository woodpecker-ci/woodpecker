# Setup

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

To find out more about the differences between the two releases, please read the [FAQ](/faq).

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

- Using [docker-compose](#docker-compose) with the official [container images](../80-downloads.md#docker-images)
- By deploying to a [Kubernetes](./80-kubernetes.md) with manifests or Woodpeckers official Helm charts
- Using [binaries](../80-downloads.md)

### docker-compose

The below [docker-compose](https://docs.docker.com/compose/) configuration can be used to start a Woodpecker server with a single agent.

It relies on a number of environment variables that you must set before running `docker-compose up`. The variables are described below.

```yaml
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    ports:
      - 8000:8000
    volumes:
      - woodpecker-server-data:/var/lib/woodpecker/
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}

  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    command: agent
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - woodpecker-agent-config:/etc/woodpecker
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}

volumes:
  woodpecker-server-data:
  woodpecker-agent-config:
```

Woodpecker needs to know its own address. You must therefore provide the public address of it in `<scheme>://<hostname>` format. Please omit trailing slashes:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_HOST=${WOODPECKER_HOST}
```

Woodpecker can also have its port's configured. It uses a separate port for gRPC and for HTTP. The agent performs gRPC calls and connects to the gRPC port.
They can be configured with ADDR variables:

```diff
# docker-compose.yml
version: '3'
services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_GRPC_ADDR=${WOODPECKER_GRPC_ADDR}
+     - WOODPECKER_SERVER_ADDR=${WOODPECKER_HTTP_ADDR}
```

Reverse proxying can also be [configured for gRPC](./proxy#caddy). If the agents are connecting over the internet, it should also be SSL encrypted. The agent then needs to be configured to be secure:

```diff
# docker-compose.yml
version: '3'
services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_GRPC_SECURE=true # defaults to false
+     - WOODPECKER_GRPC_VERIFY=true # default
```

As agents run pipeline steps as docker containers they require access to the host machine's Docker daemon:

```diff
# docker-compose.yml
version: '3'

services:
  [...]
  woodpecker-agent:
    [...]
+   volumes:
+     - /var/run/docker.sock:/var/run/docker.sock
```

Agents require the server address for agent-to-server communication. The agent connects to the server's gRPC port:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-agent:
    [...]
    environment:
+     - WOODPECKER_SERVER=woodpecker-server:9000
```

The server and agents use a shared secret to authenticate communication. This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`:

```diff
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    [...]
    environment:
      - [...]
+     - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}
  woodpecker-agent:
    [...]
    environment:
      - [...]
+     - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}
```

## Authentication

Authentication is done using OAuth and is delegated to your forge which is configured by using environment variables. The example above demonstrates basic GitHub integration.

See the complete reference for all supported forges [here](./11-forges/10-overview.md).

## Database

By default Woodpecker uses a SQLite database which requires zero installation or configuration. See the [database settings](./30-database.md) page to further configure it or use MySQL or Postgres.

## SSL

Woodpecker supports SSL configuration by using Let's encrypt or by using own certificates. See the [SSL guide](./60-ssl.md).

## Metrics

A [Prometheus endpoint](./90-prometheus.md) is exposed.

## Behind a proxy

See the [proxy guide](./70-proxy.md) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.

In the case you need to use Woodpecker with a URL path prefix (like: https://example.org/woodpecker/), you can use the option [`WOODPECKER_ROOT_URL`](./10-server-config.md#woodpecker_root_url).

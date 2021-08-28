# Server Setup

## Installation

The below [docker-compose](https://docs.docker.com/compose/) configuration can be used to start Woodpecker with a single agent.

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
      - woodpecker-server-data:/var/lib/drone/
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}

  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    command: agent
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}

volumes:
  woodpecker-server-data:
```

> Each agent is able to process one build by default.
>
> If you have 4 agents installed and connected to the Drone server, your system will process 4 builds in parallel.
>
> You can add more agents to increase the number of parallel builds or set the agent's `WOODPECKER_MAX_PROCS=1` environment variable to increase the number of parallel builds for that agent.


Woodpecker needs to know its own address.

You must therefore provide the address in `<scheme>://<hostname>` format. Please omit trailing slashes.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
+     - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

Agents require access to the host machine's Docker daemon.

```diff
services:
  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    command: agent
    restart: always
    depends_on: [ woodpecker-server ]
+   volumes:
+     - /var/run/docker.sock:/var/run/docker.sock
```

Agents require the server address for agent-to-server communication.

```diff
services:
  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    command: agent
    restart: always
    depends_on: [ woodpecker-server ]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
+     - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

The server and agents use a shared secret to authenticate communication.

This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    environment:
      - WOODPECKER_OPEN=true
      - WOODPECKER_HOST=${WOODPECKER_HOST}
      - WOODPECKER_GITHUB=true
      - WOODPECKER_GITHUB_CLIENT=${WOODPECKER_GITHUB_CLIENT}
      - WOODPECKER_GITHUB_SECRET=${WOODPECKER_GITHUB_SECRET}
+     - WOODPECKER_SECRET=${WOODPECKER_SECRET}
  woodpecker-agent:
    image: woodpeckerci/woodpecker-agent:latest
    environment:
      - WOODPECKER_SERVER=woodpecker-server:9000
      - WOODPECKER_DEBUG=true
+     - WOODPECKER_SECRET=${WOODPECKER_SECRET}
```

## Authentication

Authentication is done using OAuth and is delegated to one of multiple version control providers, configured using environment variables. The example above demonstrates basic GitHub integration.

See the complete reference for [Github](/docs/administration/vcs/github), [Bitbucket Cloud](/docs/administration/vcs/bitbucket), [Bitbucket Server](/docs/administration/vcs/bitbucket_server) and [Gitlab](/docs/administration/vcs/gitlab).

## Database

Woodpecker mounts a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the sqlite database.

See the [database settings](/docs/administration/database) page to configure Postgresql or MySQL as database.

```diff
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:latest
    ports:
      - 80:8000
      - 9000
+   volumes:
+     - woodpecker-server-data:/var/lib/drone/
    restart: always
```

## SSL

Woodpecker supports ssl configuration by mounting certificates into your container. See the [SSL guide](/docs/administration/ssl).

Automated [Lets Encrypt](/docs/administration/lets-encrypt) is also supported.

## Metrics

A [Prometheus endpoint](/docs/administration/prometheus) is exposed.

## Behind a proxy

See the [proxy guide](/docs/administration/proxy) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.

## Deploy to Kubernetes

See the [Kubernetes guide](/docs/administration/kubernetes) if you want to deploy Woodpecker to your Kubernetes cluster.

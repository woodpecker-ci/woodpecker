## Installation

The below [docker-compose](https://docs.docker.com/compose/) configuration can be used to start Woodpecker with a single agent.

It relies on a number of environment variables that you must set before running `docker-compose up`. The variables are described below.

```yaml
# docker-compose.yml
version: '3'

services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    ports:
      - 80:8000
      - 9000
    volumes:
      - woodpecker-server-data:/var/lib/drone/
    restart: always
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}

  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    command: agent
    restart: always
    depends_on:
      - woodpecker-server
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - DRONE_SERVER=woodpecker-server:9000
      - DRONE_SECRET=${DRONE_SECRET}

volumes:
  woodpecker-server-data:
```

> Each agent is able to process one build by default.
>
> If you have 4 agents installed and connected to the Drone server, your system will process 4 builds in parallel.
>
> You can add more agents to increase the number of parallel builds or set the agent's `DRONE_MAX_PROCS=1` environment variable to increase the number of parallel builds for that agent.


Woodpecker needs to know its own address.

You must therefore provide the address in `<scheme>://<hostname>` format. Please omit trailing slashes.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
      - DRONE_OPEN=true
+     - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

Agents require access to the host machine's Docker daemon.

```diff
services:
  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
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
    image: laszlocloud/woodpecker-agent:v0.9.0
    command: agent
    restart: always
    depends_on: [ woodpecker-server ]
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
+     - DRONE_SERVER=woodpecker-server:9000
      - DRONE_SECRET=${DRONE_SECRET}
```

The server and agents use a shared secret to authenticate communication.

This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
      - DRONE_OPEN=true
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
+     - DRONE_SECRET=${DRONE_SECRET}
  woodpecker-agent:
    image: laszlocloud/woodpecker-agent:v0.9.0
    environment:
      - DRONE_SERVER=woodpecker-server:9000
      - DRONE_DEBUG=true
+     - DRONE_SECRET=${DRONE_SECRET}
```

Registration is closed by default.

This example enables open registration for users that are members of approved GitHub organizations.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    environment:
+     - DRONE_OPEN=true
+     - DRONE_ORGS=dolores,dogpatch
      - DRONE_HOST=${DRONE_HOST}
      - DRONE_GITHUB=true
      - DRONE_GITHUB_CLIENT=${DRONE_GITHUB_CLIENT}
      - DRONE_GITHUB_SECRET=${DRONE_GITHUB_SECRET}
      - DRONE_SECRET=${DRONE_SECRET}
```

Administrators should also be enumerated in your configuration.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
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


## Authentication

Authentication is done using OAuth and is delegated to one of multiple version control providers, configured using environment variables. The example above demonstrates basic GitHub integration.

See the complete reference for [Github](/administration/github), [Bitbucket Cloud](/administration/bitbucket), [Bitbucket Server](/administration/bitbucket_server) and [Gitlab](/administration/gitlab).

## Database

Woodpecker mounts a [data volume](https://docs.docker.com/storage/volumes/#create-and-manage-volumes) to persist the sqlite database.

See the [database settings](/administration/database) page to configure Postgresql or MySQL as database.

```diff
services:
  woodpecker-server:
    image: laszlocloud/woodpecker-server:v0.9.0
    ports:
      - 80:8000
      - 9000
+   volumes:
+     - woodpecker-server-data:/var/lib/drone/
    restart: always
```

## SSL

Woodpecker supports ssl configuration by mounting certificates into your container. See the [SSL guide](/administration/ssl).

Automated [Lets Encrypt](/administration/lets-encrypt) is also supported.

## Metrics

A [Prometheus endpoint](/administration/prometheus) is exposed.

## Behind a proxy

See the [proxy guide](/administration/proxy) if you want to see a setup behind Apache, Nginx, Caddy or ngrok.
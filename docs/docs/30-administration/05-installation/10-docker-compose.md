# Docker Compose

This example [docker-compose](https://docs.docker.com/compose/) setup shows the deployment of a Woodpecker instance connected to GitHub (`WOODPECKER_GITHUB=true`). If you are using another forge, please change this including the respective secret settings.

It creates persistent volumes for the server and agent config directories. The bundled SQLite DB is stored in `/var/lib/woodpecker` and is the most important part to be persisted as it holds all users and repository information.

The server uses the default port `8000` and gets exposed to the host here, so WoodpeckerWO can be accessed through this port on the host or by a reverse proxy sitting in front of it.

```yaml title="docker-compose.yaml"
services:
  woodpecker-server:
    image: woodpeckerci/woodpecker-server:v3
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
    image: woodpeckerci/woodpecker-agent:v3
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

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_HOST=${WOODPECKER_HOST}
```

Woodpecker can also have its ports configured. It uses a separate port for gRPC and for HTTP. The agent performs gRPC calls and connects to the gRPC port.
They can be configured with `*_ADDR` variables:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_GRPC_ADDR=${WOODPECKER_GRPC_ADDR}
+      - WOODPECKER_SERVER_ADDR=${WOODPECKER_HTTP_ADDR}
```

Reverse proxying can also be [configured for gRPC](../40-advanced/10-proxy.md#caddy). If the agents are connecting over the internet, it should also be SSL encrypted. The agent then needs to be configured to be secure:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_GRPC_SECURE=true # defaults to false
+      - WOODPECKER_GRPC_VERIFY=true # default
```

As agents run pipeline steps as docker containers they require access to the host machine's Docker daemon:

```diff title="docker-compose.yaml"
 services:
   [...]
   woodpecker-agent:
     [...]
+    volumes:
+      - /var/run/docker.sock:/var/run/docker.sock
```

Agents require the server address for agent-to-server communication. The agent connects to the server's gRPC port:

```diff title="docker-compose.yaml"
 services:
   woodpecker-agent:
     [...]
     environment:
+      - WOODPECKER_SERVER=woodpecker-server:9000
```

The server and agents use a shared secret to authenticate communication. This should be a random string of your choosing and should be kept private. You can generate such string with `openssl rand -hex 32`:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}
   woodpecker-agent:
     [...]
     environment:
       - [...]
+      - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}
```

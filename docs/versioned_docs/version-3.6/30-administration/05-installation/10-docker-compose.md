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
      - WOODPECKER_SERVER=woodpecker-server:8000
      - WOODPECKER_AGENT_SECRET=${WOODPECKER_AGENT_SECRET}

volumes:
  woodpecker-server-data:
  woodpecker-agent-config:
```

Woodpecker must know its own address. You must therefore specify the public address in the format `<scheme>://<hostname>`. Please omit any trailing slashes:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_HOST=${WOODPECKER_HOST}
```

It is also possible to customize the ports used. Woodpecker uses a separate port for gRPC and for HTTP. The agent makes gRPC calls and connects to the gRPC port. They can be configured with `*_ADDR` variables:

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_GRPC_ADDR=${WOODPECKER_GRPC_ADDR}
+      - WOODPECKER_SERVER_ADDR=${WOODPECKER_HTTP_ADDR}
```

If the agents establish a connection via the Internet, TLS encryption should be activated for gRPC. The agent must then be configured properly:

```diff title="docker-compose.yaml"
 services:
   woodpecker-agent:
     [...]
     environment:
       - [...]
+      - WOODPECKER_GRPC_SECURE=true # defaults to false
+      - WOODPECKER_GRPC_VERIFY=true # default
```

As agents execute pipeline steps as Docker containers, they require access to the Docker daemon of the host machine:

```diff title="docker-compose.yaml"
 services:
   [...]
   woodpecker-agent:
     [...]
+    volumes:
+      - /var/run/docker.sock:/var/run/docker.sock
```

Agents require the server address for communication between agents and servers. The agent connects to the gRPC port of the server:

```diff title="docker-compose.yaml"
 services:
   woodpecker-agent:
     [...]
     environment:
+      - WOODPECKER_SERVER=woodpecker-server:9000
```

The server and the agents use a shared secret to authenticate the communication. This should be a random string, which you should keep secret. You can create such a string with `openssl rand -hex 32`:

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

## Handling sensitive data

There are several options for handling sensitive data in `docker compose` or `docker swarm` configurations:

For Docker Compose, you can use an `.env` file next to your compose configuration to store the secrets outside the compose file. Although this separates the configuration from the secrets, it is still not very secure.

Alternatively, you can also use `docker-secrets`. As it can be difficult to use `docker-secrets` for environment variables, Woodpecker allows reading sensitive data from files by providing a `*_FILE` option for all sensitive configuration variables. Woodpecker will then attempt to read the value directly from this file. Note that the original environment variable will overwrite the value read from the file if it is specified at the same time.

```diff title="docker-compose.yaml"
 services:
   woodpecker-server:
     [...]
     environment:
       - [...]
+      - WOODPECKER_AGENT_SECRET_FILE=/run/secrets/woodpecker-agent-secret
+    secrets:
+      - woodpecker-agent-secret
+
+ secrets:
+   woodpecker-agent-secret:
+     external: true
```

To store values in a docker secret you can use the following command:

```bash
echo "my_agent_secret_key" | docker secret create woodpecker-agent-secret -
```

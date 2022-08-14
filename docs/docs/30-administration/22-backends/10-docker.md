# Docker backend

This is the original backend used with Woodpecker. The docker backend executes each step inside a separate container started on the agent.

## Configuration

### `WOODPECKER_BACKEND_DOCKER_NETWORK`
> Default: empty

Set to the name of an existing network which will be attached to all your pipeline containers (steps). Please be careful as this allows the containers of different pipelines to access each other!

### `WOODPECKER_BACKEND_DOCKER_ENABLE_IPV6`
> Default: `false`

Enable IPv6 for the networks used by pipeline containers (steps). Make sure you configured your docker daemon to support IPv6.

## Docker credentials

Woodpecker supports [Docker credentials](https://github.com/docker/docker-credential-helpers) to securely store registry credentials. Install your corresponding credential helper and configure it in your Docker config file passed via [`WOODPECKER_DOCKER_CONFIG`](/docs/administration/server-config#woodpecker_docker_config).

To add your credential helper to the Woodpecker server container you could use the following code to build a custom image:

```dockerfile
FROM woodpeckerci/woodpecker-server:latest-alpine

RUN apk add -U --no-cache docker-credential-ecr-login
```

## Podman support

While the agent was developped with Docker/Moby, Podman can also be used by setting the environment variable `DOCKER_SOCK` to point to the podman socket. In order to work without workarounds, Podman 4.0 (or above) is required.

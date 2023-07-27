# Docker backend

This is the original backend used with Woodpecker. The docker backend executes each step inside a separate container started on the agent.

## Configuration

### `WOODPECKER_BACKEND_DOCKER_NETWORK`
> Default: empty

Set to the name of an existing network which will be attached to all your pipeline containers (steps). Please be careful as this allows the containers of different pipelines to access each other!

### `WOODPECKER_BACKEND_DOCKER_ENABLE_IPV6`
> Default: `false`

Enable IPv6 for the networks used by pipeline containers (steps). Make sure you configured your docker daemon to support IPv6.

### `WOODPECKER_BACKEND_DOCKER_VOLUMES`
> Default: empty

List of default volumes separated by comma to be mounted to all pipeline containers (steps). For example to use custom CA
certificates installed on host and host timezone use `/etc/ssl/certs:/etc/ssl/certs:ro,/etc/timezone:/etc/timezone`.

## Docker credentials

Woodpecker supports [Docker credentials](https://github.com/docker/docker-credential-helpers) to securely store registry credentials. Install your corresponding credential helper and configure it in your Docker config file passed via [`WOODPECKER_DOCKER_CONFIG`](../10-server-config.md#woodpecker_docker_config).

To add your credential helper to the Woodpecker server container you could use the following code to build a custom image:

```dockerfile
FROM woodpeckerci/woodpecker-server:latest-alpine

RUN apk add -U --no-cache docker-credential-ecr-login
```

## Podman support

While the agent was developed with Docker/Moby, Podman can also be used by setting the environment variable `DOCKER_SOCK` to point to the Podman socket. In order to work without workarounds, Podman 4.0 (or above) is required.

## Image Cleanup

The agent **will not** automatically remove images from the host. This task should be managed by the host system. For example, you can use a cron job to periodically do clean-up tasks for the CI runner.

:::danger

The following commands **are destructive** and **irreversible** it is highly recommended that you test these commands on your system before running them in production via a cron job or other automation.

:::

### Remove all unused images

```sh
docker image rm $(docker images --filter "dangling=true" -q --no-trunc)
```

### Remove Woodpecker Volumes

```sh
docker volume rm $(docker volume ls --filter name=^wp_* --filter dangling=true  -q)
```

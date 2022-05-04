# Docker backend

This is the original backend used with Woodpecker. The docker backend executes each step inside a separate container started on the agent.

## Configuration

### `WOODPECKER_BACKEND_DOCKER_NETWORK`
> Default: empty

Set to the name of an existing network which will be attached to all your pipeline containers (steps). Please be careful as this allows the containers of different pipelines to access each other!

### `WOODPECKER_BACKEND_DOCKER_ENABLE_IPV6`
> Default: `false`

Enable IPv6 for the networks used by pipeline containers (steps). Make sure you configured your docker daemon to support IPv6.

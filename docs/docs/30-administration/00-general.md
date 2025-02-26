# General

Woodpecker consists of essential components (`server` and `agent`) and an optional component (`autoscaler).

The **server** provides the user interface, processes webhook requests to the underlying forge, serves the API and analyzes the pipeline configurations from the YAML files.

The **agent** executes the [workflows](../20-usage/15-terminology/index.md) via a specific [backend](../20-usage/15-terminology/index.md) (Docker, Kubernetes, local) and connects to the server via GRPC. Multiple agents can coexist so that the job limits, choice of backend and other agent-related settings can be fine-tuned for a single instance.

The **autoscaler** allows spinning up new VMs on a cloud provider of choice to process pending builds. After the builds finished, the VMs are destroyed again (after a short transition time).

:::tip
You can add more agents to increase the number of parallel workflows or set the agent's `WOODPECKER_MAX_WORKFLOWS=1` environment variable to increase the number of parallel workflows per agent.
:::

## Database

Woodpecker uses a SQLite database by default, which requires no installation or configuration. For larger instances it is recommended to use it with a Postgres or MariaDB instance. For more details take a look at the [database settings](./10-configuration/10-server.md#databases) page.

## Forge

What would a CI/CD system be without any code. By connecting Woodpecker to your [forge](../20-usage/15-terminology/index.md), you can start pipelines on events like pushes or pull requests. Woodpecker will also use your forge to authenticate and report back the status of your pipelines. For more details take a look at the [forge settings](./10-configuration/12-forges/11-overview.md) page.

## Container images

:::info
No `latest` tag exists to prevent accidental major version upgrades. Either use a SemVer tag or one of the rolling major/minor version tags. Alternatively, the `next` tag can be used for rolling builds from the `main` branch.
:::

- `vX.Y.Z`: SemVer tags for specific releases, no entrypoint shell (scratch image)
  - `vX.Y`
  - `vX`
- `vX.Y.Z-alpine`: SemVer tags for specific releases, rootless for Server and CLI (as of v3.0).
  - `vX.Y-alpine`
  - `vX-alpine`
- `next`: Built from the `main` branch
- `pull_<PR_ID>`: Images built from Pull Request branches.

Images are pushed to DockerHub and Quay.

- woodpecker-server ([DockerHub](https://hub.docker.com/repository/docker/woodpeckerci/woodpecker-server) or [Quay](https://quay.io/repository/woodpeckerci/woodpecker-server))
- woodpecker-agent ([DockerHub](https://hub.docker.com/repository/docker/woodpeckerci/woodpecker-agent) or [Quay](https://quay.io/repository/woodpeckerci/woodpecker-agent))
- woodpecker-cli ([DockerHub](https://hub.docker.com/repository/docker/woodpeckerci/woodpecker-cli) or [Quay](https://quay.io/repository/woodpeckerci/woodpecker-cli))
- woodpecker-autoscaler ([DockerHub](https://hub.docker.com/repository/docker/woodpeckerci/autoscaler))

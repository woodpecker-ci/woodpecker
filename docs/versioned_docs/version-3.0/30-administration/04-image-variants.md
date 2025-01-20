# Image variants

:::info
The `latest` tag has been deprecated as of v3.0 and will be completely removed in the future.
This was done to prevent accidental major version upgrades.
:::

- `vX.Y.Z`: SemVer tags for specific releases, no entrypoint shell (scratch image)
  - `vX.Y`
  - `vX`
- `vX.Y.Z-alpine`: SemVer tags for specific releases, based on Alpine, rootless for Server and CLI (as of v3.0).
  - `vX.Y-alpine`
  - `vX-alpine`
- `next`: Built from the `main` branch
- `pull_<PR_ID>`: Images built from Pull Request branches.

## Image registries

Images are pushed to DockerHub and Quay.

[woodpecker-server (DockerHub)](https://hub.docker.com/repository/docker/woodpeckerci/woodpecker-server)
[woodpecker-server (Quay)](https://quay.io/repository/woodpeckerci/woodpecker-server)

[woodpecker-agent (DockerHub)](https://hub.docker.com/repository/docker/woodpeckerci/woodpecker-agent)
[woodpecker-agent (Quay)](https://quay.io/repository/woodpeckerci/woodpecker-agent)

[woodpecker-cli (DockerHub)](https://hub.docker.com/repository/docker/woodpeckerci/woodpecker-cli)
[woodpecker-cli (Quay)](https://quay.io/repository/woodpeckerci/woodpecker-cli)

[woodpecker-autoscaler (DockerHub)](https://hub.docker.com/repository/docker/woodpeckerci/autoscaler)

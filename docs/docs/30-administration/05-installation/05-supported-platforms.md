# Supported platforms

Woodpecker is shipped as container images and as pre-built binaries on the [GitHub releases](https://github.com/woodpecker-ci/woodpecker/releases/latest) page. Not every component is available for every platform: the server and the Docker/Kubernetes backends are Linux-centric, while the agent and CLI run on a wider set of operating systems via the Local backend.

## Components

| Component           | Purpose                                                                                                                                                                                                                   |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `woodpecker-server` | Web UI, API, webhook receiver, pipeline scheduler.                                                                                                                                                                        |
| `woodpecker-agent`  | Executes pipeline workflows via a backend ([Docker](../10-configuration/11-backends/10-docker.md), [Kubernetes](../10-configuration/11-backends/20-kubernetes.md), [Local](../10-configuration/11-backends/30-local.md)). |
| `woodpecker-cli`    | Command-line utility for interacting with the server.                                                                                                                                                                     |
| `plugin-git`        | Default clone plugin, invoked automatically by the agent at the start of every workflow. Distributed as a container image; binaries are also published for use with the Local backend.                                    |

## Component / platform matrix

The table lists what is officially built and published by the Woodpecker project. "Image" means a container image is pushed to DockerHub and Quay. "Binary" means a pre-built tarball or `.exe` is attached to the GitHub release.

| OS / Architecture                    | server         | agent          | cli            | plugin-git     |
| ------------------------------------ | -------------- | -------------- | -------------- | -------------- |
| linux / amd64                        | Image + Binary | Image + Binary | Image + Binary | Image + Binary |
| linux / arm64 (arm64/v8)             | Image + Binary | Image + Binary | Image + Binary | Image + Binary |
| linux / arm/v7                       | Image          | Image + Binary | Image + Binary | Image + Binary |
| linux / arm/v6                       | Image          | Image          | Image          | Image          |
| linux / 386                          | Image          | Image          | Image          | Image          |
| linux / ppc64le                      | Image          | Image          | Image          | Image          |
| linux / riscv64                      | Image + Binary | Image + Binary | Image + Binary | Image          |
| linux / s390x                        | Image          | Image          | Image          | Image          |
| windows / amd64                      | Binary         | Binary         | Binary         | Binary         |
| windows / arm64                      | –              | –              | –              | Binary         |
| darwin / amd64 (macOS Intel)         | –              | Binary         | Binary         | Binary         |
| darwin / arm64 (macOS Apple Silicon) | –              | Binary         | Binary         | Binary         |
| freebsd / amd64                      | Image + Binary | Image + Binary | Image + Binary | Binary         |
| freebsd / arm64                      | –              | Image + Binary | Image + Binary | Binary         |
| openbsd / amd64                      | –              | Binary         | Binary         | Binary         |
| openbsd / arm64                      | –              | Binary         | Binary         | Binary         |

DEB and RPM packages are produced for `linux/amd64` and `linux/arm64`; see the [Distribution packages](./30-packages.md) page for download links and systemd unit examples.

## Backend support per platform

The agent can run on any platform listed above, but the available execution backends depend on the host operating system.

| Backend                                                        | Linux     | Windows                | macOS     | FreeBSD                | OpenBSD   |
| -------------------------------------------------------------- | --------- | ---------------------- | --------- | ---------------------- | --------- |
| [Docker](../10-configuration/11-backends/10-docker.md)         | Supported | Supported[^win-docker] | –         | [WIP][^freebsd-docker] | –         |
| [Kubernetes](../10-configuration/11-backends/20-kubernetes.md) | Supported | –                      | –         | –                      | –         |
| [Local](../10-configuration/11-backends/30-local.md)           | Supported | Supported              | Supported | Supported              | Supported |

[^win-docker]: Works through WSL2 with Docker Desktop, and with native Windows containers.

[^freebsd-docker]: FreeBSD Docker backend support is a work in progress; see [woodpecker-ci/woodpecker#6655](https://github.com/woodpecker-ci/woodpecker/issues/6655).

Notes:

- The **Docker** and **Kubernetes** backends require a Linux host on the agent because they rely on Linux container runtimes. On Windows, Docker is available via WSL2 or Windows containers (see footnote above). Running the agent on macOS or OpenBSD restricts you to the Local backend.
- The **Local** backend runs pipeline commands directly on the agent host with no isolation. It is the only backend available on macOS and OpenBSD, and is intended for trusted, private setups only. See the [Local backend documentation](../10-configuration/11-backends/30-local.md) for the full security notes.
- `plugin-git` is invoked as a container by default. On hosts where the Docker and Kubernetes backends are unavailable, configure the Local backend to use the [`plugin-git` binary](https://github.com/woodpecker-ci/plugin-git/releases/latest) instead, or disable the clone step and clone manually in the pipeline.

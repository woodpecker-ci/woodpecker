# CLI

`woodpecker-cli` is the command line client for the Woodpecker server. It is used to manage repositories, pipelines, secrets and other resources, and to [lint](../../20-usage/72-linter.md) or [run workflows locally](../../20-usage/73-local-execution.md). For a list of all commands and flags, see the [CLI reference](../../40-cli.md).

## Pre-built binaries

Pre-built binaries are attached to every [GitHub release](https://github.com/woodpecker-ci/woodpecker/releases/latest). The following archives are available:

| OS      | Architectures                      | Archive                                |
| ------- | ---------------------------------- | -------------------------------------- |
| Linux   | `amd64`, `arm64`, `arm`, `riscv64` | `woodpecker-cli_linux_<arch>.tar.gz`   |
| macOS   | `amd64`, `arm64`                   | `woodpecker-cli_darwin_<arch>.tar.gz`  |
| Windows | `amd64`                            | `woodpecker-cli_windows_amd64.zip`     |
| FreeBSD | `amd64`, `arm64`                   | `woodpecker-cli_freebsd_<arch>.tar.gz` |
| OpenBSD | `amd64`, `arm64`                   | `woodpecker-cli_openbsd_<arch>.tar.gz` |

Download the archive for your platform, extract it and install the `woodpecker-cli` binary into a directory in your `PATH`. This example uses Linux AMD64:

```shell
curl -fLO https://github.com/woodpecker-ci/woodpecker/releases/latest/download/woodpecker-cli_linux_amd64.tar.gz
tar -xzf woodpecker-cli_linux_amd64.tar.gz
sudo install -m 755 woodpecker-cli /usr/local/bin/
```

On Windows, extract `woodpecker-cli.exe` from the zip archive and place it in a directory that is listed in your `PATH`.

### Verifying the download

Each release contains a `checksums.txt` file with the SHA256 checksums of all release files. Download it next to the archive and verify the archive before extracting it:

```shell
curl -fLO https://github.com/woodpecker-ci/woodpecker/releases/latest/download/checksums.txt
sha256sum --check --ignore-missing checksums.txt
```

On macOS, use `shasum -a 256 --check --ignore-missing checksums.txt` instead.

## Distribution packages

DEB and RPM packages are available for `linux/amd64` and `linux/arm64`, see [Distribution packages](./30-packages.md). They install the binary to `/usr/local/bin/woodpecker-cli`.

The CLI is also available from some community-maintained packages, for example via [Homebrew](https://formulae.brew.sh/formula/woodpecker-cli).

## Container image

The CLI is published as container image `woodpeckerci/woodpecker-cli` on [DockerHub](https://hub.docker.com/r/woodpeckerci/woodpecker-cli) and [Quay](https://quay.io/repository/woodpeckerci/woodpecker-cli). The image uses `woodpecker-cli` as entrypoint and `/woodpecker` as working directory. As the container has no access to the keyring of your operating system, pass the server address and token as environment variables:

```shell
docker run --rm \
  -e WOODPECKER_SERVER=https://ci.example.com \
  -e WOODPECKER_TOKEN=<token> \
  woodpeckerci/woodpecker-cli info
```

To lint workflow files of the current directory, mount it into the working directory:

```shell
docker run --rm -v "$(pwd):/woodpecker" woodpeckerci/woodpecker-cli lint
```

The image can also be used as step image in a pipeline:

```yaml
steps:
  - name: lint
    image: woodpeckerci/woodpecker-cli
    commands:
      - woodpecker-cli lint
```

## Building from source

The CLI can be built from source with Go. Note that `go install` names the binary `cli` after the package directory and the binary will not contain release version information:

```shell
go install go.woodpecker-ci.org/woodpecker/v3/cmd/cli@latest
mv "$(go env GOPATH)/bin/cli" "$(go env GOPATH)/bin/woodpecker-cli"
```

Alternatively, run `make build-cli` in a clone of the repository, which builds the binary to `dist/woodpecker-cli`.

## First steps

Check that the CLI is installed:

```shell
woodpecker-cli --version
```

Connect the CLI to your Woodpecker server. The `setup` command asks for a personal access token, which you can find in the Woodpecker UI under your user settings in the "CLI & API" section:

```shell
woodpecker-cli setup --server https://ci.example.com
```

The connection settings are stored as a named context (`default` unless `--context` is given) in `contexts.json` in your user configuration directory, while the token is stored in the keyring of your operating system. See the [CLI configuration](../../40-cli.md#configuration) for details. If you work with multiple servers, run `setup` with a different `--context` name for each server and switch between them:

```shell
woodpecker-cli context list
woodpecker-cli context use <name>
```

Verify the connection by showing your user information:

```shell
woodpecker-cli info
```

### Shell completion

The CLI can generate completion scripts for `bash`, `zsh`, `fish` and `pwsh`. For example, to load the completion for the current bash session:

```shell
source <(woodpecker-cli completion bash)
```

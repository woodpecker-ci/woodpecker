# Installation

woodpecker-cli is available as a pre-built binary, as a DEB or RPM package and as a container image. The [supported platforms](../30-administration/05-installation/05-supported-platforms.md) page lists the operating systems and architectures it is built for. Once it is installed, the [configuration](./20-configuration.md) page explains how to connect it to your server.

## Pre-built binaries

Each [release](https://github.com/woodpecker-ci/woodpecker/releases/latest) contains an archive per platform, named `woodpecker-cli_<os>_<arch>.tar.gz`. The Windows build is packed as `woodpecker-cli_windows_amd64.zip` instead. The `checksums.txt` file of the release lists the SHA-256 checksum of every archive.

Download the archive for your platform, verify it and install the binary. This example uses Linux on amd64:

```shell
curl -fLO https://github.com/woodpecker-ci/woodpecker/releases/latest/download/woodpecker-cli_linux_amd64.tar.gz
curl -fLO https://github.com/woodpecker-ci/woodpecker/releases/latest/download/checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf woodpecker-cli_linux_amd64.tar.gz
sudo install -m 755 woodpecker-cli /usr/local/bin/
```

On macOS, verify the archive with `shasum -a 256 --ignore-missing -c checksums.txt` instead.

To update a binary installed this way, run `woodpecker-cli update`. It replaces the binary with the one from the latest release, so it needs write access to the file. This does not work on Windows, where you download the new archive yourself.

## Distribution packages

The [distribution packages](../30-administration/05-installation/30-packages.md) include a `woodpecker-cli` package for `linux/amd64`, which installs the binary to `/usr/local/bin/woodpecker-cli`. Homebrew and several Linux distributions also package the CLI. They are listed under [community packages](../30-administration/05-installation/30-packages.md#community-packages).

## Container image

The CLI is published as the `woodpeckerci/woodpecker-cli` [container image](../30-administration/00-general.md#container-images). A container has no keyring to store a token in, so pass the server and token as environment variables. This command takes both from your shell:

```shell
docker run --rm -e WOODPECKER_SERVER -e WOODPECKER_TOKEN woodpeckerci/woodpecker-cli:v3 info
```

Pipeline steps using `commands` require a shell, so use an `-alpine` image:

```yaml
steps:
  - name: lint
    image: woodpeckerci/woodpecker-cli:v3-alpine
    commands:
      - woodpecker-cli lint .woodpecker/
```

## Build from source

Run `make build-cli` in a clone of the repository. It writes the binary to `dist/woodpecker-cli`. The [development docs](../92-development/01-getting-started.md) list the tools you need.

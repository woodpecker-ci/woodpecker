// Copyright 2021 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build generate

package main

import (
	"os"

	docs "github.com/urfave/cli-docs/v3"
)

const installationDocs = `# INSTALLATION

woodpecker-cli is available as a pre-built binary, as a DEB or RPM package and as a container image. The [supported platforms](./30-administration/05-installation/05-supported-platforms.md) page lists the operating systems and architectures it is built for.

## Pre-built binaries

Each [release](https://github.com/woodpecker-ci/woodpecker/releases/latest) contains an archive per platform, named ` + "`woodpecker-cli_<os>_<arch>.tar.gz`" + `. The Windows build is packed as ` + "`woodpecker-cli_windows_amd64.zip`" + ` instead. The ` + "`checksums.txt`" + ` file of the release lists the SHA-256 checksum of every archive.

Download the archive for your platform, verify it and install the binary. This example uses Linux on amd64:

` + "```shell" + `
curl -fLO https://github.com/woodpecker-ci/woodpecker/releases/latest/download/woodpecker-cli_linux_amd64.tar.gz
curl -fLO https://github.com/woodpecker-ci/woodpecker/releases/latest/download/checksums.txt
sha256sum --ignore-missing -c checksums.txt
tar -xzf woodpecker-cli_linux_amd64.tar.gz
sudo install -m 755 woodpecker-cli /usr/local/bin/
` + "```" + `

On macOS, verify the archive with ` + "`shasum -a 256 --ignore-missing -c checksums.txt`" + ` instead.

To update a binary installed this way, run ` + "`woodpecker-cli update`" + `. It replaces the binary with the one from the latest release, so it needs write access to the file. This does not work on Windows, where you download the new archive yourself.

## Distribution packages

The [distribution packages](./30-administration/05-installation/30-packages.md) include a ` + "`woodpecker-cli`" + ` package for ` + "`linux/amd64`" + `, which installs the binary to ` + "`/usr/local/bin/woodpecker-cli`" + `. Homebrew and several Linux distributions also package the CLI. They are listed under [community packages](./30-administration/05-installation/30-packages.md#community-packages).

## Container image

The CLI is published as the ` + "`woodpeckerci/woodpecker-cli`" + ` [container image](./30-administration/00-general.md#container-images). A container has no keyring to store a token in, so pass the server and token as environment variables. This command takes both from your shell:

` + "```shell" + `
docker run --rm -e WOODPECKER_SERVER -e WOODPECKER_TOKEN woodpeckerci/woodpecker-cli:v3-alpine info
` + "```" + `

The ` + "`-alpine`" + ` images include a shell, so they also work as the image of a pipeline step:

` + "```yaml" + `
steps:
  - name: lint
    image: woodpeckerci/woodpecker-cli:v3-alpine
    commands:
      - woodpecker-cli lint .woodpecker/
` + "```" + `

## Build from source

Run ` + "`make build-cli`" + ` in a clone of the repository. It writes the binary to ` + "`dist/woodpecker-cli`" + `. The [development docs](./92-development/01-getting-started.md) list the tools you need.

## First steps

Connect the CLI to your server with ` + "`woodpecker-cli setup`" + `. It opens the Woodpecker UI in your browser, where you confirm the login, and saves the server and your token as a context. If the machine has no browser, pass the personal access token from the **CLI & API** page of your user settings with ` + "`--token`" + `. ` + "`woodpecker-cli info`" + ` then shows the user you are logged in as:

` + "```shell" + `
woodpecker-cli setup --server https://ci.example.test
woodpecker-cli info
` + "```" + `

To work with more than one server, run ` + "`woodpecker-cli setup --context <name>`" + ` once per server and switch between them with [` + "`woodpecker-cli context use`" + `](#context-ctx).

Shell completion is available for bash, zsh, fish and PowerShell. To enable it for bash, add this line to your ` + "`.bashrc`" + `:

` + "```shell" + `
source <(woodpecker-cli completion bash)
` + "```" + `

Run ` + "`woodpecker-cli completion --help`" + ` for the other shells.

`

const configurationDocs = `# CONFIGURATION

woodpecker-cli stores named contexts in a ` + "`contexts.json`" + ` file under the user's configuration directory. ` + "`XDG_CONFIG_HOME`" + ` overrides the base directory. The default paths are:

| Platform | Path |
| --- | --- |
| Linux and other Unix systems | ` + "`~/.config/woodpecker/contexts.json`" + ` |
| macOS | ` + "`~/Library/Application Support/woodpecker/contexts.json`" + ` |
| Windows | ` + "`C:\\Users\\<user>\\AppData\\Local\\woodpecker\\contexts.json`" + ` |

The file records the selected context and its non-secret connection settings:

` + "```json" + `
{
  "current_context": "production",
  "contexts": {
    "production": {
      "name": "production",
      "server_url": "https://ci.example.test",
      "log_level": "info"
    }
  }
}
` + "```" + `

` + "`current_context`" + ` must name an entry in ` + "`contexts`" + `. The ` + "`woodpecker-cli context use`" + ` command changes that selection. Authentication tokens are never written to this JSON file; they are stored in the operating system's keyring and looked up by server URL.

## Legacy configuration

If no usable current context can be loaded, woodpecker-cli falls back to the legacy ` + "`config.json`" + ` file in the same configuration directory. ` + "`--config`" + ` or ` + "`WOODPECKER_CONFIG`" + ` selects a different legacy file. The legacy JSON accepts ` + "`server_url`" + ` and ` + "`log_level`" + `; its token is also read from the operating system's keyring, not from JSON.

## Precedence

Command-line flags take precedence over their matching environment variables, and either source takes precedence over stored context or legacy values:

| Flag | Environment variable |
| --- | --- |
| ` + "`--server`" + ` | ` + "`WOODPECKER_SERVER`" + ` |
| ` + "`--token`" + ` | ` + "`WOODPECKER_TOKEN`" + ` |
| ` + "`--log-level`" + ` | ` + "`WOODPECKER_LOG_LEVEL`" + ` |

The ` + "`--config`" + ` flag and ` + "`WOODPECKER_CONFIG`" + ` only choose the legacy configuration file; they do not replace or select a context.

`

func main() {
	app := newApp()
	md, err := docs.ToMarkdown(app)
	if err != nil {
		panic(err)
	}

	fi, err := os.Create("../../docs/docs/40-cli.md")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	if _, err := fi.WriteString("# CLI\n\n" + installationDocs + configurationDocs + md); err != nil {
		panic(err)
	}
}

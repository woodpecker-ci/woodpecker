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
	if _, err := fi.WriteString("# CLI\n\n" + configurationDocs + md); err != nil {
		panic(err)
	}
}

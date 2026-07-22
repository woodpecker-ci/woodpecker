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

The CLI normally uses contexts created by ` + "`woodpecker-cli setup`" + `. Context metadata is JSON in ` + "`woodpecker/contexts.json`" + ` below the platform configuration directory:

- Linux: ` + "`$XDG_CONFIG_HOME/woodpecker/contexts.json`" + ` (defaults to ` + "`~/.config/woodpecker/contexts.json`" + `)
- macOS: ` + "`~/Library/Application Support/woodpecker/contexts.json`" + `
- Windows: ` + "`%LOCALAPPDATA%\\woodpecker\\contexts.json`" + `

The file contains ` + "`current_context`" + ` and a ` + "`contexts`" + ` object. Each context has ` + "`name`" + `, ` + "`server_url`" + `, and optionally ` + "`log_level`" + `. Authentication tokens are stored in the operating-system keyring, not in this file. For example:

` + "```json" + `
{
  "current_context": "production",
  "contexts": {
    "production": {
      "name": "production",
      "server_url": "https://ci.example.com",
      "log_level": "info"
    }
  }
}
` + "```" + `

The ` + "`--config`" + ` / ` + "`-c`" + ` flag and ` + "`WOODPECKER_CONFIG`" + ` select a legacy JSON configuration file. It is consulted only when no usable current context exists. Without the flag, its platform-relative path is ` + "`woodpecker/config.json`" + ` in the same configuration directory. It accepts ` + "`server_url`" + ` and ` + "`log_level`" + `; tokens are not read from JSON.

Command-line flags and their matching environment variables take precedence over file values. In particular, the CLI can be used without a context or config file by setting ` + "`WOODPECKER_SERVER`" + ` and ` + "`WOODPECKER_TOKEN`" + `. ` + "`WOODPECKER_LOG_LEVEL`" + ` controls the log level. Only one platform configuration location is used; files are not merged across locations.

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

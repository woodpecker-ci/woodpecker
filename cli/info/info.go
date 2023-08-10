// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package info

import (
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

// Command exports the info command.
var Command = &cli.Command{
	Name:      "info",
	Usage:     "show information about the current user",
	ArgsUsage: " ",
	Action:    info,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplInfo, true),
	),
}

func info(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	user, err := client.Self()
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, user)
}

// template for user information
var tmplInfo = `User: {{ .Login }}
Email: {{ .Email }}`

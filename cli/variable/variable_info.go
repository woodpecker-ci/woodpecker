// Copyright 2024 Woodpecker Authors
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

package variable

import (
	"context"
	"fmt"
	"html/template"
	"os"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var variableInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display variable info",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    variableInfo,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "global",
			Usage: "global variable",
		},
		common.OrgFlag,
		common.RepoFlag,
		&cli.StringFlag{
			Name:  "name",
			Usage: "variable name",
		},
		common.FormatFlag(tmplVariableList, true),
	},
}

func variableInfo(ctx context.Context, c *cli.Command) error {
	var (
		variableName = c.String("name")
		format       = c.String("format") + "\n"
	)

	if variableName == "" {
		return fmt.Errorf("variable name is missing")
	}

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	global, orgID, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	var variable *woodpecker.Variable
	switch {
	case global:
		variable, err = client.GlobalVariable(variableName)
		if err != nil {
			return err
		}
	case orgID != -1:
		variable, err = client.OrgVariable(orgID, variableName)
		if err != nil {
			return err
		}
	default:
		variable, err = client.Variable(repoID, variableName)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.New("_").Funcs(variableFuncMap).Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, variable)
}

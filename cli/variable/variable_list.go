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
	"html/template"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var variableListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list variables",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    variableList,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "global",
			Usage: "global variable",
		},
		common.OrgFlag,
		common.RepoFlag,
		common.FormatFlag(tmplVariableList, true),
	},
}

func variableList(ctx context.Context, c *cli.Command) error {
	format := c.String("format") + "\n"

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	global, orgID, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	var list []*woodpecker.Variable
	switch {
	case global:
		list, err = client.GlobalVariableList()
		if err != nil {
			return err
		}
	case orgID != -1:
		list, err = client.OrgVariableList(orgID)
		if err != nil {
			return err
		}
	default:
		list, err = client.VariableList(repoID)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.New("_").Funcs(variableFuncMap).Parse(format)
	if err != nil {
		return err
	}
	for _, registry := range list {
		if err := tmpl.Execute(os.Stdout, registry); err != nil {
			return err
		}
	}
	return nil
}

// Template for variable list items.
var tmplVariableList = "\x1b[33m{{ .Name }} \x1b[0m" + `
Value: {{ .Value }}
`

var variableFuncMap = template.FuncMap{
	"list": func(s []string) string {
		return strings.Join(s, ", ")
	},
}

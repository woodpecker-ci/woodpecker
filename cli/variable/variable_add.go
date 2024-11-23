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
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var variableCreateCmd = &cli.Command{
	Name:      "add",
	Usage:     "adds a variable",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    variableCreate,
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
		&cli.StringFlag{
			Name:  "value",
			Usage: "variable value",
		},
	},
}

func variableCreate(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	variable := &woodpecker.Variable{
		Name:  strings.ToLower(c.String("name")),
		Value: c.String("value"),
	}
	if strings.HasPrefix(variable.Value, "@") {
		path := strings.TrimPrefix(variable.Value, "@")
		out, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		variable.Value = string(out)
	}

	global, orgID, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	if global {
		_, err = client.GlobalVariableCreate(variable)
		return err
	}

	if orgID != -1 {
		_, err = client.OrgVariableCreate(orgID, variable)
		return err
	}

	_, err = client.VariableCreate(repoID, variable)
	return err
}

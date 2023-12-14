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

package secret

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

// Command exports the secret command.
var Command = &cli.Command{
	Name:  "secret",
	Usage: "manage secrets",
	Commands: []*cli.Command{
		secretCreateCmd,
		secretDeleteCmd,
		secretUpdateCmd,
		secretInfoCmd,
		secretListCmd,
	},
}

func parseTargetArgs(client woodpecker.Client, c *cli.Context) (global bool, orgID, repoID int64, err error) {
	if c.Bool("global") {
		return true, -1, -1, nil
	}

	repoIDOrFullName := c.String("repository")
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}

	orgIDOrName := c.String("organization")
	if orgIDOrName == "" && repoIDOrFullName == "" {
		if err := cli.ShowSubcommandHelp(c); err != nil {
			return false, -1, -1, err
		}

		return false, -1, -1, fmt.Errorf("missing arguments")
	}

	if orgIDOrName != "" && repoIDOrFullName == "" {
		if orgID, err := strconv.ParseInt(orgIDOrName, 10, 64); err == nil {
			return false, orgID, -1, nil
		}

		org, err := client.OrgLookup(orgIDOrName)
		if err != nil {
			return false, -1, -1, err
		}

		return false, org.ID, -1, nil
	}

	if orgIDOrName != "" && !strings.Contains(repoIDOrFullName, "/") {
		repoIDOrFullName = orgIDOrName + "/" + repoIDOrFullName
	}

	repoID, err = internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return false, -1, -1, err
	}

	return false, -1, repoID, nil
}

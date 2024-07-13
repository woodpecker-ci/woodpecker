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

package registry

import (
	"strconv"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

// Command exports the registry command set.
var Command = &cli.Command{
	Name:  "registry",
	Usage: "manage organization registries",
	Subcommands: []*cli.Command{
		registryCreateCmd,
		registryDeleteCmd,
		registryUpdateCmd,
		registryInfoCmd,
		registryListCmd,
	},
}

func parseTargetArgs(client woodpecker.Client, c *cli.Context) (orgID int64, err error) {
	orgIDOrName := c.String("organization")
	if orgIDOrName == "" {
		orgIDOrName = c.Args().First()
	}

	if orgIDOrName == "" {
		if err := cli.ShowSubcommandHelp(c); err != nil {
			return -1, err
		}
	}

	if orgID, err := strconv.ParseInt(orgIDOrName, 10, 64); err == nil {
		return orgID, nil
	}

	org, err := client.OrgLookup(orgIDOrName)
	if err != nil {
		return -1, err
	}

	return org.ID, nil
}

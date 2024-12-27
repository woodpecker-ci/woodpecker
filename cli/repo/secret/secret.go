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
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// Command exports the secret command.
var Command = &cli.Command{
	Name:  "secret",
	Usage: "manage secrets",
	Commands: []*cli.Command{
		secretCreateCmd,
		secretDeleteCmd,
		secretListCmd,
		secretShowCmd,
		secretUpdateCmd,
	},
}

func parseTargetArgs(client woodpecker.Client, c *cli.Command) (repoID int64, err error) {
	repoIDOrFullName := c.String("repository")
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}

	return internal.ParseRepo(client, repoIDOrFullName)
}

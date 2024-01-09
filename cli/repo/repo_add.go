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

package repo

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

var repoAddCmd = &cli.Command{
	Name:      "add",
	Usage:     "add a repository",
	ArgsUsage: "<forge-remote-id>",
	Action:    repoAdd,
}

func repoAdd(c *cli.Context) error {
	_forgeRemoteID := c.Args().First()
	forgeRemoteID, err := strconv.Atoi(_forgeRemoteID)
	if err != nil {
		return fmt.Errorf("invalid forge remote id: %s", _forgeRemoteID)
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	repo, err := client.RepoPost(int64(forgeRemoteID))
	if err != nil {
		return err
	}

	fmt.Printf("Successfully activated repository with forge remote %s\n", repo.FullName)
	return nil
}

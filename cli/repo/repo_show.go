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
	"context"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var repoShowCmd = &cli.Command{
	Name:      "show",
	Usage:     "show repository information",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    Show,
	Flags:     common.OutputFlags("table"),
}

func Show(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repo, err := repoShow(c, client)
	if err != nil {
		return err
	}
	return repoOutput(c, []*woodpecker.Repo{repo})
}

func repoShow(c *cli.Command, client woodpecker.Client) (*woodpecker.Repo, error) {
	repoIDOrFullName := c.Args().First()
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return nil, err
	}

	repo, err := client.Repo(repoID)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

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

var repoListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list all repos",
	ArgsUsage: " ",
	Action:    List,
	Flags: append(common.OutputFlags("table"), []cli.Flag{
		&cli.BoolFlag{
			Name:  "all",
			Usage: "query all repos, including inactive ones",
		},
	}...),
}

func List(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repos, err := repoList(c, client)
	if err != nil {
		return err
	}
	return repoOutput(c, repos)
}

func repoList(c *cli.Command, client woodpecker.Client) ([]*woodpecker.Repo, error) {
	repos := make([]*woodpecker.Repo, 0)
	opt := woodpecker.RepoListOptions{
		All: c.Bool("all"),
	}

	raw, err := client.RepoList(opt)
	if err != nil || len(raw) == 0 {
		return nil, err
	}

	org := c.String("org")
	for _, repo := range raw {
		if org != "" && org != repo.Owner {
			continue
		}
		repos = append(repos, repo)
	}
	return repos, nil
}

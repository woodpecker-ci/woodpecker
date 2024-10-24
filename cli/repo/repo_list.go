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
	"os"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var repoListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list all repos",
	ArgsUsage: " ",
	Action:    repoList,
	Flags: []cli.Flag{
		common.FormatFlag(tmplRepoList),
		&cli.StringFlag{
			Name:  "org",
			Usage: "filter by organization",
		},
		&cli.BoolFlag{
			Name:  "all",
			Usage: "query all repos, including inactive ones",
		},
	},
}

func repoList(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	opt := woodpecker.RepoListOptions{
		All: c.Bool("all"),
	}

	repos, err := client.RepoList(opt)
	if err != nil || len(repos) == 0 {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	org := c.String("org")
	for _, repo := range repos {
		if org != "" && org != repo.Owner {
			continue
		}
		if err := tmpl.Execute(os.Stdout, repo); err != nil {
			return err
		}
	}
	return nil
}

// Template for repository list items.
var tmplRepoList = "\x1b[33m{{ .FullName }}\x1b[0m (id: {{ .ID }}, forgeRemoteID: {{ .ForgeRemoteID }}, isActive: {{ .IsActive }})"

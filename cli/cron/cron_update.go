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

package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
)

var cronUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a cron job",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    cronUpdate,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:     "id",
			Usage:    "cron id",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "cron name",
		},
		&cli.StringFlag{
			Name:  "branch",
			Usage: "cron branch",
		},
		&cli.StringFlag{
			Name:  "schedule",
			Usage: "cron schedule",
		},
		common.FormatFlag(tmplCronList, true),
	),
}

func cronUpdate(c *cli.Context) error {
	var (
		repoIDOrFullName = c.String("repository")
		jobID            = c.Int64("id")
		jobName          = c.String("name")
		branch           = c.String("branch")
		schedule         = c.String("schedule")
		format           = c.String("format") + "\n"
	)
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}
	cron := &woodpecker.Cron{
		ID:       jobID,
		Name:     jobName,
		Branch:   branch,
		Schedule: schedule,
	}
	cron, err = client.CronUpdate(repoID, cron)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, cron)
}

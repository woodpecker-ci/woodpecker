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
	"context"
	"errors"
	"html/template"
	"os"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var cronUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a cron job",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    cronUpdate,
	Flags: []cli.Flag{
		common.RepoFlag,
		&cli.Int64Flag{
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
		&cli.StringSliceFlag{
			Name:    "workflow",
			Aliases: []string{"w"},
			Usage:   "run only the named workflow, repeat to select several (omit to keep the current selection)",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.BoolFlag{
			Name:  "clear-workflows",
			Usage: "drop the workflow selection so the cron runs every workflow again",
		},
		&cli.BoolFlag{
			Name:  "enabled",
			Usage: "whether cron is enabled",
			Value: true,
		},
		common.FormatFlag(tmplCronList, true),
	},
}

func cronUpdate(ctx context.Context, c *cli.Command) error {
	var (
		repoIDOrFullName = c.String("repository")
		cronID           = c.Int64("id")
		jobName          = c.String("name")
		branch           = c.String("branch")
		schedule         = c.String("schedule")
		format           = c.String("format") + "\n"
		enabled          = c.Bool("enabled")
	)
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}
	workflows, err := cronWorkflowsFromFlags(c)
	if err != nil {
		return err
	}
	cron := &woodpecker.Cron{
		ID:        cronID,
		Name:      jobName,
		Branch:    branch,
		Schedule:  schedule,
		Enabled:   enabled,
		Workflows: workflows,
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

// cronWorkflowsFromFlags turns the --workflow and --clear-workflows flags into
// the value the server's cron patch expects: nil when neither was given, so
// the selection is left alone (the field marshals as null), and an empty,
// non-nil slice for --clear-workflows, which reaches the server as [] and
// resets the cron to every workflow. A StringSliceFlag alone cannot express
// the second case: an unset flag and a cleared one both read back as nil.
func cronWorkflowsFromFlags(c *cli.Command) ([]string, error) {
	switch {
	case c.Bool("clear-workflows") && c.IsSet("workflow"):
		return nil, errors.New("--clear-workflows cannot be combined with --workflow")
	case c.Bool("clear-workflows"):
		return []string{}, nil
	case c.IsSet("workflow"):
		return c.StringSlice("workflow"), nil
	default:
		// an unset StringSliceFlag reads back as an empty slice, which would
		// marshal as [] and clear the selection, so it has to become nil here
		return nil, nil
	}
}

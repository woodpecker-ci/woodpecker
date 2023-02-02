package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var cronUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a cron job",
	ArgsUsage: "[repo/name]",
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
		reponame = c.String("repository")
		jobID    = c.Int64("id")
		jobName  = c.String("name")
		branch   = c.String("branch")
		schedule = c.String("schedule")
		format   = c.String("format") + "\n"
	)
	if reponame == "" {
		reponame = c.Args().First()
	}
	owner, name, err := internal.ParseRepo(reponame)
	if err != nil {
		return err
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	cron := &woodpecker.Cron{
		ID:       jobID,
		Name:     jobName,
		Branch:   branch,
		Schedule: schedule,
	}
	cron, err = client.CronUpdate(owner, name, cron)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, cron)
}

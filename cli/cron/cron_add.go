package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var cronCreateCmd = &cli.Command{
	Name:      "add",
	Usage:     "adds a cron",
	ArgsUsage: "[repo/name]",
	Action:    cronCreate,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:     "title",
			Usage:    "cron job title",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "branch",
			Usage: "cron job branch",
		},
		&cli.StringFlag{
			Name:     "schedule",
			Usage:    "cron job schedule",
			Required: true,
		},
		common.FormatFlag(tmplCronList, true),
	),
}

func cronCreate(c *cli.Context) error {
	var (
		title    = c.String("title")
		branch   = c.String("branch")
		schedule = c.String("schedule")
		reponame = c.String("repository")
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
	cron := &woodpecker.CronJob{
		Title:    title,
		Branch:   branch,
		Schedule: schedule,
	}
	cron, err = client.CronCreate(owner, name, cron)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, cron)
}

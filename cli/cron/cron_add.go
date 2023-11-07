package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
)

var cronCreateCmd = &cli.Command{
	Name:      "add",
	Usage:     "add a cron job",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    cronCreate,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:     "name",
			Usage:    "cron name",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "branch",
			Usage: "cron branch",
		},
		&cli.StringFlag{
			Name:     "schedule",
			Usage:    "cron schedule",
			Required: true,
		},
		common.FormatFlag(tmplCronList, true),
	),
}

func cronCreate(c *cli.Context) error {
	var (
		jobName          = c.String("name")
		branch           = c.String("branch")
		schedule         = c.String("schedule")
		repoIDOrFullName = c.String("repository")
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
		Name:     jobName,
		Branch:   branch,
		Schedule: schedule,
	}
	cron, err = client.CronCreate(repoID, cron)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, cron)
}

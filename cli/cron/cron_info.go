package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var cronInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display info about a cron job",
	ArgsUsage: "[repo/name]",
	Action:    cronInfo,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:     "id",
			Usage:    "cron id",
			Required: true,
		},
		common.FormatFlag(tmplCronList, true),
	),
}

func cronInfo(c *cli.Context) error {
	var (
		jobID    = c.Int64("id")
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
	cron, err := client.CronGet(owner, name, jobID)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, cron)
}

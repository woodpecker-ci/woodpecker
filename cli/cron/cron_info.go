package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var cronInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display info about a cron job",
	ArgsUsage: "[repo-id|repo-full-name]",
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
		jobID            = c.Int64("id")
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

	cron, err := client.CronGet(repoID, jobID)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, cron)
}

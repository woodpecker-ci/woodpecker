package cron

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var cronListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list registries",
	ArgsUsage: "[repo/name]",
	Action:    cronList,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		common.FormatFlag(tmplCronList, true),
	),
}

func cronList(c *cli.Context) error {
	var (
		format   = c.String("format") + "\n"
		reponame = c.String("repository")
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
	list, err := client.CronList(owner, name)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	for _, cron := range list {
		if err := tmpl.Execute(os.Stdout, cron); err != nil {
			return err
		}
	}
	return nil
}

// template for build list information
var tmplCronList = "\x1b[33m{{ .Title }} \x1b[0m" + `
ID: {{ .ID }}
Branch: {{ .Branch }}
Schedule: {{ .Schedule }}
NextExec: {{ .NextExec }}
`

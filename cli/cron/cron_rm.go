package cron

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var cronDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a cron",
	ArgsUsage: "[repo/name]",
	Action:    cronDelete,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:     "id",
			Usage:    "cron job id",
			Required: true,
		},
	),
}

func cronDelete(c *cli.Context) error {
	var (
		jobID    = c.Int64("id")
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
	err = client.CronDelete(owner, name, jobID)
	if err != nil {
		return err
	}

	fmt.Println("Success")
	return nil
}

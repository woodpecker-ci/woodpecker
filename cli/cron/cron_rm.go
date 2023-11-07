package cron

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var cronDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a cron job",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    cronDelete,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:     "id",
			Usage:    "cron id",
			Required: true,
		},
	),
}

func cronDelete(c *cli.Context) error {
	var (
		jobID            = c.Int64("id")
		repoIDOrFullName = c.String("repository")
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
	err = client.CronDelete(repoID, jobID)
	if err != nil {
		return err
	}

	fmt.Println("Success")
	return nil
}

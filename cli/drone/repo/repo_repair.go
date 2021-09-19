package repo

import (
	"github.com/urfave/cli"
	"github.com/woodpecker-ci/woodpecker/cli/drone/internal"
)

var repoRepairCmd = cli.Command{
	Name:      "repair",
	Usage:     "repair repository webhooks",
	ArgsUsage: "<repo/name>",
	Action:    repoRepair,
}

func repoRepair(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	return client.RepoRepair(owner, name)
}

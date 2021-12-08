package repo

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var repoRepairCmd = &cli.Command{
	Name:      "repair",
	Usage:     "repair repository webhooks",
	ArgsUsage: "<repo/name>",
	Action:    repoRepair,
	Flags:     common.GlobalFlags,
}

func repoRepair(c *cli.Context) error {
	common.SetupConsoleLogger(c)
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

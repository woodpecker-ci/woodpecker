package repo

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var repoRepairCmd = &cli.Command{
	Name:      "repair",
	Usage:     "repair repository webhooks",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    repoRepair,
	Flags:     common.GlobalFlags,
}

func repoRepair(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	if err := client.RepoRepair(repoID); err != nil {
		return err
	}

	fmt.Printf("Successfully removed repository %s\n", repoIDOrFullName)
	return nil
}

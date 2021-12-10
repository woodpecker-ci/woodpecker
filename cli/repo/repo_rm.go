package repo

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var repoRemoveCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a repository",
	ArgsUsage: "<repo/name>",
	Action:    repoRemove,
	Flags:     common.GlobalFlags,
}

func repoRemove(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	if err := client.RepoDel(owner, name); err != nil {
		return err
	}
	fmt.Printf("Successfully removed repository %s/%s\n", owner, name)
	return nil
}

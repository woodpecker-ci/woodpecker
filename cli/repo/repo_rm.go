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
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    repoRemove,
	Flags:     common.GlobalFlags,
}

func repoRemove(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	if err := client.RepoDel(repoID); err != nil {
		return err
	}
	fmt.Printf("Successfully removed repository %s\n", repoIDOrFullName)
	return nil
}

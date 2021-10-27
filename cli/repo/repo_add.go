package repo

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var repoAddCmd = &cli.Command{
	Name:      "add",
	Usage:     "add a repository",
	ArgsUsage: "<repo/name>",
	Action:    repoAdd,
	Flags:     common.GlobalFlags,
}

func repoAdd(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	if _, err := client.RepoPost(owner, name); err != nil {
		return err
	}
	fmt.Printf("Successfully activated repository %s/%s\n", owner, name)
	return nil
}

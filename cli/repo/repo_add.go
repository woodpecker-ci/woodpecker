package repo

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var repoAddCmd = &cli.Command{
	Name:      "add",
	Usage:     "add a repository",
	ArgsUsage: "<forge-remote-id>",
	Action:    repoAdd,
	Flags:     common.GlobalFlags,
}

func repoAdd(c *cli.Context) error {
	_forgeRemoteID := c.Args().First()
	forgeRemoteID, err := strconv.Atoi(_forgeRemoteID)
	if err != nil {
		return fmt.Errorf("invalid forge remote id: %s", _forgeRemoteID)
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	repo, err := client.RepoPost(int64(forgeRemoteID))
	if err != nil {
		return err
	}

	fmt.Printf("Successfully activated repository with forge remote %s\n", repo.FullName)
	return nil
}

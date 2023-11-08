package repo

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var repoChownCmd = &cli.Command{
	Name:      "chown",
	Usage:     "assume ownership of a repository",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    repoChown,
	Flags:     common.GlobalFlags,
}

func repoChown(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	repo, err := client.RepoChown(repoID)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully assumed ownership of repository %s\n", repo.FullName)
	return nil
}

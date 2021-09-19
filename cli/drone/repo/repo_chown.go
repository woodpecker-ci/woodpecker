package repo

import (
	"fmt"

	"github.com/urfave/cli"
	"github.com/woodpecker-ci/woodpecker/cli/drone/internal"
)

var repoChownCmd = cli.Command{
	Name:      "chown",
	Usage:     "assume ownership of a repository",
	ArgsUsage: "<repo/name>",
	Action:    repoChown,
}

func repoChown(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	if _, err := client.RepoChown(owner, name); err != nil {
		return err
	}
	fmt.Printf("Successfully assumed ownership of repository %s/%s\n", owner, name)
	return nil
}

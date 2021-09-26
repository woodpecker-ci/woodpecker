package user

import (
	"fmt"

	"github.com/urfave/cli"

	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var userRemoveCmd = cli.Command{
	Name:      "rm",
	Usage:     "remove a user",
	ArgsUsage: "<username>",
	Action:    userRemove,
}

func userRemove(c *cli.Context) error {
	login := c.Args().First()

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	if err := client.UserDel(login); err != nil {
		return err
	}
	fmt.Printf("Successfully removed user %s\n", login)
	return nil
}

package user

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var userRemoveCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a user",
	ArgsUsage: "<username>",
	Action:    userRemove,
	Flags:     common.GlobalFlags,
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

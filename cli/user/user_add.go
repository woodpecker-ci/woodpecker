package user

import (
	"fmt"

	"github.com/urfave/cli"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"

	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var userAddCmd = cli.Command{
	Name:      "add",
	Usage:     "adds a user",
	ArgsUsage: "<username>",
	Action:    userAdd,
}

func userAdd(c *cli.Context) error {
	login := c.Args().First()

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	user, err := client.UserPost(&woodpecker.User{Login: login})
	if err != nil {
		return err
	}
	fmt.Printf("Successfully added user %s\n", user.Login)
	return nil
}

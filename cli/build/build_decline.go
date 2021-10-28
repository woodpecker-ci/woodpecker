package build

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var buildDeclineCmd = &cli.Command{
	Name:      "decline",
	Usage:     "decline a build",
	ArgsUsage: "<repo/name> <build>",
	Action:    buildDecline,
	Flags:     common.GlobalFlags,
}

func buildDecline(c *cli.Context) (err error) {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}
	number, err := strconv.Atoi(c.Args().Get(1))
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	_, err = client.BuildDecline(owner, name, number)
	if err != nil {
		return err
	}

	fmt.Printf("Declining build %s/%s#%d\n", owner, name, number)
	return nil
}

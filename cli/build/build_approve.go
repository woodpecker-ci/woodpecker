package build

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var buildApproveCmd = &cli.Command{
	Name:      "approve",
	Usage:     "approve a build",
	ArgsUsage: "<repo/name> <build>",
	Action:    buildApprove,
	Flags:     common.GlobalFlags,
}

func buildApprove(c *cli.Context) (err error) {
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

	_, err = client.BuildApprove(owner, name, number)
	if err != nil {
		return err
	}

	fmt.Printf("Approving build %s/%s#%d\n", owner, name, number)
	return nil
}

package build

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var buildKillCmd = &cli.Command{
	Name:      "kill",
	Usage:     "force kill a build",
	ArgsUsage: "<repo/name> <build>",
	Action:    buildKill,
	Hidden:    true,
	Flags:     common.GlobalFlags,
}

func buildKill(c *cli.Context) (err error) {
	common.SetupConsoleLogger(c)
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

	err = client.BuildKill(owner, name, number)
	if err != nil {
		return err
	}

	fmt.Printf("Force killing build %s/%s#%d\n", owner, name, number)
	return nil
}

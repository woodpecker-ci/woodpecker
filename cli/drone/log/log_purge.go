package log

import (
	"fmt"
	"strconv"

	"github.com/laszlocph/woodpecker/cli/drone/internal"
	"github.com/urfave/cli"
)

var logPurgeCmd = cli.Command{
	Name:      "purge",
	Usage:     "purge a log",
	ArgsUsage: "<repo/name> <build>",
	Action:    logPurge,
}

func logPurge(c *cli.Context) (err error) {
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

	err = client.LogsPurge(owner, name, number)
	if err != nil {
		return err
	}

	fmt.Printf("Purging logs for build %s/%s#%d\n", owner, name, number)
	return nil
}

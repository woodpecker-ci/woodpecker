package pipeline

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineDeclineCmd = &cli.Command{
	Name:      "decline",
	Usage:     "decline a pipeline",
	ArgsUsage: "<repo/name> <pipeline>",
	Action:    pipelineDecline,
	Flags:     common.GlobalFlags,
}

func pipelineDecline(c *cli.Context) (err error) {
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

	_, err = client.PipelineDecline(owner, name, number)
	if err != nil {
		return err
	}

	fmt.Printf("Declining pipeline %s/%s#%d\n", owner, name, number)
	return nil
}

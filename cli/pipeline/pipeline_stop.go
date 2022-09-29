package pipeline

import (
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineStopCmd = &cli.Command{
	Name:      "stop",
	Usage:     "stop a pipeline",
	ArgsUsage: "<repo/name> [pipeline] [job]",
	Flags:     common.GlobalFlags,
	Action:    pipelineStop,
}

func pipelineStop(c *cli.Context) (err error) {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}
	number, err := strconv.Atoi(c.Args().Get(1))
	if err != nil {
		return err
	}
	job, _ := strconv.Atoi(c.Args().Get(2))
	if job == 0 {
		job = 1
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	err = client.PipelineStop(owner, name, number, job)
	if err != nil {
		return err
	}

	fmt.Printf("Stopping pipeline %s/%s#%d.%d\n", owner, name, number, job)
	return nil
}

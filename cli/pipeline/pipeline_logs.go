package pipeline

import (
	"fmt"
	"strconv"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"

	"github.com/urfave/cli/v2"
)

var pipelineLogsCmd = &cli.Command{
	Name:      "logs",
	Usage:     "show pipeline logs",
	ArgsUsage: "<repo/name> [pipeline] [job]",
	Action:    pipelineLogs,
	Flags:     common.GlobalFlags,
}

func pipelineLogs(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	number, err := strconv.Atoi(c.Args().Get(1))
	if err != nil {
		return err
	}

	job, err := strconv.Atoi(c.Args().Get(2))
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	logs, err := client.PipelineLogs(owner, name, number, job)
	if err != nil {
		return err
	}

	for _, log := range logs {
		fmt.Print(log.Output)
	}

	return nil
}

package pipeline

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineStartCmd = &cli.Command{
	Name:      "start",
	Usage:     "start a pipeline",
	ArgsUsage: "<repo/name> [pipeline]",
	Action:    pipelineStart,
	Flags: append(common.GlobalFlags,
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "custom parameters to be injected into the job environment. Format: KEY=value",
		},
	),
}

func pipelineStart(c *cli.Context) (err error) {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	pipelineArg := c.Args().Get(1)
	var number int
	if pipelineArg == "last" {
		// Fetch the pipeline number from the last pipeline
		pipeline, err := client.PipelineLast(owner, name, "")
		if err != nil {
			return err
		}
		number = pipeline.Number
	} else {
		if len(pipelineArg) == 0 {
			return errors.New("missing job number")
		}
		number, err = strconv.Atoi(pipelineArg)
		if err != nil {
			return err
		}
	}

	params := internal.ParseKeyPair(c.StringSlice("param"))

	pipeline, err := client.PipelineStart(owner, name, number, params)
	if err != nil {
		return err
	}

	fmt.Printf("Starting pipeline %s/%s#%d\n", owner, name, pipeline.Number)
	return nil
}

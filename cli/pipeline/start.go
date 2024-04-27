// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pipeline

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var pipelineStartCmd = &cli.Command{
	Name:      "start",
	Usage:     "start a pipeline",
	ArgsUsage: "<repo-id|repo-full-name> [pipeline]",
	Action:    pipelineStart,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "custom parameters to be injected into the step environment. Format: KEY=value",
		},
	},
}

func pipelineStart(ctx context.Context, c *cli.Command) (err error) {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	pipelineArg := c.Args().Get(1)
	var number int64
	if pipelineArg == "last" {
		// Fetch the pipeline number from the last pipeline
		pipeline, err := client.PipelineLast(repoID, "")
		if err != nil {
			return err
		}
		number = pipeline.Number
	} else {
		if len(pipelineArg) == 0 {
			return errors.New("missing step number")
		}
		number, err = strconv.ParseInt(pipelineArg, 10, 64)
		if err != nil {
			return err
		}
	}

	opt := woodpecker.PipelineStartOptions{
		Params: internal.ParseKeyPair(c.StringSlice("param")),
	}

	pipeline, err := client.PipelineStart(repoID, number, opt)
	if err != nil {
		return err
	}

	fmt.Printf("Starting pipeline %s#%d\n", repoIDOrFullName, pipeline.Number)
	return nil
}

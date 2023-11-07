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
	"errors"
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var pipelineStartCmd = &cli.Command{
	Name:      "start",
	Usage:     "start a pipeline",
	ArgsUsage: "<repo-id|repo-full-name> [pipeline]",
	Action:    pipelineStart,
	Flags: append(common.GlobalFlags,
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "custom parameters to be injected into the step environment. Format: KEY=value",
		},
	),
}

func pipelineStart(c *cli.Context) (err error) {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	pipelineArg := c.Args().Get(1)
	var number int
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
		number, err = strconv.Atoi(pipelineArg)
		if err != nil {
			return err
		}
	}

	params := internal.ParseKeyPair(c.StringSlice("param"))

	pipeline, err := client.PipelineStart(repoID, number, params)
	if err != nil {
		return err
	}

	fmt.Printf("Starting pipeline %s#%d\n", repoIDOrFullName, pipeline.Number)
	return nil
}

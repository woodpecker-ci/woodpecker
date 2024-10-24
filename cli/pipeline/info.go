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
	"strconv"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var pipelineInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "show pipeline details",
	ArgsUsage: "<repo-id|repo-full-name> [pipeline]",
	Action:    pipelineInfo,
	Flags:     common.OutputFlags("table"),
}

func pipelineInfo(ctx context.Context, c *cli.Command) error {
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
	if pipelineArg == "last" || len(pipelineArg) == 0 {
		// Fetch the pipeline number from the last pipeline
		pipeline, err := client.PipelineLast(repoID, woodpecker.PipelineLastOptions{})
		if err != nil {
			return err
		}
		number = pipeline.Number
	} else {
		number, err = strconv.ParseInt(pipelineArg, 10, 64)
		if err != nil {
			return err
		}
	}

	pipeline, err := client.Pipeline(repoID, number)
	if err != nil {
		return err
	}

	return pipelineOutput(c, []woodpecker.Pipeline{*pipeline})
}

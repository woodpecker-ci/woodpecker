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

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var pipelineLastCmd = &cli.Command{
	Name:      "last",
	Usage:     "show latest pipeline details",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    pipelineLast,
	Flags: append(common.OutputFlags("table"), []cli.Flag{
		&cli.StringFlag{
			Name:  "branch",
			Usage: "branch name",
			Value: "main",
		},
	}...),
}

func pipelineLast(ctx context.Context, c *cli.Command) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	pipeline, err := client.PipelineLast(repoID, c.String("branch"))
	if err != nil {
		return err
	}

	return pipelineOutput(c, []woodpecker.Pipeline{*pipeline})
}

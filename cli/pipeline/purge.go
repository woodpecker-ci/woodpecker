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
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

//nolint:mnd
func buildPipelinePurgeCmd() *cli.Command {
	return &cli.Command{
		Name:      "purge",
		Usage:     "purge pipelines",
		ArgsUsage: "<repo-id|repo-full-name>",
		Action:    Purge,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "older-than",
				Usage:    "remove pipelines older than the specified time limit",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "keep-min",
				Usage: "minimum number of pipelines to keep",
				Value: 10,
			},
		},
	}
}

func Purge(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	return pipelinePurge(c, client)
}

func pipelinePurge(c *cli.Command, client woodpecker.Client) (err error) {
	repoIDOrFullName := c.Args().First()
	if len(repoIDOrFullName) == 0 {
		return fmt.Errorf("missing required argument repo-id / repo-full-name")
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return fmt.Errorf("invalid repo '%s': %w", repoIDOrFullName, err)
	}

	olderThan := c.String("older-than")
	keepMin := c.Int("keep-min")

	duration, err := time.ParseDuration(olderThan)
	if err != nil {
		return err
	}

	opt := woodpecker.PipelineListOptions{
		ListOptions: woodpecker.ListOptions{
			Page:    1,
			PerPage: int(keepMin),
		},
	}

	pipelinesKeep, err := client.PipelineList(repoID, opt)
	if err != nil {
		return err
	}

	opt.ListOptions = woodpecker.ListOptions{}
	opt.Before = time.Now().Add(-duration)
	opt.After = time.Now()

	pipelines, err := client.PipelineList(repoID, opt)
	if err != nil {
		return err
	}

	// Create a map of pipeline IDs to keep
	keepMap := make(map[int64]struct{})
	for _, p := range pipelinesKeep {
		keepMap[p.ID] = struct{}{}
	}

	// Filter pipelines to only include those not in keepMap
	var pipelinesToPurge []*woodpecker.Pipeline
	for _, p := range pipelines {
		if _, exists := keepMap[p.ID]; !exists {
			pipelinesToPurge = append(pipelinesToPurge, p)
		}
	}

	for _, p := range pipelinesToPurge {
		err := client.PipelineDelete(repoID, p.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

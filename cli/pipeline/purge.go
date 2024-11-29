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

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	shared_utils "go.woodpecker-ci.org/woodpecker/v2/shared/utils"
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
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "disable non-read api calls",
				Value: false,
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
	dryRun := c.Bool("dry-run")

	duration, err := time.ParseDuration(olderThan)
	if err != nil {
		return err
	}

	var pipelinesKeep []*woodpecker.Pipeline

	if keepMin > 0 {
		pipelinesKeep, err = fetchPipelinesToKeep(client, repoID, int(keepMin))
		if err != nil {
			return err
		}
	}

	pipelines, err := fetchPipelines(client, repoID, duration)
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

	msgPrefix := ""
	if dryRun {
		msgPrefix = "DRY-RUN: "
	}

	for i, p := range pipelinesToPurge {
		log.Debug().Msgf("%sprune %v/%v pipelines from repo '%v'", msgPrefix, i+1, len(pipelinesToPurge), repoIDOrFullName)
		if dryRun {
			continue
		}

		err := client.PipelineDelete(repoID, p.ID)
		if err != nil {
			return err
		}
	}

	return nil
}

func fetchPipelinesToKeep(client woodpecker.Client, repoID int64, keepMin int) ([]*woodpecker.Pipeline, error) {
	if keepMin <= 0 {
		return nil, nil
	}
	return shared_utils.Paginate(func(page int) ([]*woodpecker.Pipeline, error) {
		return client.PipelineList(repoID,
			woodpecker.PipelineListOptions{
				ListOptions: woodpecker.ListOptions{
					Page: page,
				},
			},
		)
	}, keepMin)
}

func fetchPipelines(client woodpecker.Client, repoID int64, duration time.Duration) ([]*woodpecker.Pipeline, error) {
	return shared_utils.Paginate(func(page int) ([]*woodpecker.Pipeline, error) {
		return client.PipelineList(repoID,
			woodpecker.PipelineListOptions{
				ListOptions: woodpecker.ListOptions{
					Page: page,
				},
				After: time.Now().Add(-duration),
			},
		)
	}, -1)
}

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
	"time"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	shared_utils "go.woodpecker-ci.org/woodpecker/v2/shared/utils"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

//nolint:mnd
func buildPipelineListCmd() *cli.Command {
	return &cli.Command{
		Name:      "ls",
		Usage:     "show pipeline history",
		ArgsUsage: "<repo-id|repo-full-name>",
		Action:    List,
		Flags: append(common.OutputFlags("table"), []cli.Flag{
			&cli.StringFlag{
				Name:  "branch",
				Usage: "branch filter",
			},
			&cli.StringFlag{
				Name:  "event",
				Usage: "event filter",
			},
			&cli.StringFlag{
				Name:  "status",
				Usage: "status filter",
			},
			&cli.IntFlag{
				Name:  "limit",
				Usage: "limit the list size",
				Value: 25,
			},
			&cli.TimestampFlag{
				Name:  "before",
				Usage: "only return pipelines before this date (RFC3339)",
				Config: cli.TimestampConfig{
					Layouts: []string{
						time.RFC3339,
					},
				},
			},
			&cli.TimestampFlag{
				Name:  "after",
				Usage: "only return pipelines after this date (RFC3339)",
				Config: cli.TimestampConfig{
					Layouts: []string{
						time.RFC3339,
					},
				},
			},
		}...),
	}
}

func List(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	pipelines, err := pipelineList(c, client)
	if err != nil {
		return err
	}
	return pipelineOutput(c, pipelines)
}

func pipelineList(c *cli.Command, client woodpecker.Client) ([]*woodpecker.Pipeline, error) {
	repoIDOrFullName := c.Args().First()
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return nil, err
	}

	opt := woodpecker.PipelineListOptions{}

	if before := c.Timestamp("before"); !before.IsZero() {
		opt.Before = before
	}
	if after := c.Timestamp("after"); !after.IsZero() {
		opt.After = after
	}

	branch := c.String("branch")
	event := c.String("event")
	status := c.String("status")
	limit := int(c.Int("limit"))

	pipelines, err := shared_utils.Paginate(func(page int) ([]*woodpecker.Pipeline, error) {
		return client.PipelineList(repoID,
			woodpecker.PipelineListOptions{
				ListOptions: woodpecker.ListOptions{
					Page: page,
				},
				Before: opt.Before,
				After:  opt.After,
				Branch: branch,
				Events: []string{event},
				Status: status,
			},
		)
	}, limit)
	if err != nil {
		return nil, err
	}

	return pipelines, nil
}

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
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

//nolint:mnd
var pipelineListCmd = &cli.Command{
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
			Usage: "only return pipelines before this RFC3339 date",
			Config: cli.TimestampConfig{
				Layout: time.RFC3339,
			},
		},
		&cli.TimestampFlag{
			Name:  "after",
			Usage: "only return pipelines after this RFC3339 date",
			Config: cli.TimestampConfig{
				Layout: time.RFC3339,
			},
		},
	}...),
}

func List(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	resources, err := pipelineList(ctx, c, client)
	if err != nil {
		return err
	}
	return pipelineOutput(c, resources)
}

func pipelineList(_ context.Context, c *cli.Command, client woodpecker.Client) ([]woodpecker.Pipeline, error) {
	resources := make([]woodpecker.Pipeline, 0)

	repoIDOrFullName := c.Args().First()
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return resources, err
	}

	opt := woodpecker.PipelineListOptions{}
	before := c.Timestamp("before")
	after := c.Timestamp("after")

	if !before.IsZero() {
		opt.Before = before
	}
	if !after.IsZero() {
		opt.After = after
	}

	pipelines, err := client.PipelineList(repoID, opt)
	if err != nil {
		return resources, err
	}

	branch := c.String("branch")
	event := c.String("event")
	status := c.String("status")
	limit := int(c.Int("limit"))

	var count int
	for _, pipeline := range pipelines {
		if count >= limit {
			break
		}
		if branch != "" && pipeline.Branch != branch {
			continue
		}
		if event != "" && pipeline.Event != event {
			continue
		}
		if status != "" && pipeline.Status != status {
			continue
		}
		resources = append(resources, *pipeline)
		count++
	}

	return resources, nil
}

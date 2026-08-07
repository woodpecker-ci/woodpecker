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
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var pipelineCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "create new pipeline",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    pipelineCreate,
	Flags: append(common.OutputFlags("table"), []cli.Flag{
		&cli.StringFlag{
			Name:  "branch",
			Usage: "branch to create pipeline from",
		},
		&cli.StringFlag{
			Name:  "tag",
			Usage: "tag to create pipeline from",
		},
		&cli.StringFlag{
			Name:  "sha",
			Usage: "commit SHA to create pipeline from",
		},
		&cli.StringSliceFlag{
			Name:  "var",
			Usage: "key=value",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
	}...),
}

func pipelineCreate(ctx context.Context, c *cli.Command) error {
	repoIDOrFullName := c.Args().First()
	options, err := pipelineCreateOptions(c)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	variables := make(map[string]string)

	for _, vaz := range c.StringSlice("var") {
		before, after, _ := strings.Cut(vaz, "=")
		if before != "" && after != "" {
			variables[before] = after
		}
	}

	options.Variables = variables

	pipeline, err := client.PipelineCreate(repoID, options)
	if err != nil {
		return err
	}

	return pipelineOutput(c, []*woodpecker.Pipeline{pipeline})
}

func pipelineCreateOptions(c *cli.Command) (*woodpecker.PipelineOptions, error) {
	options := &woodpecker.PipelineOptions{
		Branch: c.String("branch"),
		Tag:    c.String("tag"),
		SHA:    c.String("sha"),
	}

	set := 0
	for _, ref := range []string{options.Branch, options.Tag, options.SHA} {
		if ref != "" {
			set++
		}
	}
	if set != 1 {
		return nil, errors.New("exactly one of --branch, --tag, or --sha must be set")
	}

	return options, nil
}

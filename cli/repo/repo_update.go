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

package repo

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/internal"
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
)

var repoUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a repository",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    repoUpdate,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "trusted",
			Usage: "repository is trusted",
		},
		&cli.BoolFlag{
			Name:  "gated",
			Usage: "repository is gated",
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "repository timeout",
		},
		&cli.StringFlag{
			Name:  "visibility",
			Usage: "repository visibility",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "repository configuration path (e.g. .woodpecker.yml)",
		},
		&cli.IntFlag{
			Name:  "pipeline-counter",
			Usage: "repository starting pipeline number",
		},
		&cli.BoolFlag{
			Name:  "unsafe",
			Usage: "validate updating the pipeline-counter is unsafe",
		},
	},
}

func repoUpdate(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	var (
		visibility      = c.String("visibility")
		config          = c.String("config")
		timeout         = c.Duration("timeout")
		trusted         = c.Bool("trusted")
		gated           = c.Bool("gated")
		pipelineCounter = c.Int("pipeline-counter")
		unsafe          = c.Bool("unsafe")
	)

	patch := new(woodpecker.RepoPatch)
	if c.IsSet("trusted") {
		patch.IsTrusted = &trusted
	}
	if c.IsSet("gated") {
		patch.IsGated = &gated
	}
	if c.IsSet("timeout") {
		v := int64(timeout / time.Minute)
		patch.Timeout = &v
	}
	if c.IsSet("config") {
		patch.Config = &config
	}
	if c.IsSet("visibility") {
		switch visibility {
		case "public", "private", "internal":
			patch.Visibility = &visibility
		}
	}
	if c.IsSet("pipeline-counter") && !unsafe {
		fmt.Printf("Setting the pipeline counter is an unsafe operation that could put your repository in an inconsistent state. Please use --unsafe to proceed")
	}
	if c.IsSet("pipeline-counter") && unsafe {
		patch.PipelineCounter = &pipelineCounter
	}

	repo, err := client.RepoPatch(repoID, patch)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully updated repository %s\n", repo.FullName)
	return nil
}

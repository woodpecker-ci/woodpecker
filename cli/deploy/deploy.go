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

package deploy

import (
	"context"
	"fmt"
	"html/template"
	"os"
	"strconv"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

// Command exports the deploy command.
var Command = &cli.Command{
	Name:      "deploy",
	Usage:     "trigger a pipeline with the 'deployment' event",
	ArgsUsage: "<repo-id|repo-full-name> <pipeline> <environment>",
	Action:    deploy,
	Flags: []cli.Flag{
		common.FormatFlag(tmplDeployInfo),
		&cli.StringFlag{
			Name:  "branch",
			Usage: "branch filter",
		},
		&cli.StringFlag{
			Name:  "event",
			Usage: "event filter",
			Value: woodpecker.EventPush,
		},
		&cli.StringFlag{
			Name:  "status",
			Usage: "status filter",
			Value: woodpecker.StatusSuccess,
		},
		&cli.StringSliceFlag{
			Name:    "param",
			Aliases: []string{"p"},
			Usage:   "custom parameters to be injected into the step environment. Format: KEY=value",
		},
	},
}

func deploy(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	repo := c.Args().First()
	repoID, err := internal.ParseRepo(client, repo)
	if err != nil {
		return err
	}

	branch := c.String("branch")
	event := c.String("event")
	status := c.String("status")

	if branch == "" {
		repo, err := client.Repo(repoID)
		if err != nil {
			return err
		}

		branch = repo.DefaultBranch
	}

	pipelineArg := c.Args().Get(1)
	var number int64
	if pipelineArg == "last" {
		// Fetch the pipeline number from the last pipeline
		pipelines, err := client.PipelineList(repoID, woodpecker.PipelineListOptions{})
		if err != nil {
			return err
		}
		for _, pipeline := range pipelines {
			if branch != "" && pipeline.Branch != branch {
				continue
			}
			if event != "" && pipeline.Event != event {
				continue
			}
			if status != "" && pipeline.Status != status {
				continue
			}
			if pipeline.Number > number {
				number = pipeline.Number
			}
		}
		if number == 0 {
			return fmt.Errorf("cannot deploy failure pipeline")
		}
	} else {
		number, err = strconv.ParseInt(pipelineArg, 10, 64)
		if err != nil {
			return err
		}
	}

	envArgIndex := 2
	env := c.Args().Get(envArgIndex)
	if env == "" {
		return fmt.Errorf("please specify the target environment (i.e. production)")
	}

	params := internal.ParseKeyPair(c.StringSlice("param"))

	deploy, err := client.Deploy(repoID, number, env, params)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format"))
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, deploy)
}

// Template for deployment information.
var tmplDeployInfo = `Number: {{ .Number }}
Status: {{ .Status }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
Target: {{ .Deploy }}
`

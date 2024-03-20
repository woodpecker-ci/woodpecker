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
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

//nolint:gomnd
var pipelineListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "show pipeline history",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    pipelineList,
	Flags: []cli.Flag{
		common.FormatFlag(tmplPipelineList),
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
	},
}

func pipelineList(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	pipelines, err := client.PipelineList(repoID)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	branch := c.String("branch")
	event := c.String("event")
	status := c.String("status")
	limit := c.Int("limit")

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
		if err := tmpl.Execute(os.Stdout, pipeline); err != nil {
			return err
		}
		count++
	}
	return nil
}

// template for pipeline list information
var tmplPipelineList = "\x1b[33mPipeline #{{ .Number }} \x1b[0m" + `
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
`

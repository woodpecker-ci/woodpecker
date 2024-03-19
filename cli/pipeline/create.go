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
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var pipelineCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "create new pipeline",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    pipelineCreate,
	Flags: []cli.Flag{
		common.FormatFlag(tmplPipelineList),
		&cli.StringFlag{
			Name:     "branch",
			Usage:    "branch to create pipeline from",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "var",
			Usage: "key=value",
		},
	},
}

func pipelineCreate(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	branch := c.String("branch")
	variables := make(map[string]string)

	for _, vaz := range c.StringSlice("var") {
		before, after, _ := strings.Cut(vaz, "=")
		if before != "" && after != "" {
			variables[before] = after
		}
	}

	options := &woodpecker.PipelineOptions{
		Branch:    branch,
		Variables: variables,
	}

	pipeline, err := client.PipelineCreate(repoID, options)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, pipeline)
}

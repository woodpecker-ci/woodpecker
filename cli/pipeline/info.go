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
	"strconv"
	"text/template"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

var pipelineInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "show pipeline details",
	ArgsUsage: "<repo-id|repo-full-name> [pipeline]",
	Action:    pipelineInfo,
	Flags:     []cli.Flag{common.FormatFlag(tmplPipelineInfo)},
}

func pipelineInfo(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}
	pipelineArg := c.Args().Get(1)

	var number int64
	if pipelineArg == "last" || len(pipelineArg) == 0 {
		// Fetch the pipeline number from the last pipeline
		pipeline, err := client.PipelineLast(repoID, "")
		if err != nil {
			return err
		}
		number = pipeline.Number
	} else {
		number, err = strconv.ParseInt(pipelineArg, 10, 64)
		if err != nil {
			return err
		}
	}

	pipeline, err := client.Pipeline(repoID, number)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format"))
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, pipeline)
}

// template for pipeline information
var tmplPipelineInfo = `Number: {{ .Number }}
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Message: {{ .Message }}
Author: {{ .Author }}
`

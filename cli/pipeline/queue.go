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
	"fmt"
	"os"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

var pipelineQueueCmd = &cli.Command{
	Name:      "queue",
	Usage:     "show pipeline queue",
	ArgsUsage: " ",
	Action:    pipelineQueue,
	Flags:     []cli.Flag{common.FormatFlag(tmplPipelineQueue)},
}

func pipelineQueue(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	pipelines, err := client.PipelineQueue()
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		fmt.Println("there are no pending or running pipelines")
		return nil
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	for _, pipeline := range pipelines {
		if err := tmpl.Execute(os.Stdout, pipeline); err != nil {
			return err
		}
	}
	return nil
}

// template for pipeline list information
var tmplPipelineQueue = "\x1b[33m{{ .FullName }} #{{ .Number }} \x1b[0m" + `
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
`

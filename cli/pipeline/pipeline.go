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
	"io"
	"os"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/output"
	"go.woodpecker-ci.org/woodpecker/v3/cli/pipeline/deploy"
	"go.woodpecker-ci.org/woodpecker/v3/cli/pipeline/log"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// Command exports the pipeline command set.
var Command = &cli.Command{
	Name:  "pipeline",
	Usage: "manage pipelines",
	Commands: []*cli.Command{
		pipelineApproveCmd,
		pipelineCreateCmd,
		pipelineDeclineCmd,
		deploy.Command,
		pipelineKillCmd,
		pipelineLastCmd,
		buildPipelineListCmd(),
		log.Command,
		pipelinePsCmd,
		pipelinePurgeCmd,
		pipelineQueueCmd,
		pipelineShowCmd,
		pipelineStartCmd,
		pipelineStopCmd,
	},
}

func pipelineOutput(c *cli.Command, pipelines []*woodpecker.Pipeline, fd ...io.Writer) error {
	outFmt, outOpt := output.ParseOutputOptions(c.String("output"))
	noHeader := c.Bool("output-no-headers")

	var out io.Writer
	switch len(fd) {
	case 0:
		out = os.Stdout
	case 1:
		out = fd[0]
	default:
		out = os.Stdout
	}

	switch outFmt {
	case "go-template":
		if len(outOpt) < 1 {
			return fmt.Errorf("%w: missing template", output.ErrOutputOptionRequired)
		}

		tmpl, err := template.New("_").Parse(outOpt[0] + "\n")
		if err != nil {
			return err
		}
		if err := tmpl.Execute(out, pipelines); err != nil {
			return err
		}
	case "table":
		fallthrough
	default:
		table := output.NewTable(out)
		cols := []string{"Number", "Status", "Event", "Branch", "Message", "Author"}

		if len(outOpt) > 0 {
			cols = outOpt
		}
		if !noHeader {
			table.WriteHeader(cols)
		}
		for _, resource := range pipelines {
			// TODO get message from commit
			if err := table.Write(cols, resource); err != nil {
				return err
			}
		}
		table.Flush()
	}

	return nil
}

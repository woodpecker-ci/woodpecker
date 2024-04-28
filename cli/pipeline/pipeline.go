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

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/output"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

// Command exports the pipeline command set.
var Command = &cli.Command{
	Name:  "pipeline",
	Usage: "manage pipelines",
	Subcommands: []*cli.Command{
		pipelineListCmd,
		pipelineLastCmd,
		pipelineLogsCmd,
		pipelineInfoCmd,
		pipelineStopCmd,
		pipelineStartCmd,
		pipelineApproveCmd,
		pipelineDeclineCmd,
		pipelineQueueCmd,
		pipelineKillCmd,
		pipelinePsCmd,
		pipelineCreateCmd,
	},
}

func pipelineOutput(c *cli.Context, resources []woodpecker.Pipeline) error {
	outfmt, outopt := output.ParseOutputOptions(c.String("output"))
	noHeader := c.Bool("no-header")

	switch outfmt {
	case "go-template":
		if len(outopt) < 1 {
			return fmt.Errorf("%w: missing template", output.ErrOutputOptionRequired)
		}

		tmpl, err := template.New("_").Parse(outopt[0] + "\n")
		if err != nil {
			return err
		}
		if err := tmpl.Execute(os.Stdout, resources); err != nil {
			return err
		}
	case "table":
		fallthrough
	default:
		table := output.NewTable()
		cols := []string{"Number", "Status", "Event", "Branch", "Commit", "Author"}

		if len(outopt) > 0 {
			cols = outopt
		}
		if !noHeader {
			table.WriteHeader(cols)
		}
		for _, resource := range resources {
			if err := table.Write(cols, resource); err != nil {
				return err
			}
		}
		table.Flush()
	}

	return nil
}

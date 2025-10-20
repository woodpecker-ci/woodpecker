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

package log

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var logShowCmd = &cli.Command{
	Name:      "show",
	Usage:     "show pipeline logs",
	ArgsUsage: "<repo-id|repo-full-name> <pipeline> [step-number|step-name]",
	Action:    logShow,
}

func logShow(ctx context.Context, c *cli.Command) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	if len(repoIDOrFullName) == 0 {
		return fmt.Errorf("missing required argument repo-id / repo-full-name")
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return fmt.Errorf("invalid repo '%s': %w ", repoIDOrFullName, err)
	}

	pipelineArg := c.Args().Get(1)
	if len(pipelineArg) == 0 {
		return fmt.Errorf("missing required argument pipeline")
	}
	number, err := strconv.ParseInt(pipelineArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid pipeline '%s': %w", pipelineArg, err)
	}

	stepArg := c.Args().Get(2) //nolint:mnd
	if len(stepArg) == 0 {
		return pipelineLog(client, repoID, number)
	}

	step, err := internal.ParseStep(client, repoID, number, stepArg)
	if err != nil {
		return fmt.Errorf("invalid step '%s': %w", stepArg, err)
	}
	return stepLog(client, repoID, number, step)
}

func pipelineLog(client woodpecker.Client, repoID, number int64) error {
	pipeline, err := client.Pipeline(repoID, number)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(tmplPipelineLogs + "\n")
	if err != nil {
		return err
	}

	for _, workflow := range pipeline.Workflows {
		for _, step := range workflow.Children {
			if err := tmpl.Execute(os.Stdout, map[string]any{"workflow": workflow, "step": step}); err != nil {
				return err
			}
			err := stepLog(client, repoID, number, step.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func stepLog(client woodpecker.Client, repoID, number, step int64) error {
	logs, err := client.StepLogEntries(repoID, number, step)
	if err != nil {
		return err
	}

	for _, log := range logs {
		fmt.Println(string(log.Data))
	}

	return nil
}

// template for pipeline ps information.
var tmplPipelineLogs = "\x1b[33m{{ .workflow.Name }} > {{ .step.Name }} (#{{ .step.PID }}):\x1b[0m"

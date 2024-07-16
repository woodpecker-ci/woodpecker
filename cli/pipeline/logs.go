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
	"strconv"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var pipelineLogsCmd = &cli.Command{
	Name:      "logs",
	Usage:     "show pipeline logs",
	ArgsUsage: "<repo-id|repo-full-name> <pipeline> [stepID]",
	Action:    pipelineLogs,
}

func pipelineLogs(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
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

	stepArg := c.Args().Get(2)
	if len(stepArg) == 0 {
		return showPipelineLog(client, repoID, number)
	}

	step, err := strconv.ParseInt(stepArg, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid stepId '%s': %w", stepArg, err)
	}
	return showStepLog(client, repoID, number, step)
}

func showPipelineLog(client woodpecker.Client, repoID, number int64) error {
	pipeline, err := client.Pipeline(repoID, number)
	if err != nil {
		return err
	}

	for _, workflow := range pipeline.Workflows {
		for _, step := range workflow.Children {
			fmt.Printf("\x1b[33mWorflow #%d\x1b[0m ('%s'), \x1b[33mStep #%d\x1b[0m ('%s'):\n", workflow.PID, workflow.Name, step.PID, step.Name)
			err := showStepLog(client, repoID, number, step.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func showStepLog(client woodpecker.Client, repoID, number, step int64) error {
	logs, err := client.StepLogEntries(repoID, number, step)
	if err != nil {
		return err
	}

	for _, log := range logs {
		fmt.Println(string(log.Data))
	}

	return nil
}

// template for pipeline ps information
var tmplPipelineLogs = "\x1b[33m{{ .Workflow.Name }} > {{ .Step.Name }} (#{{ .Step.PID }}):\x1b[0m"

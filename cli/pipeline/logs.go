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

	"go.woodpecker-ci.org/woodpecker/cli/internal"

	"github.com/urfave/cli/v2"
)

var pipelineLogsCmd = &cli.Command{
	Name:      "logs",
	Usage:     "show pipeline logs",
	ArgsUsage: "<repo-id|repo-full-name> [pipeline] [stepID]",
	Action:    pipelineLogs,
}

func pipelineLogs(c *cli.Context) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	number, err := strconv.ParseInt(c.Args().Get(1), 10, 64)
	if err != nil {
		return err
	}

	step, err := strconv.ParseInt(c.Args().Get(2), 10, 64)
	if err != nil {
		return err
	}

	logs, err := client.StepLogEntries(repoID, number, step)
	if err != nil {
		return err
	}

	for _, log := range logs {
		fmt.Print(string(log.Data))
	}

	return nil
}

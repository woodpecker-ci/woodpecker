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

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var pipelineKillCmd = &cli.Command{
	Name:      "kill",
	Usage:     "force kill a pipeline",
	ArgsUsage: "<repo-id|repo-full-name> <pipeline>",
	Action:    pipelineKill,
	Hidden:    true,
	Flags:     common.GlobalFlags,
}

func pipelineKill(c *cli.Context) (err error) {
	number, err := strconv.ParseInt(c.Args().Get(1), 10, 64)
	if err != nil {
		return err
	}

	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	err = client.PipelineKill(repoID, number)
	if err != nil {
		return err
	}

	fmt.Printf("Force killing pipeline %s#%d\n", repoIDOrFullName, number)
	return nil
}

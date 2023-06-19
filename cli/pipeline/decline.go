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

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineDeclineCmd = &cli.Command{
	Name:      "decline",
	Usage:     "decline a pipeline",
	ArgsUsage: "<repo-id|repo-full-name> <pipeline>",
	Action:    pipelineDecline,
	Flags:     common.GlobalFlags,
}

func pipelineDecline(c *cli.Context) (err error) {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	number, err := strconv.Atoi(c.Args().Get(1))
	if err != nil {
		return err
	}

	_, err = client.PipelineDecline(repoID, number)
	if err != nil {
		return err
	}

	fmt.Printf("Declining pipeline %s#%d\n", repoIDOrFullName, number)
	return nil
}

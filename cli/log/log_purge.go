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
	"fmt"
	"strconv"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

var logPurgeCmd = &cli.Command{
	Name:      "purge",
	Usage:     "purge a log",
	ArgsUsage: "<repo-id|repo-full-name> <pipeline>",
	Action:    logPurge,
}

func logPurge(c *cli.Context) (err error) {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoIDOrFullName := c.Args().First()
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}
	number, err := strconv.ParseInt(c.Args().Get(1), 10, 64)
	if err != nil {
		return err
	}

	err = client.LogsPurge(repoID, number)
	if err != nil {
		return err
	}

	fmt.Printf("Purging logs for pipeline %s#%d\n", repoIDOrFullName, number)
	return nil
}

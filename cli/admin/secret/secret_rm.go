// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package secret

import (
	"context"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

var secretDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a secret",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    secretDelete,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "name",
			Usage: "secret name",
		},
	},
}

func secretDelete(ctx context.Context, c *cli.Command) error {
	secretName := c.String("name")

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	return client.GlobalSecretDelete(secretName)
}

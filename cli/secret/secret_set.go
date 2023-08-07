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
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var secretUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a secret",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    secretUpdate,
	Flags: append(common.GlobalFlags,
		&cli.BoolFlag{
			Name:  "global",
			Usage: "global secret",
		},
		common.OrgFlag,
		common.RepoFlag,
		&cli.StringFlag{
			Name:  "name",
			Usage: "secret name",
		},
		&cli.StringFlag{
			Name:  "value",
			Usage: "secret value",
		},
		&cli.StringSliceFlag{
			Name:  "event",
			Usage: "secret limited to these events",
		},
		&cli.StringSliceFlag{
			Name:  "image",
			Usage: "secret limited to these images",
		},
		&cli.BoolFlag{
			Name:  "plugins-only",
			Usage: "secret limited to plugins",
		},
	),
}

func secretUpdate(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	secret := &woodpecker.Secret{
		Name:        strings.ToLower(c.String("name")),
		Value:       c.String("value"),
		Images:      c.StringSlice("image"),
		PluginsOnly: c.Bool("plugins-only"),
		Events:      c.StringSlice("event"),
	}
	if strings.HasPrefix(secret.Value, "@") {
		path := strings.TrimPrefix(secret.Value, "@")
		out, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		secret.Value = string(out)
	}

	global, orgID, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	if global {
		_, err = client.GlobalSecretUpdate(secret)
		return err
	}
	if orgID != -1 {
		_, err = client.OrgSecretUpdate(orgID, secret)
		return err
	}
	_, err = client.SecretUpdate(repoID, secret)
	return err
}

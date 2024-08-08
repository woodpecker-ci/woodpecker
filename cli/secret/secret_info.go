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
	"fmt"
	"html/template"
	"os"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var secretInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display secret info",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    secretInfo,
	Flags: []cli.Flag{
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
		common.FormatFlag(tmplSecretList, true),
	},
}

func secretInfo(ctx context.Context, c *cli.Command) error {
	var (
		secretName = c.String("name")
		format     = c.String("format") + "\n"
	)

	if secretName == "" {
		return fmt.Errorf("secret name is missing")
	}

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	global, orgID, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	var secret *woodpecker.Secret
	switch {
	case global:
		secret, err = client.GlobalSecret(secretName)
		if err != nil {
			return err
		}
	case orgID != -1:
		secret, err = client.OrgSecret(orgID, secretName)
		if err != nil {
			return err
		}
	default:
		secret, err = client.Secret(repoID, secretName)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, secret)
}

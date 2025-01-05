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
	"html/template"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var secretListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list secrets",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    secretList,
	Flags: []cli.Flag{
		common.RepoFlag,
		common.FormatFlag(tmplSecretList, true),
	},
}

func secretList(ctx context.Context, c *cli.Command) error {
	format := c.String("format") + "\n"

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	opt := woodpecker.SecretListOptions{}

	list, err := client.SecretList(repoID, opt)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	for _, secret := range list {
		if err := tmpl.Execute(os.Stdout, secret); err != nil {
			return err
		}
	}
	return nil
}

// Template for secret list items.
var tmplSecretList = "\x1b[33m{{ .Name }} \x1b[0m" + `
Events: {{ list .Events }}
{{- if .Images }}
Images: {{ list .Images }}
{{- else }}
Images: <any>
{{- end }}
`

var secretFuncMap = template.FuncMap{
	"list": func(s []string) string {
		return strings.Join(s, ", ")
	},
}

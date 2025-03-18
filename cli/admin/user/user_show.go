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

package user

import (
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
)

var userShowCmd = &cli.Command{
	Name:      "show",
	Usage:     "show user information",
	ArgsUsage: "<username>",
	Action:    userShow,
	Flags:     []cli.Flag{common.FormatFlag(tmplUserInfo, false)},
}

func userShow(ctx context.Context, c *cli.Command) error {
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	login := c.Args().First()
	if len(login) == 0 {
		return fmt.Errorf("missing or invalid user login")
	}

	user, err := client.User(login)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, user)
}

// Template for user information.
var tmplUserInfo = `User: {{ .Login }}
Email: {{ .Email }}`

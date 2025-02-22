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

package repo

import (
	"fmt"
	"io"
	"os"
	"text/template"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/output"
	"go.woodpecker-ci.org/woodpecker/v3/cli/repo/cron"
	"go.woodpecker-ci.org/woodpecker/v3/cli/repo/registry"
	"go.woodpecker-ci.org/woodpecker/v3/cli/repo/secret"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// Command exports the repository command.
var Command = &cli.Command{
	Name:  "repo",
	Usage: "manage repositories",
	Commands: []*cli.Command{
		repoAddCmd,
		repoChownCmd,
		cron.Command,
		repoListCmd,
		registry.Command,
		repoRemoveCmd,
		repoRepairCmd,
		secret.Command,
		repoShowCmd,
		repoSyncCmd,
		repoUpdateCmd,
	},
}

func repoOutput(c *cli.Command, repos []*woodpecker.Repo, fd ...io.Writer) error {
	outFmt, outOpt := output.ParseOutputOptions(c.String("output"))
	noHeader := c.Bool("output-no-headers")

	var out io.Writer
	switch len(fd) {
	case 0:
		out = os.Stdout
	case 1:
		out = fd[0]
	default:
		out = os.Stdout
	}

	switch outFmt {
	case "go-template":
		if len(outOpt) < 1 {
			return fmt.Errorf("%w: missing template", output.ErrOutputOptionRequired)
		}

		tmpl, err := template.New("_").Parse(outOpt[0] + "\n")
		if err != nil {
			return err
		}
		if err := tmpl.Execute(out, repos); err != nil {
			return err
		}
	case "table":
		fallthrough
	default:
		table := output.NewTable(out)

		// Add custom field mapping for nested Trusted fields
		table.AddFieldFn("TrustedNetwork", func(obj any) string {
			repo, ok := obj.(*woodpecker.Repo)
			if !ok {
				return ""
			}
			return output.YesNo(repo.Trusted.Network)
		})
		table.AddFieldFn("TrustedSecurity", func(obj any) string {
			repo, ok := obj.(*woodpecker.Repo)
			if !ok {
				return ""
			}
			return output.YesNo(repo.Trusted.Security)
		})
		table.AddFieldFn("TrustedVolume", func(obj any) string {
			repo, ok := obj.(*woodpecker.Repo)
			if !ok {
				return ""
			}
			return output.YesNo(repo.Trusted.Volumes)
		})

		table.AddFieldAlias("Is_Active", "Active")
		table.AddFieldAlias("Is_SCM_Private", "SCM_Private")

		cols := []string{"Full_Name", "Branch", "Forge_URL", "Visibility", "SCM_Private", "Active", "Allow_Pull"}

		if len(outOpt) > 0 {
			cols = outOpt
		}
		if !noHeader {
			table.WriteHeader(cols)
		}
		for _, resource := range repos {
			if err := table.Write(cols, resource); err != nil {
				return err
			}
		}
		table.Flush()
	}

	return nil
}

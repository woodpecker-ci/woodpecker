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

package registry

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
)

var registryInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display registry info",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    registryInfo,
	Flags: []cli.Flag{
		common.RepoFlag,
		&cli.StringFlag{
			Name:  "hostname",
			Usage: "registry hostname",
			Value: "docker.io",
		},
		common.FormatFlag(tmplRegistryList, true),
	},
}

func registryInfo(c *cli.Context) error {
	var (
		hostname = c.String("hostname")
		format   = c.String("format") + "\n"
	)

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	registry, err := client.Registry(repoID, hostname)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, registry)
}

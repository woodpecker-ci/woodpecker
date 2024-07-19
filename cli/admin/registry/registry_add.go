// Copyright 2024 Woodpecker Authors
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
	"context"
	"os"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

var registryCreateCmd = &cli.Command{
	Name:   "add",
	Usage:  "adds a registry",
	Action: registryCreate,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "hostname",
			Usage: "registry hostname",
			Value: "docker.io",
		},
		&cli.StringFlag{
			Name:  "username",
			Usage: "registry username",
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "registry password",
		},
	},
}

func registryCreate(ctx context.Context, c *cli.Command) error {
	var (
		hostname = c.String("hostname")
		username = c.String("username")
		password = c.String("password")
	)

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	registry := &woodpecker.Registry{
		Address:  hostname,
		Username: username,
		Password: password,
	}
	if strings.HasPrefix(registry.Password, "@") {
		path := strings.TrimPrefix(registry.Password, "@")
		out, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		registry.Password = string(out)
	}

	_, err = client.GlobalRegistryCreate(registry)
	return err
}

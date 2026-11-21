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

package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var repoUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a repository",
	ArgsUsage: "<repo-id|repo-full-name>",
	Action:    repoUpdate,
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "trusted-security",
			Usage: "repository is security trusted",
		},
		&cli.BoolFlag{
			Name:  "trusted-volumes",
			Usage: "repository is volumes trusted",
		},
		&cli.BoolFlag{
			Name:  "trusted-network",
			Usage: "repository is network trusted",
		},
		&cli.BoolFlag{
			Name:   "trusted", // TODO: remove in next release
			Usage:  "repository is trusted",
			Hidden: true,
		},
		&cli.BoolFlag{
			Name:   "gated", // TODO: remove in next release
			Hidden: true,
		},
		&cli.StringFlag{
			Name:  "require-approval",
			Usage: "repository requires approval for",
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "repository timeout",
		},
		&cli.StringFlag{
			Name:  "visibility",
			Usage: "repository visibility",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "repository configuration path. Example: .woodpecker.yml",
		},
	},
}

func repoUpdate(ctx context.Context, c *cli.Command) error {
	repoIDOrFullName := c.Args().First()
	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}

	var (
		visibility      = c.String("visibility")
		config          = c.String("config")
		timeout         = c.Duration("timeout")
		requireApproval = c.String("require-approval")
	)

	patch := new(woodpecker.RepoPatch)
	// TODO remove in next release
	if c.IsSet("trusted") {
		trusted := c.Bool("trusted")
		patch.Trusted = &woodpecker.TrustedConfigurationPatch{
			Network:  &trusted,
			Security: &trusted,
			Volumes:  &trusted,
		}
	}
	if c.IsSet("trusted-security") || c.IsSet("trusted-network") || c.IsSet("trusted-volumes") {
		patch.Trusted = new(woodpecker.TrustedConfigurationPatch)

		if c.IsSet("trusted-security") {
			t := c.Bool("trusted-security")
			patch.Trusted.Security = &t
		}
		if c.IsSet("trusted-network") {
			t := c.Bool("trusted-network")
			patch.Trusted.Network = &t
		}
		if c.IsSet("trusted-volumes") {
			t := c.Bool("trusted-volumes")
			patch.Trusted.Volumes = &t
		}
	}

	// TODO: remove in next release
	if c.IsSet("gated") {
		return fmt.Errorf("'gated' option has been set in version 2.8, use 'require-approval' in >= 3.0")
	}

	if c.IsSet("require-approval") {
		if mode := woodpecker.ApprovalMode(requireApproval); mode.Valid() {
			patch.RequireApproval = &mode
		} else {
			return fmt.Errorf("update approval mode failed: '%s' is no valid mode", mode)
		}
	}
	if c.IsSet("timeout") {
		v := int64(timeout / time.Minute)
		patch.Timeout = &v
	}
	if c.IsSet("config") {
		patch.Config = &config
	}
	if c.IsSet("visibility") {
		switch visibility {
		case "public", "private", "internal":
			patch.Visibility = &visibility
		}
	}

	repo, err := client.RepoPatch(repoID, patch)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully updated repository %s\n", repo.FullName)
	return nil
}

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
		&cli.StringFlag{
			Name:  "organization",
			Usage: "organization name (e.g. octocat)",
		},
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

	global, owner, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	if global {
		_, err = client.GlobalSecretUpdate(secret)
		return err
	}
	if owner != "" {
		_, err = client.OrgSecretUpdate(owner, secret)
		return err
	}
	_, err = client.SecretUpdate(repoID, secret)
	return err
}

package secret

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var secretCreateCmd = &cli.Command{
	Name:      "add",
	Usage:     "adds a secret",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    secretCreate,
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

func secretCreate(c *cli.Context) error {
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
	if len(secret.Events) == 0 {
		secret.Events = defaultSecretEvents
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
		_, err = client.GlobalSecretCreate(secret)
		return err
	}

	if orgID != -1 {
		_, err = client.OrgSecretCreate(orgID, secret)
		return err
	}

	_, err = client.SecretCreate(repoID, secret)
	return err
}

var defaultSecretEvents = []string{
	woodpecker.EventPush,
	woodpecker.EventTag,
	woodpecker.EventDeploy,
}

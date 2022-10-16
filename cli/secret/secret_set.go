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
	ArgsUsage: "[org/repo|org]",
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
	),
}

func secretUpdate(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	secretVal := strings.ToLower(c.String("value"))
	secret := &woodpecker.Secret{
		Name:   c.String("name"),
		Value:  secretVal,
		Images: c.StringSlice("image"),
		Events: c.StringSlice("event"),
	}
	if strings.HasPrefix(secret.Value, "@") {
		path := strings.TrimPrefix(secret.Value, "@")
		out, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		secret.Value = string(out)
	}

	global, owner, repo, err := parseTargetArgs(c)
	if err != nil {
		return err
	}

	if global {
		_, err = client.GlobalSecretUpdate(secret)
		return err
	}
	if repo == "" {
		_, err = client.OrgSecretUpdate(owner, secret)
		return err
	}
	_, err = client.SecretUpdate(owner, repo, secret)
	return err
}

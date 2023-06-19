package secret

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var secretDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a secret",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    secretDelete,
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
	),
}

func secretDelete(c *cli.Context) error {
	secretName := c.String("name")

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	global, owner, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	if global {
		return client.GlobalSecretDelete(secretName)
	}
	if owner != "" {
		return client.OrgSecretDelete(owner, secretName)
	}
	return client.SecretDelete(repoID, secretName)
}

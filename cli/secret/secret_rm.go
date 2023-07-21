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
		common.OrgFlag,
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

	global, orgID, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	if global {
		return client.GlobalSecretDelete(secretName)
	}
	if orgID != -1 {
		return client.OrgSecretDelete(orgID, secretName)
	}
	return client.SecretDelete(repoID, secretName)
}

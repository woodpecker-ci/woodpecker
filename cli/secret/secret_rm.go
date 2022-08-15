package secret

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var secretDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a secret",
	ArgsUsage: "[org/repo|org]",
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
		&cli.StringFlag{
			Name:  "repository",
			Usage: "repository name (e.g. octocat/hello-world)",
		},
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

	global, owner, repo, err := parseTargetArgs(c)
	if err != nil {
		return err
	}

	if global {
		return client.GlobalSecretDelete(secretName)
	}
	if repo == "" {
		return client.OrgSecretDelete(owner, secretName)
	}
	return client.SecretDelete(owner, repo, secretName)
}

package secret

import (
	"strings"

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
			Usage: "organizations name (e.g. octocat)",
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
	var (
		secretName = c.String("name")
		orgName    = c.String("organization")
		repoName   = c.String("repository")
	)
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	if c.Bool("global") {
		return client.GlobalSecretDelete(secretName)
	}
	if orgName == "" && repoName == "" {
		repoName = c.Args().First()
	}
	if orgName == "" && !strings.Contains(repoName, "/") {
		orgName = repoName
	}
	if orgName != "" {
		return client.OrgSecretDelete(orgName, secretName)
	}
	owner, name, err := internal.ParseRepo(repoName)
	if err != nil {
		return err
	}
	return client.SecretDelete(owner, name, secretName)
}

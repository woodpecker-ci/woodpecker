package secret

import (
	"html/template"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var secretInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display secret info",
	ArgsUsage: "[org/repo|org]",
	Action:    secretInfo,
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
		common.FormatFlag(tmplSecretList, true),
	),
}

func secretInfo(c *cli.Context) error {
	var (
		secretName = c.String("name")
		orgName    = c.String("organization")
		repoName   = c.String("repository")
		format     = c.String("format") + "\n"
	)
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	var secret *woodpecker.Secret
	if c.Bool("global") {
		secret, err = client.GlobalSecret(secretName)
		if err != nil {
			return err
		}
	} else {
		if orgName == "" && repoName == "" {
			repoName = c.Args().First()
		}
		if orgName == "" && !strings.Contains(repoName, "/") {
			orgName = repoName
		}
		if orgName != "" {
			secret, err = client.OrgSecret(orgName, secretName)
			if err != nil {
				return err
			}
		} else {
			owner, name, err := internal.ParseRepo(repoName)
			if err != nil {
				return err
			}
			secret, err = client.Secret(owner, name, secretName)
			if err != nil {
				return err
			}
		}
	}

	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, secret)
}

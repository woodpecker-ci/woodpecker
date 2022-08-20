package secret

import (
	"html/template"
	"os"

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
		format     = c.String("format") + "\n"
	)
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	global, owner, repo, err := parseTargetArgs(c)
	if err != nil {
		return err
	}

	var secret *woodpecker.Secret
	if global {
		secret, err = client.GlobalSecret(secretName)
		if err != nil {
			return err
		}
	} else if repo == "" {
		secret, err = client.OrgSecret(owner, secretName)
		if err != nil {
			return err
		}
	} else {
		secret, err = client.Secret(owner, repo, secretName)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, secret)
}

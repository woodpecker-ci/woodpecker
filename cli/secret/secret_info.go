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
	ArgsUsage: "[repo-id|repo-full-name]",
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
		common.RepoFlag,
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

	global, owner, repoID, err := parseTargetArgs(client, c)
	if err != nil {
		return err
	}

	var secret *woodpecker.Secret
	if global {
		secret, err = client.GlobalSecret(secretName)
		if err != nil {
			return err
		}
	} else if owner != "" {
		secret, err = client.OrgSecret(owner, secretName)
		if err != nil {
			return err
		}
	} else {
		secret, err = client.Secret(repoID, secretName)
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

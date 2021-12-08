package secret

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var secretInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display secret info",
	ArgsUsage: "[repo/name]",
	Action:    secretInfo,
	Flags: append(common.GlobalFlags,
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
	common.SetupConsoleLogger(c)
	var (
		secretName = c.String("name")
		repoName   = c.String("repository")
		format     = c.String("format") + "\n"
	)
	if repoName == "" {
		repoName = c.Args().First()
	}
	owner, name, err := internal.ParseRepo(repoName)
	if err != nil {
		return err
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	secret, err := client.Secret(owner, name, secretName)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, secret)
}

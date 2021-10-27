package secret

import (
	"html/template"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var secretListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list secrets",
	ArgsUsage: "[repo/name]",
	Action:    secretList,
	Flags: append(common.GlobalFlags,
		&cli.StringFlag{
			Name:  "repository",
			Usage: "repository name (e.g. octocat/hello-world)",
		},
		common.FormatFlag(tmplSecretList, true),
	),
}

func secretList(c *cli.Context) error {
	var (
		format   = c.String("format") + "\n"
		reponame = c.String("repository")
	)
	if reponame == "" {
		reponame = c.Args().First()
	}
	owner, name, err := internal.ParseRepo(reponame)
	if err != nil {
		return err
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	list, err := client.SecretList(owner, name)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	for _, registry := range list {
		tmpl.Execute(os.Stdout, registry)
	}
	return nil
}

// template for secret list items
var tmplSecretList = "\x1b[33m{{ .Name }} \x1b[0m" + `
Events: {{ list .Events }}
{{- if .Images }}
Images: {{ list .Images }}
{{- else }}
Images: <any>
{{- end }}
`

var secretFuncMap = template.FuncMap{
	"list": func(s []string) string {
		return strings.Join(s, ", ")
	},
}

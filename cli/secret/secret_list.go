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

var secretListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list secrets",
	ArgsUsage: "[org/name|org]",
	Action:    secretList,
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
		common.FormatFlag(tmplSecretList, true),
	),
}

func secretList(c *cli.Context) error {
	format := c.String("format") + "\n"

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	global, owner, repo, err := parseTargetArgs(c)
	if err != nil {
		return err
	}

	var list []*woodpecker.Secret
	if global {
		list, err = client.GlobalSecretList()
		if err != nil {
			return err
		}
	} else if repo == "" {
		list, err = client.OrgSecretList(owner)
		if err != nil {
			return err
		}
	} else {
		list, err = client.SecretList(owner, repo)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.New("_").Funcs(secretFuncMap).Parse(format)
	if err != nil {
		return err
	}
	for _, registry := range list {
		if err := tmpl.Execute(os.Stdout, registry); err != nil {
			return err
		}
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

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
			Usage: "organizations name (e.g. octocat)",
		},
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
		orgName  = c.String("organization")
		repoName = c.String("repository")
	)
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	var list []*woodpecker.Secret
	if c.Bool("global") {
		list, err = client.GlobalSecretList()
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
			list, err = client.OrgSecretList(orgName)
			if err != nil {
				return err
			}
		} else {
			owner, name, err := internal.ParseRepo(repoName)
			if err != nil {
				return err
			}
			list, err = client.SecretList(owner, name)
			if err != nil {
				return err
			}
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

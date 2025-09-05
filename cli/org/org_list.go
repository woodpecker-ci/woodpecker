package org

import (
	"context"
	"os"
	"strings"
	"text/template"

	"github.com/urfave/cli/v3"
	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/internal"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

var orgListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list organizations",
	ArgsUsage: "",
	Action:    orgList,
	Flags: []cli.Flag{
		common.FormatFlag(tmplOrgList, true),
	},
}

func orgList(ctx context.Context, c *cli.Command) error {
	format := c.String("format") + "\n"

	client, err := internal.NewClient(ctx, c)
	if err != nil {
		return err
	}

	opt := woodpecker.ListOptions{}

	list, err := client.OrgList(opt)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Funcs(orgFuncMap).Parse(format)
	if err != nil {
		return err
	}

	for _, org := range list {
		if err := tmpl.Execute(os.Stdout, org); err != nil {
			return err
		}
	}
	return nil
}

// Template for org list items.
var tmplOrgList = "\x1b[33m{{ .Name }} \x1b[0m" + `
Organization ID: {{ .ID }}
`

var orgFuncMap = template.FuncMap{
	"list": func(s []string) string {
		return strings.Join(s, ", ")
	},
}

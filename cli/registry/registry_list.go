package registry

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var registryListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list registries",
	ArgsUsage: "[repo/name]",
	Action:    registryList,
	Flags: append(common.GlobalFlags,
		&cli.StringFlag{
			Name:  "repository",
			Usage: "repository name (e.g. octocat/hello-world)",
		},
		common.FormatFlag(tmplRegistryList, true),
	),
}

func registryList(c *cli.Context) error {
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
	list, err := client.RegistryList(owner, name)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	for _, registry := range list {
		tmpl.Execute(os.Stdout, registry)
	}
	return nil
}

// template for build list information
var tmplRegistryList = "\x1b[33m{{ .Address }} \x1b[0m" + `
Username: {{ .Username }}
Email: {{ .Email }}
`

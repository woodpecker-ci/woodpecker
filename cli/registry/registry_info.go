package registry

import (
	"html/template"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var registryInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "display registry info",
	ArgsUsage: "[repo/name]",
	Action:    registryInfo,
	Flags: append(common.GlobalFlags,
		&cli.StringFlag{
			Name:  "repository",
			Usage: "repository name (e.g. octocat/hello-world)",
		},
		&cli.StringFlag{
			Name:  "hostname",
			Usage: "registry hostname",
			Value: "docker.io",
		},
		common.FormatFlag(tmplRegistryList, true),
	),
}

func registryInfo(c *cli.Context) error {
	var (
		hostname = c.String("hostname")
		reponame = c.String("repository")
		format   = c.String("format") + "\n"
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
	registry, err := client.Registry(owner, name, hostname)
	if err != nil {
		return err
	}
	tmpl, err := template.New("_").Parse(format)
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, registry)
}

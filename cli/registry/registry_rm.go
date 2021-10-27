package registry

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var registryDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a registry",
	ArgsUsage: "[repo/name]",
	Action:    registryDelete,
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
	),
}

func registryDelete(c *cli.Context) error {
	var (
		hostname = c.String("hostname")
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
	return client.RegistryDelete(owner, name, hostname)
}

package registry

import (
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
)

var registryDeleteCmd = &cli.Command{
	Name:      "rm",
	Usage:     "remove a registry",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    registryDelete,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:  "hostname",
			Usage: "registry hostname",
			Value: "docker.io",
		},
	),
}

func registryDelete(c *cli.Context) error {
	var (
		hostname         = c.String("hostname")
		repoIDOrFullName = c.String("repository")
	)
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	repoID, err := internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return err
	}
	return client.RegistryDelete(repoID, hostname)
}

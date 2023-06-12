package registry

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var registryUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a registry",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    registryUpdate,
	Flags: append(common.GlobalFlags,
		common.RepoFlag,
		&cli.StringFlag{
			Name:  "hostname",
			Usage: "registry hostname",
			Value: "docker.io",
		},
		&cli.StringFlag{
			Name:  "username",
			Usage: "registry username",
		},
		&cli.StringFlag{
			Name:  "password",
			Usage: "registry password",
		},
	),
}

func registryUpdate(c *cli.Context) error {
	var (
		hostname         = c.String("hostname")
		username         = c.String("username")
		password         = c.String("password")
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
	registry := &woodpecker.Registry{
		Address:  hostname,
		Username: username,
		Password: password,
	}
	if strings.HasPrefix(registry.Password, "@") {
		path := strings.TrimPrefix(registry.Password, "@")
		out, ferr := os.ReadFile(path)
		if ferr != nil {
			return ferr
		}
		registry.Password = string(out)
	}
	_, err = client.RegistryUpdate(repoID, registry)
	return err
}

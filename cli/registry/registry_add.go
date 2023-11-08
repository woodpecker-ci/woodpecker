package registry

import (
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
)

var registryCreateCmd = &cli.Command{
	Name:      "add",
	Usage:     "adds a registry",
	ArgsUsage: "[repo-id|repo-full-name]",
	Action:    registryCreate,
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

func registryCreate(c *cli.Context) error {
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
	_, err = client.RegistryCreate(repoID, registry)
	if err != nil {
		return err
	}
	return nil
}

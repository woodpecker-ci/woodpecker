package registry

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the registry command set.
var Command = &cli.Command{
	Name:  "registry",
	Usage: "manage registries",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		registryCreateCmd,
		registryDeleteCmd,
		registryUpdateCmd,
		registryInfoCmd,
		registryListCmd,
	},
}

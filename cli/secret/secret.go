package secret

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the secret command.
var Command = &cli.Command{
	Name:  "secret",
	Usage: "manage secrets",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		secretCreateCmd,
		secretDeleteCmd,
		secretUpdateCmd,
		secretInfoCmd,
		secretListCmd,
	},
}

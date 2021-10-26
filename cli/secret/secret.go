package secret

import "github.com/urfave/cli/v2"

// Command exports the secret command.
var Command = &cli.Command{
	Name:  "secret",
	Usage: "manage secrets",
	Subcommands: []*cli.Command{
		secretCreateCmd,
		secretDeleteCmd,
		secretUpdateCmd,
		secretInfoCmd,
		secretListCmd,
	},
}

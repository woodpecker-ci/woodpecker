package server

import "github.com/urfave/cli"

// Command exports the user command set.
var Command = cli.Command{
	Name:  "server",
	Usage: "manage servers",
	Subcommands: []cli.Command{
		serverListCmd,
		serverInfoCmd,
		serverOpenCmd,
		serverCreateCmd,
		serverDestroyCmd,
		serverEnvCmd,
	},
}

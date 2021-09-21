package log

import "github.com/urfave/cli"

// Command exports the build command set.
var Command = cli.Command{
	Name:  "log",
	Usage: "manage logs",
	Subcommands: []cli.Command{
		logPurgeCmd,
	},
}

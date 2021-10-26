package build

import "github.com/urfave/cli/v2"

// Command exports the build command set.
var Command = cli.Command{
	Name:  "build",
	Usage: "manage builds",
	Subcommands: []cli.Command{
		buildListCmd,
		buildLastCmd,
		buildLogsCmd,
		buildInfoCmd,
		buildStopCmd,
		buildStartCmd,
		buildApproveCmd,
		buildDeclineCmd,
		buildQueueCmd,
		buildKillCmd,
		buildPsCmd,
	},
}

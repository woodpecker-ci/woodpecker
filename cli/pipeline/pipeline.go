package pipeline

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the build command set.
var Command = &cli.Command{
	Name:    "pipeline",
	Aliases: []string{"build"},
	Usage:   "manage pipelines",
	Flags:   common.GlobalFlags,
	Subcommands: []*cli.Command{
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

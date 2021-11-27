package log

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the build command set.
var Command = &cli.Command{
	Name:  "log",
	Usage: "manage logs",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		logPurgeCmd,
	},
}

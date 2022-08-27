package cron

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the cron command set.
var Command = &cli.Command{
	Name:  "cron",
	Usage: "manage cron jobs",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		cronCreateCmd,
		cronDeleteCmd,
		cronUpdateCmd,
		cronInfoCmd,
		cronListCmd,
	},
}

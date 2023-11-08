package user

import (
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
)

// Command exports the user command set.
var Command = &cli.Command{
	Name:  "user",
	Usage: "manage users",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		userListCmd,
		userInfoCmd,
		userAddCmd,
		userRemoveCmd,
	},
}

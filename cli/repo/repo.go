package repo

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the repository command.
var Command = &cli.Command{
	Name:  "repo",
	Usage: "manage repositories",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		repoListCmd,
		repoInfoCmd,
		repoAddCmd,
		repoUpdateCmd,
		repoRemoveCmd,
		repoRepairCmd,
		repoChownCmd,
		repoSyncCmd,
	},
}

package repo

import "github.com/urfave/cli/v2"

// Command exports the repository command.
var Command = cli.Command{
	Name:  "repo",
	Usage: "manage repositories",
	Subcommands: []cli.Command{
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

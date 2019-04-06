package autoscale

import "github.com/urfave/cli"

// Command exports the user command set.
var Command = cli.Command{
	Name:  "autoscale",
	Usage: "manage autoscaling",
	Subcommands: []cli.Command{
		autoscalePauseCmd,
		autoscaleResumeCmd,
		autoscaleVersionCmd,
	},
}

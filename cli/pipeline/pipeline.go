package pipeline

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
)

// Command exports the pipeline command set.
var Command = &cli.Command{
	Name:    "pipeline",
	Aliases: []string{"build"},
	Usage:   "manage pipelines",
	Flags:   common.GlobalFlags,
	Subcommands: []*cli.Command{
		pipelineListCmd,
		pipelineLastCmd,
		pipelineLogsCmd,
		pipelineInfoCmd,
		pipelineStopCmd,
		pipelineStartCmd,
		pipelineApproveCmd,
		pipelineDeclineCmd,
		pipelineQueueCmd,
		pipelineKillCmd,
		pipelinePsCmd,
		pipelineCreateCmd,
	},
}

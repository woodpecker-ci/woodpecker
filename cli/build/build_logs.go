package build

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var buildLogsCmd = &cli.Command{
	Name:      "logs",
	Usage:     "show build logs",
	ArgsUsage: "<repo/name> [build] [job]",
	Action:    buildLogs,
}

func buildLogs(c *cli.Context) error {
	// TODO: add logs command
	return fmt.Errorf("Command temporarily disabled. See https://github.com/woodpecker-ci/woodpecker/issues/383")
}

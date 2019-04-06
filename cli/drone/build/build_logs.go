package build

import (
	"fmt"

	"github.com/urfave/cli"
)

var buildLogsCmd = cli.Command{
	Name:      "logs",
	Usage:     "show build logs",
	ArgsUsage: "<repo/name> [build] [job]",
	Action:    buildLogs,
}

func buildLogs(c *cli.Context) error {
	return fmt.Errorf("Command temporarily disabled. See https://github.com/drone/drone/issues/2005")
}

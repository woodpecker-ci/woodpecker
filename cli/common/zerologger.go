package common

import (
	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cmd/common"
)

func SetupGlobalLogger(c *cli.Context) error {
	common.SetupGlobalLogger(c)
	return nil
}

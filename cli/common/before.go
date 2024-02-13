package common

import (
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal/config"
)

func Before(c *cli.Context) error {
	err := setupGlobalLogger(c)
	if err != nil {
		return err
	}

	return config.Load(c)
}

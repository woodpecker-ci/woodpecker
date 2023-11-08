package loglevel

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
)

// Command exports the log-level command used to change the servers log-level.
var Command = &cli.Command{
	Name:      "log-level",
	ArgsUsage: "[level]",
	Usage:     "get the logging level of the server, or set it with [level]",
	Action:    logLevel,
	Flags:     common.GlobalFlags,
}

func logLevel(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	var ll *woodpecker.LogLevel
	arg := c.Args().First()
	if arg != "" {
		lvl, err := zerolog.ParseLevel(arg)
		if err != nil {
			return err
		}
		ll, err = client.SetLogLevel(&woodpecker.LogLevel{
			Level: lvl.String(),
		})
		if err != nil {
			return err
		}
	} else {
		ll, err = client.LogLevel()
		if err != nil {
			return err
		}
	}

	log.Info().Msgf("Logging level: %s", ll.Level)
	return nil
}

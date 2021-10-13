package loglevel

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli"

	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

// Command exports the log-level command.
var Command = cli.Command{
	Name:   "log-level",
	Usage:  "get the logging level of the server, or set it with --level",
	Action: logLevel,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "level",
			Usage: "set the logging level",
		},
	},
}

func logLevel(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	var ll *woodpecker.LogLevel
	if c.IsSet("level") {
		lvl, err := zerolog.ParseLevel(c.String("level"))
		if err != nil {
			return err
		}
		ll, err = client.SetLogLevel(&woodpecker.LogLevel{
			Level: lvl.String(),
		})
	} else {
		ll, err = client.LogLevel()
	}

	log.Info().Msgf("Logging level: %s", ll.Level)
	return nil
}

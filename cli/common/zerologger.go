package common

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func SetupConsoleLogger(c *cli.Context) error {
	level := c.String("log-level")
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Fatal().Msgf("unknown logging level: %s", level)
	}
	zerolog.SetGlobalLevel(lvl)
	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
		log.Logger = log.With().Caller().Logger()
		log.Log().Msgf("LogLevel = %s", zerolog.GlobalLevel().String())
	}
	return nil
}

package common

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/update"
)

func Before(c *cli.Context) error {
	if err := SetupGlobalLogger(c); err != nil {
		return err
	}

	// TODO: background update check
	go func() {
		log.Debug().Msg("Checking for updates ...")

		newVersion, err := update.CheckForUpdate(c, false)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to check for updates")
			return
		}

		if newVersion != nil {
			log.Warn().Msgf("A new version of woodpecker-cli is available: %s", newVersion)
		} else {
			log.Debug().Msgf("You are using the latest version of woodpecker-cli")
		}
	}()

	return nil
}

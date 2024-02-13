package common

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/update"
)

var waitForUpdateCheck chan struct{}

func Before(c *cli.Context) error {
	if err := SetupGlobalLogger(c); err != nil {
		return err
	}

	go func() {
		waitForUpdateCheck = make(chan struct{})

		log.Debug().Msg("Checking for updates ...")

		newVersion, err := update.CheckForUpdate(context.Background(), true)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to check for updates")
			return
		}

		if newVersion != nil {
			log.Warn().Msgf("A new version of woodpecker-cli is available: %s. Update by running: %s update", newVersion.Version, c.App.Name)
		} else {
			log.Debug().Msgf("No update required")
		}

		close(waitForUpdateCheck)
	}()

	return nil
}

func After(_ *cli.Context) error {
	if waitForUpdateCheck != nil {
		<-waitForUpdateCheck
	}

	return nil
}

package common

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/v2/cli/update"
)

var (
	waitForUpdateCheck  context.Context
	cancelWaitForUpdate context.CancelCauseFunc
)

func Before(c *cli.Context) error {
	if err := SetupGlobalLogger(c); err != nil {
		return err
	}

	go func() {
		if c.Bool("disable-update-check") {
			return
		}

		// Don't check for updates when the update command is executed
		if firstArg := c.Args().First(); firstArg == "update" {
			return
		}

		waitForUpdateCheck, cancelWaitForUpdate = context.WithCancelCause(context.Background())
		defer cancelWaitForUpdate(errors.New("update check finished"))

		log.Debug().Msg("Checking for updates ...")

		newVersion, err := update.CheckForUpdate(waitForUpdateCheck, true)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to check for updates")
			return
		}

		if newVersion != nil {
			log.Warn().Msgf("A new version of woodpecker-cli is available: %s. Update by running: %s update", newVersion.Version, c.App.Name)
		} else {
			log.Debug().Msgf("No update required")
		}
	}()

	return nil
}

func After(_ *cli.Context) error {
	if waitForUpdateCheck != nil {
		select {
		case <-waitForUpdateCheck.Done():
		// When the actual command already finished, we still wait 250ms for the update check to finish
		case <-time.After(time.Millisecond * 250):
			log.Debug().Msg("Update check stopped due to timeout")
			cancelWaitForUpdate(errors.New("update check timeout"))
		}
	}

	return nil
}

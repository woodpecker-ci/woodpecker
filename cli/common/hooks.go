package common

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/internal/config"
	"go.woodpecker-ci.org/woodpecker/v2/cli/update"
)

var (
	waitForUpdateCheck  context.Context
	cancelWaitForUpdate context.CancelCauseFunc
)

func Before(ctx context.Context, c *cli.Command) (context.Context, error) {
	if err := setupGlobalLogger(ctx, c); err != nil {
		return ctx, err
	}

	go func(context.Context) {
		if c.Bool("disable-update-check") {
			return
		}

		// Don't check for updates when the update command is executed
		if firstArg := c.Args().First(); firstArg == "update" {
			return
		}

		waitForUpdateCheck, cancelWaitForUpdate = context.WithCancelCause(context.Background())
		defer cancelWaitForUpdate(errors.New("update check finished"))

		log.Debug().Msg("checking for updates ...")

		newVersion, err := update.CheckForUpdate(waitForUpdateCheck, false) //nolint:contextcheck
		if err != nil {
			log.Error().Err(err).Msgf("failed to check for updates")
			return
		}

		if newVersion != nil {
			log.Warn().Msgf("new version of woodpecker-cli is available: %s, update with: %s update", newVersion.Version, c.Root().Name)
		} else {
			log.Debug().Msgf("no update required")
		}
	}(ctx)

	return ctx, config.Load(ctx, c)
}

func After(_ context.Context, _ *cli.Command) error {
	if waitForUpdateCheck != nil {
		select {
		case <-waitForUpdateCheck.Done():
		// When the actual command already finished, we still wait 500ms for the update check to finish
		case <-time.After(time.Millisecond * 500):
			log.Debug().Msg("update check stopped due to timeout")
			cancelWaitForUpdate(errors.New("update check timeout"))
		}
	}

	return nil
}

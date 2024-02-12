package common

import (
	"github.com/urfave/cli/v2"
)

func Before(c *cli.Context) error {
	if err := SetupGlobalLogger(c); err != nil {
		return err
	}

	// TODO: background update check
	// go func() {
	// 	log.Debug().Msg("Checking for updates ...")

	// 	newVersion, err := update.CheckForUpdate(c, false)
	// 	if err != nil {
	// 		log.Printf("Failed to check for updates: %s", err)
	// 		return
	// 	}

	// 	if newVersion != nil {
	// 		log.Info().Msgf("A new version of woodpecker-cli is available: %s", newVersion)
	// 	} else {
	// 		log.Debug().Msgf("You are using the latest version of woodpecker-cli")
	// 	}
	// }()

	return nil
}

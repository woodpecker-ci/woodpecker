package update

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

// Command exports the update command.
var Command = &cli.Command{
	Name:  "update",
	Usage: "update the woodpecker-cli to the latest version",
	Flags: []cli.Flag{
		&cli.BoolFlag{
			Name:  "force",
			Usage: "force update even if the latest version is already installed",
		},
	},
	Action: update,
}

func update(c *cli.Context) error {
	log.Info().Msg("Checking for updates ...")

	newVersion, err := CheckForUpdate(c.Context, c.Bool("force"))
	if err != nil {
		return err
	}

	if newVersion == nil {
		fmt.Println("You are using the latest version of woodpecker-cli")
		return nil
	}

	log.Info().Msgf("New version %s is available! Updating ...", newVersion.Version)

	var tarFilePath string
	tarFilePath, err = downloadNewVersion(c.Context, newVersion.AssetURL)
	if err != nil {
		return err
	}

	log.Debug().Msgf("New version %s has been downloaded successfully! Installing ...", newVersion.Version)

	binFile, err := extractNewVersion(tarFilePath)
	if err != nil {
		return err
	}

	log.Debug().Msgf("New version %s has been extracted to %s", newVersion.Version, binFile)

	executablePathOrSymlink, err := os.Executable()
	if err != nil {
		return err
	}

	executablePath, err := filepath.EvalSymlinks(executablePathOrSymlink)
	if err != nil {
		return err
	}

	if err := os.Rename(binFile, executablePath); err != nil {
		return err
	}

	log.Info().Msgf("woodpecker-cli has been updated to version %s successfully!", newVersion.Version)

	return nil
}

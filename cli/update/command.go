package update

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
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

func update(ctx context.Context, c *cli.Command) error {
	log.Info().Msg("checking for updates ...")

	newVersion, err := CheckForUpdate(ctx, c.Bool("force"))
	if err != nil {
		return err
	}

	if newVersion == nil {
		fmt.Println("you are using the latest version of woodpecker-cli")
		return nil
	}

	log.Info().Msgf("new version %s is available! Updating ...", newVersion.Version)

	var tarFilePath string
	tarFilePath, err = downloadNewVersion(ctx, newVersion.AssetURL)
	if err != nil {
		return err
	}

	log.Debug().Msgf("new version %s has been downloaded successfully! Installing ...", newVersion.Version)

	binFile, err := extractNewVersion(tarFilePath)
	if err != nil {
		return err
	}

	log.Debug().Msgf("new version %s has been extracted to %s", newVersion.Version, binFile)

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

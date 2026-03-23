// Copyright 2024 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

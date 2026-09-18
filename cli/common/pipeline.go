// Copyright 2023 Woodpecker Authors
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

package common

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
)

func DetectPipelineConfig() (isDir bool, config string, _ error) {
	isDir, config, found, err := FindPipelineConfig(".")
	if err != nil {
		return false, "", err
	}
	if found {
		return isDir, config, nil
	}

	return false, "", fmt.Errorf("could not detect pipeline config")
}

// FindPipelineConfig searches dir using the default pipeline config order.
func FindPipelineConfig(dir string) (isDir bool, config string, found bool, _ error) {
	for _, configPath := range constant.DefaultConfigOrder {
		shouldBeDir := strings.HasSuffix(configPath, "/")
		configPath = filepath.Join(dir, strings.TrimSuffix(configPath, "/"))

		fi, err := os.Stat(configPath)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, "", false, err
		}

		if shouldBeDir == fi.IsDir() {
			return fi.IsDir(), configPath, true, nil
		}
	}

	return false, "", false, nil
}

func RunPipelineFunc(ctx context.Context, c *cli.Command, fileFunc, dirFunc func(context.Context, *cli.Command, string) error) error {
	if c.Args().Len() == 0 {
		isDir, path, err := DetectPipelineConfig()
		if err != nil {
			return err
		}
		if isDir {
			return dirFunc(ctx, c, path)
		}
		return fileFunc(ctx, c, path)
	}

	multiArgs := c.Args().Len() > 1
	for _, arg := range c.Args().Slice() {
		fi, err := os.Stat(arg)
		if err != nil {
			return err
		}
		if multiArgs {
			fmt.Println("#", fi.Name())
		}
		if fi.IsDir() {
			if err := dirFunc(ctx, c, arg); err != nil {
				return err
			}
		} else {
			if err := fileFunc(ctx, c, arg); err != nil {
				return err
			}
		}
		if multiArgs {
			fmt.Println("")
		}
	}

	return nil
}

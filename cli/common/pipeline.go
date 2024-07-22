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
	"strings"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

func DetectPipelineConfig() (isDir bool, config string, _ error) {
	for _, config := range constant.DefaultConfigOrder {
		shouldBeDir := strings.HasSuffix(config, "/")
		config = strings.TrimSuffix(config, "/")

		if fi, err := os.Stat(config); err == nil && shouldBeDir == fi.IsDir() {
			return fi.IsDir(), config, nil
		}
	}

	return false, "", fmt.Errorf("could not detect pipeline config")
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

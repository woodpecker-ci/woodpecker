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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

// TODO: use don't import from server => move FileMeta to pipeline package
func GetConfigs(c *cli.Context, dir string) ([]*forge_types.FileMeta, error) {
	stat, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}

	if stat.Mode().IsRegular() {
		data, err := os.ReadFile(dir)
		if err != nil {
			return nil, err
		}

		return []*forge_types.FileMeta{{
			Name: dir,
			Data: data,
		}}, nil
	}

	var files []*forge_types.FileMeta
	err = filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		if !strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		files = append(files, &forge_types.FileMeta{
			Name: path,
			Data: data,
		})

		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

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

func RunPipelineFunc(c *cli.Context, fileFunc, dirFunc func(*cli.Context, string) error) error {
	if c.Args().Len() == 0 {
		isDir, path, err := DetectPipelineConfig()
		if err != nil {
			return err
		}
		if isDir {
			return dirFunc(c, path)
		}
		return fileFunc(c, path)
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
			if err := dirFunc(c, arg); err != nil {
				return err
			}
		} else {
			if err := fileFunc(c, arg); err != nil {
				return err
			}
		}
		if multiArgs {
			fmt.Println("")
		}
	}

	return nil
}

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

package lint

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/muesli/termenv"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v2/cli/common"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter"
)

// Command exports the info command.
var Command = &cli.Command{
	Name:      "lint",
	Usage:     "lint a pipeline configuration file",
	ArgsUsage: "[path/to/.woodpecker.yaml]",
	Action:    lint,
}

func lint(c *cli.Context) error {
	return common.RunPipelineFunc(c, lintFile, lintDir)
}

func lintDir(c *cli.Context, dir string) error {
	var errorStrings []string
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			fmt.Println("#", info.Name())
			if err := lintFile(c, path); err != nil {
				errorStrings = append(errorStrings, err.Error())
			}
			fmt.Println("")
			return nil
		}

		return nil
	}); err != nil {
		return err
	}

	if len(errorStrings) != 0 {
		return fmt.Errorf("ERRORS: %s", strings.Join(errorStrings, "; "))
	}
	return nil
}

func lintFile(_ *cli.Context, file string) error {
	output := termenv.NewOutput(os.Stdout)

	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()

	buf, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	rawConfig := string(buf)

	c, err := yaml.ParseString(rawConfig)
	if err != nil {
		return err
	}

	config := &linter.WorkflowConfig{
		File:      path.Base(file),
		RawConfig: rawConfig,
		Workflow:  c,
	}

	// TODO: lint multiple files at once to allow checks for sth like "depends_on" to work
	err = linter.New(linter.WithTrusted(true)).Lint([]*linter.WorkflowConfig{config})
	if err != nil {
		fmt.Printf("🔥 %s has errors:\n", output.String(config.File).Underline())

		hasErrors := true
		for _, err := range pipeline_errors.GetPipelineErrors(err) {
			line := "  "

			if err.IsWarning {
				line = fmt.Sprintf("%s ⚠️ ", line)
			} else {
				line = fmt.Sprintf("%s ❌", line)
				hasErrors = true
			}

			if data := err.GetLinterData(); data != nil {
				line = fmt.Sprintf("%s %s\t%s", line, output.String(data.Field).Bold(), err.Message)
			} else {
				line = fmt.Sprintf("%s %s", line, err.Message)
			}

			// TODO: use table output
			fmt.Printf("%s\n", line)
		}

		if hasErrors {
			return errors.New("config has errors")
		}

		return nil
	}

	fmt.Println("✅ Config is valid")
	return nil
}

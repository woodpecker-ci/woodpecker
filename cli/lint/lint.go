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
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/urfave/cli/v3"
	"go.uber.org/multierr"

	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/linter"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
)

// Command exports the info command.
var Command = &cli.Command{
	Name:      "lint",
	Usage:     "lint a pipeline configuration file",
	ArgsUsage: "[path/to/.woodpecker.yaml]",
	Action:    lint,
	Flags: []cli.Flag{
		&cli.StringSliceFlag{
			Sources: cli.EnvVars("WOODPECKER_PLUGINS_PRIVILEGED"),
			Name:    "plugins-privileged",
			Usage:   "allow plugins to run in privileged mode, if set empty, there is no",
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.StringSliceFlag{
			Sources: cli.EnvVars("WOODPECKER_PLUGINS_TRUSTED_CLONE"),
			Name:    "plugins-trusted-clone",
			Usage:   "plugins that are trusted to handle Git credentials in cloning steps",
			Value:   constant.TrustedClonePlugins,
			Config: cli.StringConfig{
				TrimSpace: true,
			},
		},
		&cli.BoolFlag{
			Sources: cli.EnvVars("WOODPECKER_LINT_STRICT"),
			Name:    "strict",
			Usage:   "treat warnings as errors",
		},
	},
}

func lint(ctx context.Context, c *cli.Command) error {
	return common.RunPipelineFunc(ctx, c, lintFile, lintDir)
}

func lintDir(_ context.Context, c *cli.Command, dir string) error {
	var files []string
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && (strings.HasSuffix(info.Name(), ".yaml") || strings.HasSuffix(info.Name(), ".yml")) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		return err
	}

	// lint all workflows at once, so their depends_on can be checked
	issues := make(map[string]error, len(files))
	fileOfWorkflow := make(map[string]string, len(files))
	var configs []*linter.WorkflowConfig
	for _, file := range files {
		config, err := parseFile(file, builder.SanitizePath(file))
		if err != nil {
			issues[file] = err
			continue
		}
		fileOfWorkflow[config.File] = file
		configs = append(configs, config)
	}
	for _, err := range pipeline_errors.GetPipelineErrors(newLinter(c, linter.WithWorkflowDependencies()).Lint(configs)) {
		file := fileOfWorkflow[pipeline_errors.GetFile(err)]
		issues[file] = multierr.Append(issues[file], err)
	}

	var errorStrings []string
	for _, file := range files {
		if err := report(c, path.Base(file), issues[file]); err != nil {
			errorStrings = append(errorStrings, err.Error())
		}
	}

	if len(errorStrings) != 0 {
		return fmt.Errorf("ERRORS: %s", strings.Join(errorStrings, "; "))
	}
	return nil
}

func lintFile(_ context.Context, c *cli.Command, file string) error {
	config, err := parseFile(file, path.Base(file))
	if err == nil {
		err = newLinter(c).Lint([]*linter.WorkflowConfig{config})
	}
	return report(c, path.Base(file), err)
}

// parseFile reads a workflow file and checks its syntax.
func parseFile(file, name string) (*linter.WorkflowConfig, error) {
	buf, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}

	rawConfig := string(buf)

	// only check the substitution syntax, as the values are unknown here
	if _, err := metadata.EnvVarSubst(rawConfig, nil); err != nil {
		return nil, err
	}

	parsedConfig, err := yaml.ParseString(rawConfig)
	if err != nil {
		return nil, err
	}

	return &linter.WorkflowConfig{
		File:      name,
		RawConfig: rawConfig,
		Workflow:  parsedConfig,
	}, nil
}

func newLinter(c *cli.Command, opts ...linter.Option) *linter.Linter {
	return linter.New(append([]linter.Option{
		linter.WithTrusted(linter.TrustedConfiguration{
			Network:  true,
			Volumes:  true,
			Security: true,
		}),
		linter.PrivilegedPlugins(c.StringSlice("plugins-privileged")),
		linter.WithTrustedClonePlugins(c.StringSlice("plugins-trusted-clone")),
	}, opts...)...)
}

// report prints the issues of a file and returns an error if it has errors.
func report(c *cli.Command, file string, err error) error {
	if err == nil {
		_, _ = lipgloss.Fprintln(c.Writer, "✅ "+underline.Render(file)+" is valid")
		return nil
	}

	str, err := FormatLintError(file, err, c.Bool("strict"))
	_, _ = lipgloss.Fprint(c.Writer, str)
	return err
}

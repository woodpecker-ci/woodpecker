package lint

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/multierr"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter"
)

// Command exports the info command.
var Command = &cli.Command{
	Name:      "lint",
	Usage:     "lint a pipeline configuration file",
	ArgsUsage: "[path/to/.woodpecker.yml]",
	Action:    lint,
	Flags:     common.GlobalFlags,
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
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".yml") {
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
	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()

	buf, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	rawConfig := string(buf)

	c, err := yaml.ParseString(rawConfig)
	if err != nil {
		return err
	}

	lerr := linter.New(linter.WithTrusted(true)).Lint(string(buf), c)
	if lerr != nil {
		fmt.Println("❌ Config is invalid")

		linterErrors := multierr.Errors(lerr)
		for _, err := range linterErrors {
			var linterError *linter.LinterError
			if errors.As(err, &linterError) {
				fmt.Printf("\tIn %s: %s\n", linterError.Field, linterError.Message)
			} else {
				return err
			}
		}

		return nil
	}

	fmt.Println("✅ Config is valid")
	return nil
}

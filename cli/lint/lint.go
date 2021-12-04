package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/pipeline/schema"
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
	if c.Args().Len() == 0 {
		isDir, path, err := common.DetectPipelineConfig()
		if err != nil {
			return err
		}
		if isDir {
			return lintDir(path)
		}
		return lintFile(path)
	}

	for _, arg := range c.Args().Slice() {
		fi, err := os.Stat(arg)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			if err := lintDir(arg); err != nil {
				return err
			}
		} else {
			if err := lintFile(arg); err != nil {
				return err
			}
		}
	}

	return nil
}

func lintDir(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".yml") {
			fmt.Println("#", info.Name())
			_ = lintFile(path) // TODO: should we drop errors or store them and report back?
			fmt.Println("")
			return nil
		}

		return nil
	})
}

func lintFile(file string) error {
	configErrors, err := schema.Lint(file)
	if err != nil {
		fmt.Println("❌ Config is invalid")
		for _, configError := range configErrors {
			fmt.Println("In", configError.Field()+":", configError.Description())
		}
		return err
	}

	fmt.Println("✅ Config is valid")
	return nil
}

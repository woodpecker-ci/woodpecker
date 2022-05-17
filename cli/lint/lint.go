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
	return common.RunPipelineFunc(c, lintFile, lintDir)
}

func lintDir(c *cli.Context, dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".yml") {
			fmt.Println("#", info.Name())
			_ = lintFile(c, path) // TODO: should we drop errors or store them and report back?
			fmt.Println("")
			return nil
		}

		return nil
	})
}

func lintFile(_ *cli.Context, file string) error {
	fi, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fi.Close()

	configErrors, err := schema.Lint(fi)
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

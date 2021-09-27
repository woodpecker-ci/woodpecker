package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"
	"github.com/woodpecker-ci/woodpecker/pipeline/schema"
)

// Command exports the info command.
var Command = cli.Command{
	Name:      "lint",
	Usage:     "lint a pipeline configuration file",
	ArgsUsage: "[path/to/.woodpecker.yml]",
	Action:    lint,
	Flags:     []cli.Flag{},
}

func lint(c *cli.Context) error {
	file := c.Args().First()
	if file == "" {
		file = ".woodpecker.yml"
	}

	fi, err := os.Stat(file)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return lintFile(file)
	}

	return filepath.Walk(file, func(path string, info os.FileInfo, e error) error {
		if e != nil {
			return e
		}

		// check if it is a regular file (not dir)
		if info.Mode().IsRegular() && strings.HasSuffix(info.Name(), ".yml") {
			fmt.Println("#", info.Name())
			lintFile(path)
			fmt.Println("")
			return nil
		}

		return nil
	})
}

func lintFile(file string) error {
	err, configErrors := schema.Lint(file)
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

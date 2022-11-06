package common

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/shared/constant"
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

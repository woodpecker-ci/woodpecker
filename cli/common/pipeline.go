package common

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func DetectPipelineConfig() (multiplies bool, config string, _ error) {
	config = ".woodpecker"
	if fi, err := os.Stat(config); err == nil && fi.IsDir() {
		return true, config, nil
	}

	config = ".woodpecker.yml"
	if fi, err := os.Stat(config); err == nil && !fi.IsDir() {
		return true, config, nil
	}

	config = ".woodpecker.yaml"
	if fi, err := os.Stat(config); err == nil && !fi.IsDir() {
		return true, config, nil
	}

	config = ".drone.yml"
	fi, err := os.Stat(config)
	if err == nil && !fi.IsDir() {
		return false, config, nil
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

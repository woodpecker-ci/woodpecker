package common

import (
	"fmt"
	"os"
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

	config = ".drone.yml"
	fi, err := os.Stat(config)
	if err == nil && !fi.IsDir() {
		return false, config, nil
	}
	return false, "", fmt.Errorf("could not detect pipeline config")
}

package docker

import (
	"github.com/mitchellh/mapstructure"

	backend "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

// BackendOptions defines all the advanced options for the docker backend.
type BackendOptions struct {
	User string `mapstructure:"user"`
}

func parseBackendOptions(step *backend.Step) (BackendOptions, error) {
	var result BackendOptions
	if step == nil || step.BackendOptions == nil {
		return result, nil
	}
	err := mapstructure.Decode(step.BackendOptions[EngineName], &result)
	return result, err
}

package docker

import (
	"github.com/go-viper/mapstructure/v2"

	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
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
	err := mapstructure.WeakDecode(step.BackendOptions[EngineName], &result)
	return result, err
}

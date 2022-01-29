package backend

import (
	"fmt"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

var engines map[string]types.Engine

func init() {
	engines = make(map[string]types.Engine)

	// TODO: disabled for now as kubernetes backend has not been implemented yet
	// kubernetes
	// engine = kubernetes.New("", "", "")
	// engines[engine.Name()] = engine

	// docker
	engine := docker.New()
	engines[engine.Name()] = engine
}

func FindEngine(engineName string) (types.Engine, error) {
	if engineName == "auto-detect" {
		for _, engine := range engines {
			if engine.IsAvailable() {
				return engine, nil
			}
		}

		return nil, fmt.Errorf("Can't detect an available backend engine")
	}

	engine, ok := engines[engineName]
	if !ok {
		return nil, fmt.Errorf("Backend engine '%s' not found", engineName)
	}

	return engine, nil
}

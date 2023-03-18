package backend

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/kubernetes"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/local"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/ssh"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

var (
	enginesByName map[string]types.Engine
	engines       []types.Engine
)

func Init(ctx context.Context) {
	engines = []types.Engine{
		docker.New(),
		local.New(),
		ssh.New(),
		kubernetes.New(ctx),
	}

	enginesByName = make(map[string]types.Engine)
	for _, engine := range engines {
		enginesByName[engine.Name()] = engine
	}
}

func FindEngine(engineName string, ctx context.Context) (types.Engine, error) {
	if engineName == "auto-detect" {
		for _, engine := range engines {
			if engine.IsAvailable(ctx) {
				return engine, nil
			}
		}

		return nil, fmt.Errorf("can't detect an available backend engine")
	}

	engine, ok := enginesByName[engineName]
	if !ok {
		return nil, fmt.Errorf("backend engine '%s' not found", engineName)
	}

	return engine, nil
}

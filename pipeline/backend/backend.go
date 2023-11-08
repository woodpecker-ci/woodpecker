package backend

import (
	"context"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/pipeline/backend/docker"
	"go.woodpecker-ci.org/woodpecker/pipeline/backend/kubernetes"
	"go.woodpecker-ci.org/woodpecker/pipeline/backend/local"
	"go.woodpecker-ci.org/woodpecker/pipeline/backend/ssh"
	"go.woodpecker-ci.org/woodpecker/pipeline/backend/types"
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

func FindEngine(ctx context.Context, engineName string) (types.Engine, error) {
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

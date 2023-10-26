// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package backend

import (
	"context"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/kubernetes"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/local"
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

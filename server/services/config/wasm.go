// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/utils/wasm"
)

// wasmArgs are the arguments a module is started with.
// The second one tells a module that serves more than one kind of extension what is asked for.
var wasmArgs = []string{"woodpecker-extension", "config"}

type wasmService struct {
	runner *wasm.Runner
}

// NewWasm returns the in-process counterpart of NewHTTP.
//
// The request is written to stdin of the module and the response is read from its stdout.
// An empty output is what "204 No Content" is for the HTTP extension: keep the configuration as it is.
//
// The module never gets netrc data, as it has no network access it could use it for.
// And different to the HTTP extension a failing module fails the pipeline,
// instead of continuing with a configuration the module did not see through.
func NewWasm(runner *wasm.Runner) Service {
	return &wasmService{runner}
}

func (w *wasmService) Fetch(ctx context.Context, _ forge.Forge, _ *model.User, repo *model.Repo, pipeline *model.Pipeline, oldConfigData []*types.FileMeta, _ bool) ([]*types.FileMeta, error) {
	request, err := json.Marshal(requestStructure{
		Repo:          repo,
		Pipeline:      pipeline,
		Configuration: toConfigData(oldConfigData),
	})
	if err != nil {
		return nil, fmt.Errorf("could not encode request for wasm config extension: %w", err)
	}

	// one repo is one tenant, it can not keep the others from getting their config
	output, err := w.runner.Run(ctx, strconv.FormatInt(repo.ID, 10), wasmArgs, request)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch config via wasm: %w", err)
	}

	if len(bytes.TrimSpace(output)) == 0 {
		log.Debug().Str("repo", repo.FullName).Msg("wasm config extension returned nothing, using fallback config")
		return oldConfigData, nil
	}

	response := new(responseStructure)
	if err := json.Unmarshal(output, response); err != nil {
		return nil, fmt.Errorf("wasm config extension returned an invalid response: %w", err)
	}
	if response.Configs == nil {
		return nil, fmt.Errorf("wasm config extension returned a response without configs")
	}

	return toFileMeta(response.Configs)
}

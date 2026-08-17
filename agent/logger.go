// Copyright 2022 Woodpecker Authors
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

package agent

import (
	"io"
	"maps"
	"slices"
	"sync"

	"github.com/rs/zerolog"

	"go.woodpecker-ci.org/woodpecker/v3/agent/log"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/compile"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

func (r *Runner) createLogger(_logger zerolog.Logger, workflow *rpc.Workflow, results *compileResults) logging.Logger {
	return func(step *backend_types.Step, rc io.ReadCloser) error {
		defer rc.Close()

		logger := _logger.With().
			Str("image", step.Image).
			Logger()

		var secrets []string
		for _, secret := range workflow.Config.Secrets {
			secrets = append(secrets, secret.Value)
		}

		logger.Debug().Msg("log stream opened")

		lineWriter := log.NewLineWriter(r.client, step.UUID, secrets...)

		var (
			dst     io.Writer = lineWriter
			emitted func() ([]compile.Config, error)
		)

		if workflow.Phase == rpc.WorkflowPhaseCompile && step.Type != backend_types.StepTypeClone {
			// The scanner sits upstream of the masking line writer on purpose.
			// The masker substring-replaces every secret in every line, and
			// base64's alphabet makes a chance collision with a short secret
			// likely enough in a large payload that a masked copy cannot be
			// trusted as the authoritative one. What the writers receive is
			// still masked; only the copy the scanner decodes is not.
			dst, emitted = compile.ScanWriter(lineWriter, lineWriter.WithType(rpc.LogEntryCompileConfig))
		}

		if err := pipeline_utils.CopyLineByLine(dst, rc, pipeline.MaxLogLineLength); err != nil {
			logger.Error().Err(err).Msg("copy limited logStream part")
		}

		if emitted != nil {
			configs, err := emitted()
			results.set(step.UUID, configs, err)
		}

		logger.Debug().Msg("log stream copied, close ...")
		return nil
	}
}

// compileResults collects what each compile step emitted, keyed by step UUID.
// The logger runs one goroutine per step, so access is synchronized.
type compileResults struct {
	mu      sync.Mutex
	configs map[string][]compile.Config
	errs    map[string]error
}

func newCompileResults() *compileResults {
	return &compileResults{
		configs: map[string][]compile.Config{},
		errs:    map[string]error{},
	}
}

func (r *compileResults) set(stepUUID string, configs []compile.Config, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if err != nil {
		r.errs[stepUUID] = err
		return
	}
	r.configs[stepUUID] = configs
}

// report folds the per-step results into the single result the workflow sends
// with Done.
//
// It returns nil when no step was scanned, which is how the server tells an
// ordinary workflow from a compile one. Steps are folded in UUID order so the
// result does not depend on which step happened to finish first. A workflow
// whose steps disagree, one emitting configs and another failing, reports the
// failure: acting on half a response is worse than not acting at all.
func (r *compileResults) report() *rpc.CompileResult {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.configs) == 0 && len(r.errs) == 0 {
		return nil
	}

	if failed := slices.Sorted(maps.Keys(r.errs)); len(failed) > 0 {
		return &rpc.CompileResult{Error: r.errs[failed[0]].Error()}
	}

	result := new(rpc.CompileResult)
	for _, stepUUID := range slices.Sorted(maps.Keys(r.configs)) {
		for _, config := range r.configs[stepUUID] {
			result.Configs = append(result.Configs, rpc.CompileConfig{Name: config.Name, Data: config.Data})
		}
	}

	return result
}

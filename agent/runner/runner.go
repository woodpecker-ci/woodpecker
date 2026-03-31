// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package runner

import (
	"context"
	"fmt"
	"sync"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_runtime "go.woodpecker-ci.org/woodpecker/v3/pipeline/runtime"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// Runner fetches workflows from the server, executes them using a backend,
// tracks their state and reports results back.
type Runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
	counter  *State
	backend  backend_types.Backend
}

// NewRunner creates a Runner wired to the given RPC peer and backend.
func NewRunner(workEngine rpc.Peer, f rpc.Filter, h string, state *State, backend backend_types.Backend) Runner {
	return Runner{
		client:   workEngine,
		filter:   f,
		hostname: h,
		counter:  state,
		backend:  backend,
	}
}

// Run executes a single workflow lifecycle:
// fetch → build context → init → execute pipeline → finalize.
func (r *Runner) Run(runnerCtx, shutdownCtx context.Context) error {
	log.Debug().Msg("request next execution")

	// Preserve metadata AND cancellation from runnerCtx.
	meta, _ := metadata.FromOutgoingContext(runnerCtx)
	ctxMeta := metadata.NewOutgoingContext(shutdownCtx, meta)

	// Fetch next workflow from the queue.
	workflow, err := r.client.Next(runnerCtx, r.filter)
	if err != nil {
		return err
	}
	if workflow == nil {
		return nil
	}

	timeout := computeTimeout(workflow)
	repoName := extractRepositoryName(workflow.Config)
	pipelineNumber := extractPipelineNumber(workflow.Config)

	// Track workflow execution in runner state.
	r.counter.Add(workflow.ID, timeout, repoName, pipelineNumber)
	defer r.counter.Done(workflow.ID)

	logger := log.With().
		Str("repo", repoName).
		Str("pipeline", pipelineNumber).
		Str("workflow_id", workflow.ID).
		Logger()

	logger.Debug().Msg("received execution")

	// Build the workflow-scoped context with timeout, sigterm and cancellation support.
	workflowCtx, cancelWorkflow := buildWorkflowContext(ctxMeta, timeout, logger)
	defer cancelWorkflow(nil)

	// Start background goroutines for remote cancel and lease extension.
	startCancelListener(workflowCtx, cancelWorkflow, r.client, workflow.ID, logger)
	startLeaseExtender(workflowCtx, r.client, workflow.ID, logger)

	// Signal workflow initialization to server.
	if err := initWorkflow(runnerCtx, r.client, workflow.ID); err != nil {
		logger.Error().Err(err).Msg("signaling workflow initialization to server failed")
		cancelWorkflow(err)
		return err
	}

	var uploads sync.WaitGroup

	// Run pipeline.
	pipelineErr := pipeline_runtime.New(
		workflow.Config,
		r.backend,
		pipeline_runtime.WithContext(workflowCtx),
		pipeline_runtime.WithTaskUUID(fmt.Sprint(workflow.ID)),
		pipeline_runtime.WithLogger(r.createLogger(logger, &uploads, workflow)),
		pipeline_runtime.WithTracer(r.createTracer(ctxMeta, &uploads, logger, workflow)),
		pipeline_runtime.WithDescription(map[string]string{
			"workflow_id":     workflow.ID,
			"repo":            repoName,
			"pipeline_number": pipelineNumber,
		}),
	).Run(runnerCtx)

	// Wait for all async uploads and report final state.
	return finalizeWorkflow(runnerCtx, shutdownCtx, r.client, workflow.ID, pipelineErr, &uploads, logger)
}

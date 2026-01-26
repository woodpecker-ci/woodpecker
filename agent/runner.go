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

package agent

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

type Runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
	counter  *State
	backend  *backend.Backend
}

func NewRunner(workEngine rpc.Peer, f rpc.Filter, h string, state *State, backend *backend.Backend) Runner {
	return Runner{
		client:   workEngine,
		filter:   f,
		hostname: h,
		counter:  state,
		backend:  backend,
	}
}

// Run executes an workflow via an backend, tracks its state and reports back the state to the server
func (r *Runner) Run(runnerCtx, shutdownCtx context.Context) error { //nolint:contextcheck
	log.Debug().Msg("request next execution")

	// Preserve metadata AND cancellation from runnerCtx.
	meta, _ := metadata.FromOutgoingContext(runnerCtx)
	ctxMeta := metadata.NewOutgoingContext(runnerCtx, meta)

	// Fetch next workflow from the queue
	workflow, err := r.client.Next(runnerCtx, r.filter)
	if err != nil {
		return err
	}
	if workflow == nil {
		return nil
	}

	// Compute workflow timeout
	timeout := time.Hour
	if minutes := workflow.Timeout; minutes != 0 {
		timeout = time.Duration(minutes) * time.Minute
	}

	repoName := extractRepositoryName(workflow.Config)       // hack
	pipelineNumber := extractPipelineNumber(workflow.Config) // hack

	// Track workflow execution in runner state
	r.counter.Add(workflow.ID, timeout, repoName, pipelineNumber)
	defer r.counter.Done(workflow.ID)

	logger := log.With().
		Str("repo", repoName).
		Str("pipeline", pipelineNumber).
		Str("workflow_id", workflow.ID).
		Logger()

	logger.Debug().Msg("received execution")

	// Workflow execution context.
	// This context is the SINGLE source of truth for cancellation.
	workflowCtx, cancelWorkflowCtx := context.WithTimeout(ctxMeta, timeout)
	defer cancelWorkflowCtx()

	// Add sigterm support for internal context.
	// Required when the pipeline is terminated by external signals
	// like kubernetes.
	workflowCtx = utils.WithContextSigtermCallback(workflowCtx, func() {
		logger.Error().Msg("received sigterm termination signal")
		cancelWorkflowCtx()
	})

	// canceled indicates whether the workflow was canceled remotely (UI/API).
	// Must be atomic because it is written from a goroutine and read later.
	var canceled atomic.Bool

	// Listen for remote cancel events (UI / API).
	// When canceled, we MUST cancel the workflow context
	// so that pipeline execution and backend processes stop immediately.
	go func() {
		logger.Debug().Msg("listening for cancel signal")

		if err := r.client.Wait(workflowCtx, workflow.ID); err != nil {
			if errors.Is(err, pipeline.ErrCancel) {
				canceled.Store(true)
				logger.Debug().Err(err).Msg("cancel signal received")
				cancelWorkflowCtx()
			} else {
				logger.Error().Err(err).Msg("server returned unexpected err while waiting for workflow to finish run")
				cancelWorkflowCtx()
			}
		} else {
			// Wait returned without error, meaning the workflow finished normally
			logger.Debug().Msg("cancel listener exited normally")
		}
	}()

	// Periodically extend the workflow lease while running
	go func() {
		for {
			select {
			case <-workflowCtx.Done():
				logger.Debug().Msg("workflow context done")
				return

			case <-time.After(constant.TaskTimeout / 3):
				logger.Debug().Msg("renewing workflow lease")
				if err := r.client.Extend(workflowCtx, workflow.ID); err != nil {
					logger.Error().Err(err).Msg("failed to extend workflow lease")
				}
			}
		}
	}()

	state := rpc.WorkflowState{
		Started: time.Now().Unix(),
	}

	// Initialize workflow on the server
	if err := r.client.Init(runnerCtx, workflow.ID, state); err != nil {
		logger.Error().Err(err).Msg("signaling workflow initialization to server failed")
		// This should never happen, still it did so lets clean up and end
		// tdood:
	}

	var uploads sync.WaitGroup

	// Run pipeline
	err = pipeline.New(
		workflow.Config,
		pipeline.WithContext(workflowCtx),
		pipeline.WithTaskUUID(fmt.Sprint(workflow.ID)),
		pipeline.WithLogger(r.createLogger(logger, &uploads, workflow)),
		pipeline.WithTracer(r.createTracer(ctxMeta, &uploads, logger, workflow)),
		pipeline.WithBackend(*r.backend),
		pipeline.WithDescription(map[string]string{
			"workflow_id":     workflow.ID,
			"repo":            repoName,
			"pipeline_number": pipelineNumber,
		}),
	).Run(runnerCtx)

	state.Finished = time.Now().Unix()

	// Normalize cancellation error
	if errors.Is(err, pipeline.ErrCancel) || canceled.Load() {
		canceled.Store(true)
		err = pipeline.ErrCancel
	}

	if err != nil {
		state.Error = err.Error()
	}

	logger.Debug().
		Str("error", state.Error).
		Bool("canceled", canceled.Load()).
		Msg("workflow finished")

	// Ensure all logs/traces are uploaded before finishing
	logger.Debug().Msg("waiting for logs and traces upload")
	uploads.Wait()
	logger.Debug().Msg("logs and traces uploaded")

	// Update workflow state
	doneCtx := runnerCtx
	if doneCtx.Err() != nil {
		doneCtx = shutdownCtx
	}

	if err := r.client.Done(doneCtx, workflow.ID, state); err != nil {
		logger.Error().Err(err).Msg("failed to update workflow status")
	}

	return nil
}

func extractRepositoryName(config *backend.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_REPO"]
}

func extractPipelineNumber(config *backend.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_PIPELINE_NUMBER"]
}

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

// Run executes a workflow using a backend, tracks its state and reports the state back to the server.
func (r *Runner) Run(runnerCtx, shutdownCtx context.Context) error {
	log.Debug().Msg("request next execution")

	// Preserve metadata AND cancellation from runnerCtx.
	meta, _ := metadata.FromOutgoingContext(runnerCtx)
	ctxMeta := metadata.NewOutgoingContext(shutdownCtx, meta)

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
	workflowCtx, _ := context.WithTimeout(ctxMeta, timeout) //nolint:govet
	workflowCtx, cancelWorkflowCtx := context.WithCancelCause(workflowCtx)
	defer cancelWorkflowCtx(nil)

	// recoveryManager is declared here so the cancel listener can mark it as canceled.
	// It will be initialized later after workflow state is set up.
	var recoveryManager *pipeline.RecoveryManager

	// Add sigterm support for internal context.
	// Required to be able to terminate the running workflow by external signals.
	workflowCtx = utils.WithContextSigtermCallback(workflowCtx, func() {
		logger.Error().Msg("received sigterm termination signal")
		// WithContextSigtermCallback would cancel the context too, but  we want our own custom error
		cancelWorkflowCtx(pipeline.ErrCancel)
	})

	// Listen for remote cancel events (UI / API).
	// When canceled, we MUST cancel the workflow context
	// so that workflow execution stop immediately.
	go func() {
		logger.Debug().Msg("start listening for server side cancel signal")

		if canceled, err := r.client.Wait(workflowCtx, workflow.ID); err != nil {
			logger.Error().Err(err).Msg("server returned unexpected err while waiting for workflow to finish run")
			cancelWorkflowCtx(err)
		} else {
			if canceled {
				logger.Debug().Msg("server side cancel signal received")
				if recoveryManager != nil {
					recoveryManager.SetCanceled()
				}
				cancelWorkflowCtx(pipeline.ErrCancel)
			}
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

	if err := r.client.Init(runnerCtx, workflow.ID, state); err != nil {
		logger.Error().Err(err).Msg("signaling workflow initialization to server failed")
		// We have an error, maybe the server is currently unreachable or other server-side errors occurred.
		// So let's clean up and end this not yet started workflow run.
		cancelWorkflowCtx(err)
		return err
	}

	// Initialize recovery manager; if not enabled on server, it will be a no-op
	recoveryManager = pipeline.NewRecoveryManager(r.client, workflow.ID, true)
	if err := recoveryManager.InitRecoveryState(runnerCtx, workflow.Config, int64(timeout.Seconds())); err != nil {
		logger.Warn().Err(err).Msg("failed to initialize recovery state, continuing without recovery")
		recoveryManager = pipeline.NewRecoveryManager(nil, workflow.ID, false)
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
		pipeline.WithRecoveryManager(recoveryManager),
		pipeline.WithDescription(map[string]string{
			"workflow_id":     workflow.ID,
			"repo":            repoName,
			"pipeline_number": pipelineNumber,
		}),
	).Run(runnerCtx)

	state.Finished = time.Now().Unix()

	if err != nil {
		state.Error = err.Error()
		if errors.Is(err, pipeline.ErrCancel) {
			state.Canceled = true
			// cleanup joined error messages
			state.Error = pipeline.ErrCancel.Error()
		}
	}

	logger.Debug().
		Str("error", state.Error).
		Bool("canceled", state.Canceled).
		Msg("workflow finished")

	// Ensure all logs/traces are uploaded before finishing
	logger.Debug().Msg("waiting for logs and traces upload")
	uploads.Wait()
	logger.Debug().Msg("logs and traces uploaded")

	// If workflow is recoverable (context canceled, recovery enabled, not user cancel),
	// skip marking as done. The workflow will be picked up by a new agent after restart.
	if recoveryManager != nil && recoveryManager.IsRecoverable(runnerCtx) {
		logger.Info().Msg("workflow is recoverable, not marking as done")
		return nil
	}

	// Update workflow state
	doneCtx := runnerCtx
	if doneCtx.Err() != nil {
		doneCtx = shutdownCtx
	}

	if err := r.client.Done(doneCtx, workflow.ID, state); err != nil {
		logger.Error().Err(err).Msg("failed to update workflow status")
	} else {
		logger.Debug().Msg("signaling workflow stopped done")
	}

	return nil
}

func extractRepositoryName(config *backend.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_REPO"]
}

func extractPipelineNumber(config *backend.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_PIPELINE_NUMBER"]
}

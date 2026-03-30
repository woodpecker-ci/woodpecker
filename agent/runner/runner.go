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

package runner

import (
	"context"
	"errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/metadata"

	agent_log "go.woodpecker-ci.org/woodpecker/v3/agent/log"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/logging"
	pipeline_runtime "go.woodpecker-ci.org/woodpecker/v3/pipeline/runtime"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// AgentState tracks the number of active and polling workflows on the agent.
// Implemented by agent.State.
type AgentState interface {
	Add(id string, timeout time.Duration, repo, pipeline string)
	Done(id string)
}

// Runner fetches workflows from the server, executes them using a backend,
// and reports state back to the server.
type Runner struct {
	client   rpc.Peer
	filter   rpc.Filter
	hostname string
	counter  AgentState
	backend  backend_types.Backend
}

// NewRunner creates a Runner that polls for work and executes workflows.
func NewRunner(workEngine rpc.Peer, f rpc.Filter, h string, state AgentState, backend backend_types.Backend) Runner {
	return Runner{
		client:   workEngine,
		filter:   f,
		hostname: h,
		counter:  state,
		backend:  backend,
	}
}

// Run fetches the next workflow from the server and executes it.
// runnerCtx is the long-lived agent context; shutdownCtx is used for
// cleanup when runnerCtx is already canceled.
func (r *Runner) Run(runnerCtx, shutdownCtx context.Context) error {
	log.Debug().Msg("request next execution")

	// Preserve gRPC metadata from runnerCtx on the shutdown context so that
	// late RPCs (e.g. Done) still carry the auth token.
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

	timeout := workflowTimeout(workflow)
	repoName := extractRepositoryName(workflow.Config)
	pipelineNumber := extractPipelineNumber(workflow.Config)

	// Track workflow execution in agent state.
	r.counter.Add(workflow.ID, timeout, repoName, pipelineNumber)
	defer r.counter.Done(workflow.ID)

	logger := log.With().
		Str("repo", repoName).
		Str("pipeline", pipelineNumber).
		Str("workflow_id", workflow.ID).
		Logger()

	logger.Debug().Msg("received execution")

	// Build the workflow execution context with timeout, sigterm handler,
	// server-side cancel listener, and lease extension goroutine.
	workflowCtx, cancelWorkflow := buildWorkflowContext(
		ctxMeta, timeout, workflow.ID, r.client, logger,
	)
	defer cancelWorkflow(nil)

	// Signal the server that the workflow is starting.
	startedAt, err := initWorkflow(runnerCtx, r.client, workflow.ID, logger)
	if err != nil {
		cancelWorkflow(err)
		return err
	}

	var uploads sync.WaitGroup

	// Execute the pipeline via the shared runtime.
	err = pipeline_runtime.New(
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

	// Report the final workflow state to the server.
	finalizeWorkflow(runnerCtx, shutdownCtx, r.client, workflow.ID, startedAt, err, &uploads, logger)

	return nil
}

// createLogger returns a logging.Logger that streams step output through the
// RPC client, masking secrets.
func (r *Runner) createLogger(_logger zerolog.Logger, uploads *sync.WaitGroup, workflow *rpc.Workflow) logging.Logger {
	return func(step *backend_types.Step, rc io.ReadCloser) error {
		defer rc.Close()

		logger := _logger.With().
			Str("image", step.Image).
			Logger()

		uploads.Add(1)
		defer uploads.Done()

		var secrets []string
		for _, secret := range workflow.Config.Secrets {
			secrets = append(secrets, secret.Value)
		}

		logger.Debug().Msg("log stream opened")

		logStream := agent_log.NewLineWriter(r.client, step.UUID, secrets...)
		if err := pipeline_utils.CopyLineByLine(logStream, rc, pipeline.MaxLogLineLength); err != nil {
			logger.Error().Err(err).Msg("copy limited logStream part")
		}

		logger.Debug().Msg("log stream copied, close ...")
		return nil
	}
}

// createTracer returns a tracing.TraceFunc that reports step state changes
// back to the server via the RPC client.
func (r *Runner) createTracer(ctxMeta context.Context, uploads *sync.WaitGroup, logger zerolog.Logger, workflow *rpc.Workflow) tracing.TraceFunc {
	return func(st *state.State) error {
		uploads.Add(1)
		defer uploads.Done()

		stepLogger := logger.With().
			Str("image", st.CurrStep.Image).
			Str("workflow_id", workflow.ID).
			Err(st.CurrStepState.Error).
			Int("exit_code", st.CurrStepState.ExitCode).
			Bool("exited", st.CurrStepState.Exited).
			Logger()

		stepState := rpc.StepState{
			StepUUID: st.CurrStep.UUID,
			Exited:   st.CurrStepState.Exited,
			ExitCode: st.CurrStepState.ExitCode,
			Started:  st.CurrStepState.Started,
			Canceled: errors.Is(st.CurrStepState.Error, pipeline_errors.ErrCancel),
			Skipped:  st.CurrStepState.Skipped,
		}
		if st.CurrStepState.Error != nil {
			stepState.Error = st.CurrStepState.Error.Error()
		}
		if st.CurrStepState.Exited {
			stepState.Finished = time.Now().Unix()
		}

		defer func() {
			stepLogger.Debug().Msg("update step status")

			if err := r.client.Update(ctxMeta, workflow.ID, stepState); err != nil {
				stepLogger.Debug().
					Err(err).
					Msg("update step status error")
			}

			stepLogger.Debug().Msg("update step status complete")
		}()

		if st.CurrStepState.Exited {
			return nil
		}
		if st.CurrStep.Environment == nil {
			st.CurrStep.Environment = map[string]string{}
		}

		// TODO: find better way to update this state and move it to pipeline to have the same env in cli-exec
		st.CurrStep.Environment["CI_MACHINE"] = r.hostname
		st.CurrStep.Environment["CI_PIPELINE_STARTED"] = strconv.FormatInt(st.Workflow.Started, 10)
		st.CurrStep.Environment["CI_STEP_STARTED"] = strconv.FormatInt(st.Workflow.Started, 10)
		st.CurrStep.Environment["CI_SYSTEM_PLATFORM"] = runtime.GOOS + "/" + runtime.GOARCH

		return nil
	}
}

// workflowTimeout returns the timeout for the given workflow, defaulting to 1 hour.
func workflowTimeout(workflow *rpc.Workflow) time.Duration {
	if minutes := workflow.Timeout; minutes != 0 {
		return time.Duration(minutes) * time.Minute
	}
	return time.Hour
}

func extractRepositoryName(config *backend_types.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_REPO"]
}

func extractPipelineNumber(config *backend_types.Config) string {
	return config.Stages[0].Steps[0].Environment["CI_PIPELINE_NUMBER"]
}

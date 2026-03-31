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
	"sync"
	"time"

	"github.com/rs/zerolog"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// initWorkflow signals workflow initialization to the server.
func initWorkflow(ctx context.Context, client rpc.Peer, workflowID string) error {
	state := rpc.WorkflowState{
		Started: time.Now().Unix(),
	}
	return client.Init(ctx, workflowID, state)
}

// finalizeWorkflow waits for all async uploads (logs, traces), builds the final
// workflow state from the pipeline error, and reports it to the server via Done.
func finalizeWorkflow(
	runnerCtx, shutdownCtx context.Context,
	client rpc.Peer,
	workflowID string,
	pipelineErr error,
	uploads *sync.WaitGroup,
	logger zerolog.Logger,
) error {
	state := buildFinalState(pipelineErr)

	logger.Debug().
		Str("error", state.Error).
		Bool("canceled", state.Canceled).
		Msg("workflow finished")

	// Ensure all logs/traces are uploaded before finishing.
	logger.Debug().Msg("waiting for logs and traces upload")
	uploads.Wait()
	logger.Debug().Msg("logs and traces uploaded")

	// Pick a context that is still alive for the Done RPC.
	doneCtx := runnerCtx
	if doneCtx.Err() != nil {
		doneCtx = shutdownCtx
	}

	if err := client.Done(doneCtx, workflowID, state); err != nil {
		logger.Error().Err(err).Msg("failed to update workflow status")
		return err
	}

	logger.Debug().Msg("signaling workflow stopped done")
	return nil
}

// buildFinalState translates a pipeline error into the WorkflowState that is
// sent to the server.
func buildFinalState(pipelineErr error) rpc.WorkflowState {
	state := rpc.WorkflowState{
		Finished: time.Now().Unix(),
	}

	if pipelineErr != nil {
		state.Error = pipelineErr.Error()
		if errors.Is(pipelineErr, pipeline_errors.ErrCancel) {
			state.Canceled = true
			state.Error = pipeline_errors.ErrCancel.Error()
		}
	}

	return state
}

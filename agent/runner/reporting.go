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

// initWorkflow tells the server that the workflow is starting and returns the
// recorded start time. If the server is unreachable the error is returned so
// the caller can abort early.
func initWorkflow(
	ctx context.Context,
	client rpc.Peer,
	workflowID string,
	logger zerolog.Logger,
) (startedAt int64, err error) {
	startedAt = time.Now().Unix()

	state := rpc.WorkflowState{
		Started: startedAt,
	}

	if err := client.Init(ctx, workflowID, state); err != nil {
		logger.Error().Err(err).Msg("signaling workflow initialization to server failed")
		return 0, err
	}

	return startedAt, nil
}

// finalizeWorkflow waits for all in-flight log/trace uploads, builds the final
// WorkflowState, and reports it to the server via Done.
func finalizeWorkflow(
	runnerCtx context.Context,
	shutdownCtx context.Context,
	client rpc.Peer,
	workflowID string,
	startedAt int64,
	pipelineErr error,
	uploads *sync.WaitGroup,
	logger zerolog.Logger,
) {
	state := rpc.WorkflowState{
		Started:  startedAt,
		Finished: time.Now().Unix(),
	}

	if pipelineErr != nil {
		state.Error = pipelineErr.Error()
		if errors.Is(pipelineErr, pipeline_errors.ErrCancel) {
			state.Canceled = true
			// Use the canonical cancel message, not a joined error chain.
			state.Error = pipeline_errors.ErrCancel.Error()
		}
	}

	logger.Debug().
		Str("error", state.Error).
		Bool("canceled", state.Canceled).
		Msg("workflow finished")

	// Ensure all logs/traces are uploaded before signaling Done.
	logger.Debug().Msg("waiting for logs and traces upload")
	uploads.Wait()
	logger.Debug().Msg("logs and traces uploaded")

	// Pick the best available context for the Done RPC.
	doneCtx := runnerCtx
	if doneCtx.Err() != nil {
		doneCtx = shutdownCtx
	}

	if err := client.Done(doneCtx, workflowID, state); err != nil {
		logger.Error().Err(err).Msg("failed to update workflow status")
	} else {
		logger.Debug().Msg("signaling workflow stopped done")
	}
}

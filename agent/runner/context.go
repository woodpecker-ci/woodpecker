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
	"time"

	"github.com/rs/zerolog"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
	"go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// computeTimeout returns the workflow timeout derived from the workflow spec.
// Defaults to one hour when no timeout is configured.
func computeTimeout(workflow *rpc.Workflow) time.Duration {
	if minutes := workflow.Timeout; minutes != 0 {
		return time.Duration(minutes) * time.Minute
	}
	return time.Hour
}

// buildWorkflowContext creates the workflow-scoped context with a timeout and
// sigterm handler. The returned CancelCauseFunc must be deferred by the caller.
func buildWorkflowContext(ctxMeta context.Context, timeout time.Duration, logger zerolog.Logger) (context.Context, context.CancelCauseFunc) {
	workflowCtx, _ := context.WithTimeout(ctxMeta, timeout) //nolint:govet
	workflowCtx, cancelFn := context.WithCancelCause(workflowCtx)

	// Add sigterm support — allows external signals to cancel the running workflow.
	workflowCtx = utils.WithContextSigtermCallback(workflowCtx, func() {
		logger.Error().Msg("received sigterm termination signal")
		cancelFn(pipeline_errors.ErrCancel)
	})

	return workflowCtx, cancelFn
}

// startCancelListener spawns a goroutine that waits for the server to signal
// cancellation (e.g. via UI or API) and cancels the workflow context accordingly.
func startCancelListener(workflowCtx context.Context, cancelFn context.CancelCauseFunc, client rpc.Peer, workflowID string, logger zerolog.Logger) {
	go func() {
		logger.Debug().Msg("start listening for server side cancel signal")

		canceled, err := client.Wait(workflowCtx, workflowID)
		switch {
		case err != nil:
			logger.Error().Err(err).Msg("server returned unexpected err while waiting for workflow to finish run")
			cancelFn(err)
		case canceled:
			logger.Debug().Msg("server side cancel signal received")
			cancelFn(pipeline_errors.ErrCancel)
		default:
			logger.Debug().Msg("cancel listener exited normally")
		}
	}()
}

// startLeaseExtender spawns a goroutine that periodically extends the workflow
// deadline on the server so that long-running workflows are not reclaimed.
func startLeaseExtender(workflowCtx context.Context, client rpc.Peer, workflowID string, logger zerolog.Logger) {
	go func() {
		for {
			select {
			case <-workflowCtx.Done():
				logger.Debug().Msg("workflow context done")
				return
			case <-time.After(constant.TaskTimeout / 3):
				logger.Debug().Msg("renewing workflow lease")
				if err := client.Extend(workflowCtx, workflowID); err != nil {
					logger.Error().Err(err).Msg("failed to extend workflow lease")
				}
			}
		}
	}()
}

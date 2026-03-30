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

// buildWorkflowContext creates the single context used for workflow execution.
// It layers timeout, SIGTERM handling, a server-side cancel listener, and
// periodic lease extension on top of the provided parent context.
//
// The returned context.CancelCauseFunc must be deferred by the caller to ensure
// all spawned goroutines are cleaned up.
func buildWorkflowContext(
	parent context.Context,
	timeout time.Duration,
	workflowID string,
	client rpc.Peer,
	logger zerolog.Logger,
) (context.Context, context.CancelCauseFunc) {
	// Apply the workflow timeout.
	ctx, _ := context.WithTimeout(parent, timeout) //nolint:govet

	// Wrap with CancelCause so every cancellation carries a reason.
	ctx, cancel := context.WithCancelCause(ctx)

	// Cancel on SIGTERM so the agent can be stopped gracefully.
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		logger.Error().Msg("received sigterm termination signal")
		cancel(pipeline_errors.ErrCancel)
	})

	// Listen for server-side cancel events (UI / API).
	go listenForCancel(ctx, cancel, workflowID, client, logger)

	// Periodically extend the workflow lease while the context is alive.
	go extendLease(ctx, workflowID, client, logger)

	return ctx, cancel
}

// listenForCancel blocks until the server signals that the workflow should be
// canceled, or until the context is done.
func listenForCancel(
	ctx context.Context,
	cancel context.CancelCauseFunc,
	workflowID string,
	client rpc.Peer,
	logger zerolog.Logger,
) {
	logger.Debug().Msg("start listening for server side cancel signal")

	canceled, err := client.Wait(ctx, workflowID)
	if err != nil {
		logger.Error().Err(err).Msg("server returned unexpected err while waiting for workflow to finish run")
		cancel(err)
		return
	}

	if canceled {
		logger.Debug().Msg("server side cancel signal received")
		cancel(pipeline_errors.ErrCancel)
		return
	}

	// Wait returned without error and without cancel — workflow finished normally.
	logger.Debug().Msg("cancel listener exited normally")
}

// extendLease periodically renews the workflow lease on the server so the
// server knows the agent is still working on it.
func extendLease(
	ctx context.Context,
	workflowID string,
	client rpc.Peer,
	logger zerolog.Logger,
) {
	for {
		select {
		case <-ctx.Done():
			logger.Debug().Msg("workflow context done")
			return

		case <-time.After(constant.TaskTimeout / 3):
			logger.Debug().Msg("renewing workflow lease")
			if err := client.Extend(ctx, workflowID); err != nil {
				logger.Error().Err(err).Msg("failed to extend workflow lease")
			}
		}
	}
}

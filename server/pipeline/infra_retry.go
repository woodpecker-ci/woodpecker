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

package pipeline

import (
	"context"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// shouldRetryInfraFailure reports whether a finished pipeline failed solely
// because of an infrastructure event (a step the backend flagged as an
// InfraFailure, e.g. spot-node preemption or pod eviction) and is therefore
// safe to restart automatically.
//
// It deliberately refuses to retry when any genuine (non-infra) step failure
// is present: restarting would just burn resources reproducing a real error.
// It is a pure function so the decision is unit-testable without a store.
func shouldRetryInfraFailure(pipeline *model.Pipeline, maxAttempts int64) bool {
	if maxAttempts <= 0 {
		return false
	}
	// Only failed pipelines are candidates. StatusError (config/parse
	// problems) and StatusKilled (user/superseded cancels) are never infra.
	if pipeline.Status != model.StatusFailure {
		return false
	}
	if pipeline.InfraRetryCount >= maxAttempts {
		return false
	}

	sawInfraFailure := false
	for _, workflow := range pipeline.Workflows {
		for _, step := range workflow.Children {
			// A step the backend flagged as an infra failure makes the
			// pipeline retryable, whether it surfaced as a failure or was
			// reported killed (e.g. a fail-fast cancel racing the exit).
			if step.InfraFailure {
				sawInfraFailure = true
				continue
			}
			if step.State != model.StatusFailure {
				// Skipped, killed (fail-fast cascade) and successful steps do
				// not, on their own, make the pipeline unretryable.
				continue
			}
			// A step allowed to fail never fails the pipeline.
			if step.Failure == model.FailureIgnore {
				continue
			}
			// A genuine failure: a retry will not change the outcome.
			return false
		}
	}

	return sawInfraFailure
}

// RetryOnInfraFailure restarts pipeline when it failed only because of an
// infrastructure event and the configured attempt budget is not yet spent.
// It returns true when a restart was triggered.
//
// Automatic retries have no initiating user, so it runs as the repo owner —
// the same identity cron-triggered pipelines use.
func RetryOnInfraFailure(ctx context.Context, store store.Store, repo *model.Repo, pipeline *model.Pipeline) (bool, error) {
	if !shouldRetryInfraFailure(pipeline, server.Config.Pipeline.InfraRetryMaxAttempts) {
		return false, nil
	}

	user, err := store.GetUser(repo.UserID)
	if err != nil {
		return false, fmt.Errorf("infra-retry: cannot load repo owner %d: %w", repo.UserID, err)
	}

	// Atomically claim the single retry for this pipeline. For a multi-workflow
	// pipeline whose final workflows finish concurrently, more than one Done
	// can observe the pipeline as just-terminal and reach here; only the caller
	// that wins the claim restarts it.
	claimed, err := store.ClaimInfraRetry(pipeline.ID)
	if err != nil {
		return false, fmt.Errorf("infra-retry: cannot claim retry: %w", err)
	}
	if !claimed {
		return false, nil
	}

	// restart with incrementInfraRetry=true bumps InfraRetryCount on the new
	// pipeline so it is persisted atomically by CreatePipeline. This bounds the
	// chain at InfraRetryMaxAttempts with no separate, individually-fallible
	// update that could otherwise leave a stale count and retry forever.
	if _, err := restart(ctx, store, pipeline, user, repo, nil, true); err != nil {
		return false, fmt.Errorf("infra-retry: restart failed: %w", err)
	}

	return true, nil
}

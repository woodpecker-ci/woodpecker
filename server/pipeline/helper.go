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

package pipeline

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// maxConcurrentStatusUpdates bounds the parallel commit-status calls per
// pipeline. Forges rate limit concurrent writes from a single user much more
// aggressively than sequential ones, so this stays deliberately low: it is
// enough to keep pipelines with many workflows from taking tens of seconds,
// without looking like a burst.
const maxConcurrentStatusUpdates = 4

// updatePipelineStatus posts one commit status per workflow. Every workflow is
// attempted even if some fail, so a single forge error cannot leave the
// remaining workflows without a status.
func updatePipelineStatus(ctx context.Context, forge forge.Forge, pipeline *model.Pipeline, repo *model.Repo, user *model.User) {
	// setting one status per workflow sequentially delays pipelines with many
	// workflows by tens of seconds, so post them with bounded concurrency
	var wg sync.WaitGroup
	var failed atomic.Int64
	sem := make(chan struct{}, maxConcurrentStatusUpdates)

	for _, workflow := range pipeline.Workflows {
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				failed.Add(1)
				return
			}

			if err := forge.Status(ctx, user, repo, pipeline, workflow); err != nil {
				failed.Add(1)
				log.Error().Err(err).Msgf("error setting commit status for %s/%d", repo.FullName, pipeline.Number)
			}
		}()
	}
	wg.Wait()

	// individual failures are logged above, but a pipeline whose statuses all
	// failed is invisible on the forge and must not be silent
	if n := failed.Load(); n > 0 {
		log.Error().Msgf("failed to set %d of %d commit statuses for %s/%d", n, len(pipeline.Workflows), repo.FullName, pipeline.Number)
	}
}

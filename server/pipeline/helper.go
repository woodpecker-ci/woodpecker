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

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// maxConcurrentStatusUpdates bounds the parallel commit-status calls per
// pipeline, so pipelines with many workflows do not trip forge rate limits.
const maxConcurrentStatusUpdates = 10

func updatePipelineStatus(ctx context.Context, forge forge.Forge, pipeline *model.Pipeline, repo *model.Repo, user *model.User) {
	// setting one status per workflow sequentially delays pipelines with many
	// workflows by tens of seconds, so post them concurrently
	var group errgroup.Group
	group.SetLimit(maxConcurrentStatusUpdates)
	for _, workflow := range pipeline.Workflows {
		group.Go(func() error {
			err := forge.Status(ctx, user, repo, pipeline, workflow)
			if err != nil {
				log.Error().Err(err).Msgf("error setting commit status for %s/%d", repo.FullName, pipeline.Number)
			}
			return nil
		})
	}
	_ = group.Wait()
}

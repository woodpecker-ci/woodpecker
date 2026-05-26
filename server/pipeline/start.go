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

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// dispatchPipeline cancels superseded pipelines for the same context and
// pushes the pipeline's workflows onto the agent queue. Publishing the status
// to subscribers is the caller's responsibility, so the pipeline is not
// published here.
func dispatchPipeline(ctx context.Context, forge forge.Forge, store store.Store, activePipeline *model.Pipeline, user *model.User, repo *model.Repo, pipelineItems []*builder.Item) (*model.Pipeline, error) {
	if err := cancelPreviousPipelines(ctx, forge, store, activePipeline, repo, user); err != nil {
		// should not be breaking
		log.Error().Err(err).Msg("failed to cancel previous pipelines")
	}

	if err := queuePipeline(ctx, repo, activePipeline, pipelineItems); err != nil {
		log.Error().Err(err).Msg("queuePipeline")
		return nil, err
	}

	return activePipeline, nil
}

func publishPipeline(ctx context.Context, forge forge.Forge, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User) {
	if err := publishToTopic(ctx, pipeline, repo); err != nil {
		log.Error().Err(err).Msg("could not push pipeline status change to pubsub provider")
	}
	updatePipelineStatus(ctx, forge, pipeline, repo, repoUser)
}

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
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// start a pipeline, make sure it was stored persistent in the store before.
func start(ctx context.Context, forge forge.Forge, store store.Store, activePipeline *model.Pipeline, user *model.User, repo *model.Repo, pipelineItems []*builder.Item) (*model.Pipeline, error) {
	// call to cancel previous pipelines if needed
	if err := cancelPreviousPipelines(ctx, forge, store, activePipeline, repo, user); err != nil {
		// should be not breaking
		log.Error().Err(err).Msg("failed to cancel previous pipelines")
	}

	tasks, err := pipelineTasks(repo, activePipeline, pipelineItems)
	if err != nil {
		return nil, err
	}

	// announce the new pipeline to UI subscribers and enqueue its tasks in one go
	if err := server.Config.Services.Scheduler.StartPipeline(ctx, repo, activePipeline, tasks); err != nil {
		log.Error().Err(err).Msg("startPipeline")
		return nil, err
	}

	updatePipelineStatus(ctx, forge, activePipeline, repo, user)

	return activePipeline, nil
}

func publishPipeline(ctx context.Context, forge forge.Forge, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User) {
	if err := server.Config.Services.Scheduler.PublishPipelineEvent(ctx, repo, pipeline); err != nil {
		log.Error().Err(err).Msg("could not push pipeline status change to pubsub provider")
	}
	updatePipelineStatus(ctx, forge, pipeline, repo, repoUser)
}

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

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline/stepbuilder"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

// start a pipeline, make sure it was stored persistent in the store before
func start(ctx context.Context, store store.Store, activePipeline *model.Pipeline, user *model.User, repo *model.Repo, pipelineItems []*stepbuilder.Item) (*model.Pipeline, error) {
	// call to cancel previous pipelines if needed
	if err := cancelPreviousPipelines(ctx, store, activePipeline, repo, user); err != nil {
		// should be not breaking
		log.Error().Err(err).Msg("failed to cancel previous pipelines")
	}

	if err := store.WorkflowsCreate(activePipeline.Workflows); err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error persisting steps for %s#%d", repo.FullName, activePipeline.Number)
		return nil, err
	}

	publishPipeline(ctx, activePipeline, repo, user)

	if err := queuePipeline(ctx, repo, pipelineItems); err != nil {
		log.Error().Err(err).Msg("queuePipeline")
		return nil, err
	}

	return activePipeline, nil
}

func prepareStart(ctx context.Context, store store.Store, activePipeline *model.Pipeline, user *model.User, repo *model.Repo) error {
	if err := store.WorkflowsCreate(activePipeline.Workflows); err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error persisting steps for %s#%d", repo.FullName, activePipeline.Number)
		return err
	}

	publishPipeline(ctx, activePipeline, repo, user)
	return nil
}

func publishPipeline(ctx context.Context, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User) {
	publishToTopic(pipeline, repo)
	updatePipelineStatus(ctx, pipeline, repo, repoUser)
}

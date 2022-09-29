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

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// start a build, make sure it was stored persistent in the store before
func start(ctx context.Context, store store.Store, activeBuild *model.Pipeline, user *model.User, repo *model.Repo, buildItems []*shared.PipelineItem) (*model.Pipeline, error) {
	// call to cancel previous builds if needed
	if err := cancelPreviousPipelines(ctx, store, activeBuild, repo); err != nil {
		// should be not breaking
		log.Error().Err(err).Msg("Failed to cancel previous builds")
	}

	if err := store.ProcCreate(activeBuild.Procs); err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error persisting procs for %s#%d", repo.FullName, activeBuild.Number)
		return nil, err
	}

	if err := publishToTopic(ctx, activeBuild, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	if err := queueBuild(activeBuild, repo, buildItems); err != nil {
		log.Error().Err(err).Msg("queueBuild")
		return nil, err
	}

	if err := updatePipelineStatus(ctx, activeBuild, repo, user); err != nil {
		log.Error().Err(err).Msg("updateBuildStatus")
	}

	return activeBuild, nil
}

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
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/forge/loader"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Decline update the status to declined for blocked pipeline because of a gated repo
func Decline(ctx context.Context, store store.Store, pipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	forge, err := loader.GetForge(store, repo)
	if err != nil {
		msg := fmt.Sprintf("failure to load forge for repo '%s'", repo.FullName)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if pipeline.Status != model.StatusBlocked {
		return nil, fmt.Errorf("cannot decline a pipeline with status %s", pipeline.Status)
	}

	_, err = UpdateToStatusDeclined(store, *pipeline, user.Login)
	if err != nil {
		return nil, fmt.Errorf("error updating pipeline. %s", err)
	}

	if pipeline.Steps, err = store.StepList(pipeline); err != nil {
		log.Error().Err(err).Msg("can not get step list from store")
	}
	if pipeline.Steps, err = model.Tree(pipeline.Steps); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
	}

	if err := updatePipelineStatus(ctx, forge, pipeline, repo, user); err != nil {
		log.Error().Err(err).Msg("updateBuildStatus")
	}

	if err := publishToTopic(ctx, pipeline, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	return pipeline, nil
}

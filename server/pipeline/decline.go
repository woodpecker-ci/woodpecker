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

	"go.woodpecker-ci.org/woodpecker/server"
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/store"
)

// Decline updates the status to declined for blocked pipelines because of a gated repo
func Decline(ctx context.Context, store store.Store, pipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	forge, err := server.Config.Services.Forge.FromRepo(repo)
	if err != nil {
		msg := fmt.Sprintf("failure to load forge for repo '%s'", repo.FullName)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if pipeline.Status != model.StatusBlocked {
		return nil, fmt.Errorf("cannot decline a pipeline with status %s", pipeline.Status)
	}

	pipeline, err = UpdateToStatusDeclined(store, *pipeline, user.Login)
	if err != nil {
		return nil, fmt.Errorf("error updating pipeline. %w", err)
	}

	if pipeline.Workflows, err = store.WorkflowGetTree(pipeline); err != nil {
		log.Error().Err(err).Msg("can not build tree from step list")
	}

	updatePipelineStatus(ctx, forge, pipeline, repo, user)

	publishToTopic(pipeline, repo)

	return pipeline, nil
}

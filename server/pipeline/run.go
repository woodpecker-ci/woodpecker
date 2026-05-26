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
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// loadForge resolves the forge for a repo, wrapping the failure in a uniform
// error. Create, Approve and Restart all start this way.
func loadForge(repo *model.Repo) (forge.Forge, error) {
	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("failure to load forge for repo '%s'", repo.FullName)
		return nil, fmt.Errorf("failure to load forge for repo '%s'", repo.FullName)
	}
	return _forge, nil
}

// finishPipeline publishes the pipeline status and, unless the pipeline is
// blocked awaiting approval, dispatches its workflows to the agent queue. It is
// the shared tail of Create, Approve and Restart: the caller has already
// persisted the pipeline and its workflows and set the final pre-dispatch
// status.
func finishPipeline(ctx context.Context, _forge forge.Forge, _store store.Store, pipeline *model.Pipeline, user *model.User, repo *model.Repo, pipelineItems []*builder.Item) (*model.Pipeline, error) {
	publishPipeline(ctx, _forge, pipeline, repo, user)

	// a gated pipeline stops here until it is approved
	if pipeline.Status == model.StatusBlocked {
		return pipeline, nil
	}

	pipeline, err := dispatchPipeline(ctx, _forge, _store, pipeline, user, repo, pipelineItems)
	if err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("failed to dispatch pipeline for %s", repo.FullName)
		return nil, errors.New("failed to dispatch pipeline for " + repo.FullName)
	}

	return pipeline, nil
}

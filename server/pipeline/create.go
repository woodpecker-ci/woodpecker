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
	"time"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Create a new pipeline and start it
func Create(ctx context.Context, _store store.Store, repo *model.Repo, pipeline *model.Pipeline) (*model.Pipeline, error) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	// if the forge has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the pipeline.
	if refresher, ok := server.Config.Services.Forge.(forge.Refresher); ok {
		refreshed, err := refresher.Refresh(ctx, repoUser)
		if err != nil {
			log.Error().Err(err).Msgf("failed to refresh oauth2 token for repoUser: %s", repoUser.Login)
		} else if refreshed {
			if err := _store.UpdateUser(repoUser); err != nil {
				log.Error().Err(err).Msgf("error while updating repoUser: %s", repoUser.Login)
				// move forward
			}
		}
	}

	var (
		forgeYamlConfigs []*types.FileMeta
		configFetchErr   error
		filtered         bool
		parseErr         error
	)

	// fetch the pipeline file from the forge
	configFetcher := forge.NewConfigFetcher(server.Config.Services.Forge, server.Config.Services.Timeout, server.Config.Services.ConfigService, repoUser, repo, pipeline)
	forgeYamlConfigs, configFetchErr = configFetcher.Fetch(ctx)
	if configFetchErr == nil {
		filtered, parseErr = checkIfFiltered(pipeline, forgeYamlConfigs)
		if parseErr == nil {
			if filtered {
				err := ErrFiltered{Msg: "branch does not match restrictions defined in yaml"}
				log.Debug().Str("repo", repo.FullName).Msgf("%v", err)
				return nil, err
			}

			if zeroSteps(pipeline, forgeYamlConfigs) {
				err := ErrFiltered{Msg: "step conditions yield zero runnable steps"}
				log.Debug().Str("repo", repo.FullName).Msgf("%v", err)
				return nil, err
			}
		}
	}

	// update some pipeline fields
	pipeline.RepoID = repo.ID
	pipeline.Verified = true
	pipeline.Status = model.StatusPending

	if configFetchErr != nil {
		log.Debug().Str("repo", repo.FullName).Err(configFetchErr).Msgf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, pipeline.Ref, repoUser.Login)
		pipeline.Started = time.Now().Unix()
		pipeline.Finished = pipeline.Started
		pipeline.Status = model.StatusError
		pipeline.Error = fmt.Sprintf("pipeline definition not found in %s", repo.FullName)
	} else if parseErr != nil {
		log.Debug().Str("repo", repo.FullName).Err(parseErr).Msg("failed to parse yaml")
		pipeline.Started = time.Now().Unix()
		pipeline.Finished = pipeline.Started
		pipeline.Status = model.StatusError
		pipeline.Error = fmt.Sprintf("failed to parse pipeline: %s", parseErr.Error())
	} else if repo.IsGated {
		// TODO(336) extend gated feature with an allow/block List
		pipeline.Status = model.StatusBlocked
	}

	err = _store.CreatePipeline(pipeline, pipeline.Steps...)
	if err != nil {
		msg := fmt.Sprintf("failure to save pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	// persist the pipeline config for historical correctness, restarts, etc
	for _, forgeYamlConfig := range forgeYamlConfigs {
		_, err := findOrPersistPipelineConfig(_store, pipeline, forgeYamlConfig)
		if err != nil {
			msg := fmt.Sprintf("failure to find or persist pipeline config for %s", repo.FullName)
			log.Error().Err(err).Msg(msg)
			return nil, fmt.Errorf(msg)
		}
	}

	if pipeline.Status == model.StatusError {
		if err := publishToTopic(ctx, pipeline, repo); err != nil {
			log.Error().Err(err).Msg("publishToTopic")
		}

		if err := updatePipelineStatus(ctx, pipeline, repo, repoUser); err != nil {
			log.Error().Err(err).Msg("updatePipelineStatus")
		}

		return pipeline, nil
	}

	pipeline, pipelineItems, err := createPipelineItems(ctx, _store, pipeline, repoUser, repo, forgeYamlConfigs, nil)
	if err != nil {
		msg := fmt.Sprintf("failure to createPipelineItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if pipeline.Status == model.StatusBlocked {
		if err := publishToTopic(ctx, pipeline, repo); err != nil {
			log.Error().Err(err).Msg("publishToTopic")
		}

		if err := updatePipelineStatus(ctx, pipeline, repo, repoUser); err != nil {
			log.Error().Err(err).Msg("updatePipelineStatus")
		}

		return pipeline, nil
	}

	pipeline, err = start(ctx, _store, pipeline, repoUser, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return pipeline, nil
}

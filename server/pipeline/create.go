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
	"regexp"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/pipeline/errors"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

var skipPipelineRegex = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

// Create a new pipeline and start it
func Create(ctx context.Context, _store store.Store, repo *model.Repo, pipeline *model.Pipeline) (*model.Pipeline, error) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	skipMatch := skipPipelineRegex.FindString(pipeline.Message)
	if len(skipMatch) > 0 {
		log.Debug().Str("repo", repo.FullName).Msgf("ignoring pipeline as skip-ci was found in the commit (%s) message '%s'", pipeline.Commit, pipeline.Message)
		return nil, ErrFiltered
	}

	// If the forge has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the pipeline.
	forge.Refresh(ctx, server.Config.Services.Forge, _store, repoUser)

	// update some pipeline fields
	pipeline.RepoID = repo.ID
	pipeline.Status = model.StatusPending

	// fetch the pipeline file from the forge
	configFetcher := forge.NewConfigFetcher(server.Config.Services.Forge, server.Config.Services.Timeout, server.Config.Services.ConfigService, repoUser, repo, pipeline)
	forgeYamlConfigs, configFetchErr := configFetcher.Fetch(ctx)

	if configFetchErr != nil {
		log.Debug().Str("repo", repo.FullName).Err(configFetchErr).Msgf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, pipeline.Ref, repoUser.Login)
		return nil, persistPipelineWithErr(ctx, _store, pipeline, repo, repoUser, fmt.Errorf("pipeline definition not found in %s", repo.FullName))
	}

	pipelineItems, parseErr := parsePipeline(_store, pipeline, repoUser, repo, forgeYamlConfigs, nil)
	if errors.HasBlockingErrors(parseErr) {
		log.Debug().Str("repo", repo.FullName).Err(parseErr).Msg("failed to parse yaml")
		return nil, persistPipelineWithErr(ctx, _store, pipeline, repo, repoUser, parseErr)
	} else if parseErr != nil {
		pipeline.Errors = errors.GetPipelineErrors(parseErr)
	}

	if len(pipelineItems) == 0 {
		log.Debug().Str("repo", repo.FullName).Msg(ErrFiltered.Error())
		return nil, ErrFiltered
	}

	setGatedState(repo, pipeline)

	err = _store.CreatePipeline(pipeline)
	if err != nil {
		msg := fmt.Errorf("failed to save pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg.Error())
		return nil, msg
	}

	pipeline = setPipelineStepsOnPipeline(pipeline, pipelineItems)

	// persist the pipeline config for historical correctness, restarts, etc
	var configs []*model.Config
	for _, forgeYamlConfig := range forgeYamlConfigs {
		config, err := findOrPersistPipelineConfig(_store, pipeline, forgeYamlConfig)
		if err != nil {
			msg := fmt.Sprintf("failed to find or persist pipeline config for %s", repo.FullName)
			log.Error().Err(err).Msg(msg)
			return nil, fmt.Errorf(msg)
		}
		configs = append(configs, config)
	}
	// link pipeline to persisted configs
	if err := linkPipelineConfigs(_store, configs, pipeline.ID); err != nil {
		msg := fmt.Sprintf("failed to find or persist pipeline config for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if pipeline.Status == model.StatusBlocked {
		publishPipeline(ctx, pipeline, repo, repoUser)
		return pipeline, nil
	}

	pipeline, err = start(ctx, _store, pipeline, repoUser, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failed to start pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return pipeline, nil
}

func persistPipelineWithErr(ctx context.Context, _store store.Store, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User, err error) error {
	pipeline.Started = time.Now().Unix()
	pipeline.Finished = pipeline.Started
	pipeline.Status = model.StatusError
	pipeline.Errors = errors.GetPipelineErrors(err)
	dbErr := _store.CreatePipeline(pipeline)
	if dbErr != nil {
		msg := fmt.Errorf("failed to save pipeline for %s", repo.FullName)
		log.Error().Err(dbErr).Msg(msg.Error())
		return msg
	}

	publishPipeline(ctx, pipeline, repo, repoUser)

	return nil
}

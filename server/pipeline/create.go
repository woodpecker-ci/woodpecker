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
	"errors"
	"fmt"
	"regexp"

	"github.com/rs/zerolog/log"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

var skipPipelineRegex = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

// Create a new pipeline and start it.
func Create(ctx context.Context, _store store.Store, repo *model.Repo, pipeline *model.Pipeline) (*model.Pipeline, error) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, errors.New(msg)
	}

	if pipeline.Event == model.EventPush || pipeline.Event == model.EventPull || pipeline.Event == model.EventPullClosed {
		skipMatch := skipPipelineRegex.FindString(pipeline.Message)
		if len(skipMatch) > 0 {
			ref := pipeline.Commit
			if len(ref) == 0 {
				ref = pipeline.Ref
			}
			log.Debug().Str("repo", repo.FullName).Msgf("ignoring pipeline as skip-ci was found in the commit (%s) message '%s'", ref, pipeline.Message)
			return nil, ErrFiltered
		}
	}

	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		msg := fmt.Sprintf("failure to load forge for repo '%s'", repo.FullName)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, errors.New(msg)
	}

	// If the forge has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the pipeline.
	forge.Refresh(ctx, _forge, _store, repoUser)

	// update some pipeline fields
	pipeline.RepoID = repo.ID
	pipeline.Status = model.StatusCreated
	setGatedState(repo, pipeline)
	err = _store.CreatePipeline(pipeline)
	if err != nil {
		msg := fmt.Errorf("failed to save pipeline for %s", repo.FullName)
		log.Error().Str("repo", repo.FullName).Err(err).Msg(msg.Error())
		return nil, msg
	}

	// fetch the pipeline file from the forge
	configService := server.Config.Services.Manager.ConfigServiceFromRepo(repo)
	forgeYamlConfigs, configFetchErr := configService.Fetch(ctx, _forge, repoUser, repo, pipeline, nil, false)
	if errors.Is(configFetchErr, &forge_types.ErrConfigNotFound{}) {
		log.Debug().Str("repo", repo.FullName).Err(configFetchErr).Msgf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, pipeline.Ref, repoUser.Login)
		if err := _store.DeletePipeline(pipeline); err != nil {
			log.Error().Str("repo", repo.FullName).Err(err).Msg("failed to delete pipeline without config")
		}

		return nil, ErrFiltered
	} else if configFetchErr != nil {
		log.Debug().Str("repo", repo.FullName).Err(configFetchErr).Msgf("error while fetching config '%s' in '%s' with user: '%s'", repo.Config, pipeline.Ref, repoUser.Login)
		return nil, updatePipelineWithErr(ctx, _forge, _store, pipeline, repo, repoUser, fmt.Errorf("could not load config from forge: %w", err))
	}

	pipelineItems, parseErr := parsePipeline(_forge, _store, pipeline, repoUser, repo, forgeYamlConfigs, nil)
	if pipeline_errors.HasBlockingErrors(parseErr) {
		log.Debug().Str("repo", repo.FullName).Err(parseErr).Msg("failed to parse yaml")
		return pipeline, updatePipelineWithErr(ctx, _forge, _store, pipeline, repo, repoUser, parseErr)
	} else if parseErr != nil {
		pipeline.Errors = pipeline_errors.GetPipelineErrors(parseErr)
	}

	if len(pipelineItems) == 0 {
		log.Debug().Str("repo", repo.FullName).Msg(ErrFiltered.Error())
		if err := _store.DeletePipeline(pipeline); err != nil {
			log.Error().Str("repo", repo.FullName).Err(err).Msg("failed to delete empty pipeline")
		}

		return nil, ErrFiltered
	}

	pipeline = setPipelineStepsOnPipeline(pipeline, pipelineItems)

	// persist the pipeline config for historical correctness, restarts, etc
	var configs []*model.Config
	for _, forgeYamlConfig := range forgeYamlConfigs {
		config, err := findOrPersistPipelineConfig(_store, pipeline, forgeYamlConfig)
		if err != nil {
			msg := fmt.Sprintf("failed to find or persist pipeline config for %s", repo.FullName)
			log.Error().Err(err).Msg(msg)
			return nil, errors.New(msg)
		}
		configs = append(configs, config)
	}
	// link pipeline to persisted configs
	if err := linkPipelineConfigs(_store, configs, pipeline.ID); err != nil {
		msg := fmt.Sprintf("failed to find or persist pipeline config for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, errors.New(msg)
	}

	if err := prepareStart(ctx, _forge, _store, pipeline, repoUser, repo); err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error preparing pipeline for %s#%d", repo.FullName, pipeline.Number)
		return nil, err
	}

	if pipeline.Status == model.StatusBlocked {
		return pipeline, nil
	}

	if err := updatePipelinePending(ctx, _forge, _store, pipeline, repo, repoUser); err != nil {
		return nil, err
	}

	pipeline, err = start(ctx, _forge, _store, pipeline, repoUser, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failed to start pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, errors.New(msg)
	}

	return pipeline, nil
}

func updatePipelineWithErr(ctx context.Context, _forge forge.Forge, _store store.Store, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User, err error) error {
	_pipeline, err := UpdateToStatusError(_store, *pipeline, err)
	if err != nil {
		return err
	}
	// update value in ref
	*pipeline = *_pipeline

	publishPipeline(ctx, _forge, pipeline, repo, repoUser)

	return nil
}

func updatePipelinePending(ctx context.Context, _forge forge.Forge, _store store.Store, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User) error {
	_pipeline, err := UpdateToStatusPending(_store, *pipeline, "")
	if err != nil {
		return err
	}
	// update value in ref
	*pipeline = *_pipeline

	publishPipeline(ctx, _forge, pipeline, repo, repoUser)

	return nil
}

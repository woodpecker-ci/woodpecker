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
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

// Create a new pipeline and start it.
func Create(ctx context.Context, _store store.Store, repo *model.Repo, pipeline *model.Pipeline) (*model.Pipeline, error) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, errors.New(msg)
	}

	if constraint.IsSkipCommitMessage(metadata.Event(pipeline.Event), pipeline.Message) {
		ref := pipeline.Commit
		if len(ref) == 0 {
			ref = pipeline.Ref
		}
		log.Debug().Str("repo", repo.FullName).Msgf("ignoring pipeline as skip-ci was found in the commit (%s) message '%s'", ref, pipeline.Message)
		return nil, ErrFiltered
	}

	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		msg := fmt.Sprintf("failure to load forge for repo '%s'", repo.FullName)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, errors.New(msg)
	}

	// If the repoUser has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the pipeline.
	forge.Refresh(ctx, _forge, _store, repoUser)

	// update some pipeline fields
	pipeline.RepoID = repo.ID
	pipeline.Status = model.StatusCreated
	pipeline.Version = version.String()
	setApprovalState(repo, pipeline)
	err = _store.CreatePipeline(pipeline)
	if err != nil {
		msg := fmt.Errorf("failed to save pipeline for %s", repo.FullName)
		log.Error().Str("repo", repo.FullName).Err(err).Msg(msg.Error())
		return nil, msg
	}

	// fetch the pipeline file from the forge
	phaseStart := time.Now()
	configService := server.Config.Services.Manager.ConfigServiceFromRepo(repo)
	forgeYamlConfigs, configFetchErr := configService.Fetch(ctx, _forge, repoUser, repo, pipeline, nil, false)
	configFetchDuration := time.Since(phaseStart)
	switch {
	case errors.Is(configFetchErr, &forge_types.ErrConfigNotFound{}):
		log.Debug().Str("repo", repo.FullName).Err(configFetchErr).Msgf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, pipeline.Ref, repoUser.Login)
		if err := _store.DeletePipeline(pipeline); err != nil {
			log.Error().Str("repo", repo.FullName).Err(err).Msg("failed to delete pipeline without config")
		}

		return nil, ErrFiltered
	case configFetchErr != nil && forgeYamlConfigs != nil:
		// unexpected status code from config endpoint - using previous config as fallback
		log.Warn().Str("repo", repo.FullName).Err(configFetchErr).Msgf("error while fetching config '%s' in '%s' with user: '%s', will fallback to old config", repo.Config, pipeline.Ref, repoUser.Login)
	case configFetchErr != nil:
		// error while fetching config - not using the old config
		log.Error().Str("repo", repo.FullName).Err(configFetchErr).Msgf("error while fetching config '%s' in '%s' with user: '%s', and did not get any config", repo.Config, pipeline.Ref, repoUser.Login)
		return nil, updatePipelineWithErr(ctx, _forge, _store, pipeline, repo, repoUser, fmt.Errorf("could not load config from forge: %w", configFetchErr))
	}

	phaseStart = time.Now()
	currentPipeline, pipelineItems, parseErr, err := createPipelineItems(ctx, _forge, _store, pipeline, repoUser, repo, forgeYamlConfigs, nil, false)
	compileDuration := time.Since(phaseStart)
	*pipeline = *currentPipeline
	if handleParseErrors(pipeline, parseErr) {
		log.Debug().Str("repo", repo.FullName).Err(parseErr).Msg("failed to parse yaml")
		return pipeline, updatePipelineWithErr(ctx, _forge, _store, pipeline, repo, repoUser, parseErr)
	}
	if err != nil {
		return nil, fmt.Errorf("createPipelineItems failed: %w", err)
	}

	if len(pipelineItems) == 0 {
		log.Debug().Str("repo", repo.FullName).Msg(ErrFiltered.Error())
		if err := _store.DeletePipeline(pipeline); err != nil {
			log.Error().Str("repo", repo.FullName).Err(err).Msg("failed to delete empty pipeline")
		}

		return nil, ErrFiltered
	}

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

	phaseStart = time.Now()
	publishPipelineEvent(ctx, pipeline, repo)

	if pipeline.Status == model.StatusBlocked {
		// blocked pipelines never reach start(), so their statuses are posted here
		updatePipelineStatus(ctx, _forge, pipeline, repo, repoUser)
		return pipeline, nil
	}
	publishDuration := time.Since(phaseStart)

	phaseStart = time.Now()
	if err := updatePipelinePending(ctx, _store, pipeline, repo); err != nil {
		return nil, err
	}
	pendingDuration := time.Since(phaseStart)

	phaseStart = time.Now()
	startedPipeline, err := start(ctx, _forge, _store, pipeline, repoUser, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failed to start pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		// transition the pipeline to error and post the statuses, so neither the
		// pipeline nor the commit is stuck in a pending state forever
		if uErr := updatePipelineWithErr(ctx, _forge, _store, pipeline, repo, repoUser, err); uErr != nil {
			log.Error().Err(uErr).Msgf("error setting error status of pipeline for %s#%d", repo.FullName, pipeline.Number)
		}
		return nil, errors.New(msg)
	}
	pipeline = startedPipeline

	log.Debug().
		Str("repo", repo.FullName).
		Int64("pipeline", pipeline.Number).
		Int("workflows", len(pipeline.Workflows)).
		Dur("config_fetch", configFetchDuration).
		Dur("compile", compileDuration).
		Dur("publish_created", publishDuration).
		Dur("publish_pending", pendingDuration).
		Dur("start", time.Since(phaseStart)).
		Msg("pipeline creation timing")

	return pipeline, nil
}

func updatePipelineWithErr(ctx context.Context, _forge forge.Forge, _store store.Store, pipeline *model.Pipeline, repo *model.Repo, repoUser *model.User, err error) error {
	_pipeline, err := UpdateToStatusError(_store, *pipeline, err)
	if err != nil {
		return err
	}
	// update value in ref
	*pipeline = *_pipeline

	// transition the persisted workflows as well: forges post the workflow
	// states as commit statuses, so leaving them pending would publish
	// statuses that never resolve
	for _, workflow := range pipeline.Workflows {
		if workflow.State != model.StatusPending {
			continue
		}
		workflow.State = model.StatusError
		workflow.Finished = pipeline.Finished
		if uErr := _store.WorkflowUpdate(workflow); uErr != nil {
			log.Error().Err(uErr).Msgf("cannot update workflow with id %d state", workflow.ID)
		}
	}

	publishPipeline(ctx, _forge, pipeline, repo, repoUser)

	return nil
}

func updatePipelinePending(ctx context.Context, _store store.Store, pipeline *model.Pipeline, repo *model.Repo) error {
	_pipeline, err := UpdateToStatusPending(_store, *pipeline, "")
	if err != nil {
		return err
	}
	// update value in ref
	*pipeline = *_pipeline

	// the forge statuses are posted by start() right after
	publishPipelineEvent(ctx, pipeline, repo)

	return nil
}

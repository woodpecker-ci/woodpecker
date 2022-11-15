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

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge/loader"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Restart a pipeline by creating a new one out of the old and start it
func Restart(ctx context.Context, store store.Store, lastBuild *model.Pipeline, user *model.User, repo *model.Repo, envs map[string]string) (*model.Pipeline, error) {
	switch lastBuild.Status {
	case model.StatusDeclined,
		model.StatusBlocked:
		return nil, ErrBadRequest{Msg: fmt.Sprintf("cannot restart a pipeline with status %s", lastBuild.Status)}
	}

	forge, err := loader.GetForge(store, repo)
	if err != nil {
		msg := fmt.Sprintf("failure to load forge for repo '%s'", repo.FullName)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	var pipelineFiles []*forge_types.FileMeta

	// fetch the old pipeline config from database
	configs, err := store.ConfigsForPipeline(lastBuild.ID)
	if err != nil {
		msg := fmt.Sprintf("failure to get pipeline config for %s. %s", repo.FullName, err)
		log.Error().Msgf(msg)
		return nil, ErrNotFound{Msg: msg}
	}

	for _, y := range configs {
		pipelineFiles = append(pipelineFiles, &forge_types.FileMeta{Data: y.Data, Name: y.Name})
	}

	// If config extension is active we should refetch the config in case something changed
	if server.Config.Services.ConfigService != nil && server.Config.Services.ConfigService.IsConfigured() {
		currentFileMeta := make([]*forge_types.FileMeta, len(configs))
		for i, cfg := range configs {
			currentFileMeta[i] = &forge_types.FileMeta{Name: cfg.Name, Data: cfg.Data}
		}

		newConfig, useOld, err := server.Config.Services.ConfigService.FetchConfig(ctx, repo, lastBuild, currentFileMeta)
		if err != nil {
			return nil, ErrBadRequest{
				Msg: fmt.Sprintf("On fetching external pipeline config: %s", err),
			}
		}
		if !useOld {
			pipelineFiles = newConfig
		}
	}

	newBuild := createNewOutOfOld(lastBuild)
	newBuild.Parent = lastBuild.ID

	err = store.CreatePipeline(newBuild)
	if err != nil {
		msg := fmt.Sprintf("failure to save pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if len(configs) == 0 {
		newBuild, uerr := UpdateToStatusError(store, *newBuild, errors.New("pipeline definition not found"))
		if uerr != nil {
			log.Debug().Err(uerr).Msg("failure to update pipeline status")
		}
		return newBuild, nil
	}
	if err := persistPipelineConfigs(store, configs, newBuild.ID); err != nil {
		msg := fmt.Sprintf("failure to persist pipeline config for %s.", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	newBuild, pipelineItems, err := createPipelineItems(ctx, forge, store, newBuild, user, repo, pipelineFiles, envs)
	if err != nil {
		if errors.Is(err, &yaml.PipelineParseError{}) {
			return newBuild, nil
		}
		msg := fmt.Sprintf("failure to createBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	newBuild, err = start(ctx, forge, store, newBuild, user, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start pipeline for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return newBuild, nil
}

// TODO: reuse at create.go too
func persistPipelineConfigs(store store.Store, configs []*model.Config, pipelineID int64) error {
	for _, conf := range configs {
		pipelineConfig := &model.PipelineConfig{
			ConfigID:   conf.ID,
			PipelineID: pipelineID,
		}
		err := store.PipelineConfigCreate(pipelineConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func createNewOutOfOld(old *model.Pipeline) *model.Pipeline {
	new := *old
	new.ID = 0
	new.Number = 0
	new.Status = model.StatusPending
	new.Started = 0
	new.Finished = 0
	new.Enqueued = time.Now().UTC().Unix()
	new.Error = ""
	return &new
}

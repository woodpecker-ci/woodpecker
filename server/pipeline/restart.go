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
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func ReStart(ctx context.Context, store store.Store,
	lastBuild *model.Build, user *model.User, repo *model.Repo, envs map[string]string,
) (*model.Build, error) {
	switch lastBuild.Status {
	case model.StatusDeclined,
		model.StatusBlocked:
		return nil, ErrBadRequest{Msg: fmt.Sprintf("cannot restart a build with status %s", lastBuild.Status)}
	}

	var pipelineFiles []*remote.FileMeta

	// fetch the old pipeline config from database
	configs, err := store.ConfigsForBuild(lastBuild.ID)
	if err != nil {
		msg := fmt.Sprintf("failure to get build config for %s. %s", repo.FullName, err)
		log.Error().Msgf(msg)
		return nil, ErrNotFound{Msg: msg}
	}

	for _, y := range configs {
		pipelineFiles = append(pipelineFiles, &remote.FileMeta{Data: y.Data, Name: y.Name})
	}

	// If config extension is active we should refetch the config in case something changed
	if server.Config.Services.ConfigService != nil && server.Config.Services.ConfigService.IsConfigured() {
		currentFileMeta := make([]*remote.FileMeta, len(configs))
		for i, cfg := range configs {
			currentFileMeta[i] = &remote.FileMeta{Name: cfg.Name, Data: cfg.Data}
		}

		newConfig, useOld, err := server.Config.Services.ConfigService.FetchConfig(ctx, repo, lastBuild, currentFileMeta)
		if err != nil {
			return nil, ErrBadRequest{
				Msg: fmt.Sprintf("On fetching external build config: %s", err),
			}
		}
		if !useOld {
			pipelineFiles = newConfig
		}
	}

	newBuild := &model.Build{
		Parent:   lastBuild.ID,
		Deploy:   lastBuild.Deploy,
		Event:    lastBuild.Event,
		Status:   model.StatusPending,
		Enqueued: time.Now().UTC().Unix(),
	}

	err = store.CreateBuild(newBuild)
	if err != nil {
		msg := fmt.Sprintf("failure to save build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if err := persistBuildConfigs(store, configs, newBuild.ID); err != nil {
		msg := fmt.Sprintf("failure to persist build config for %s.", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	build, buildItems, err := CreateBuildItems(ctx, store, newBuild, user, repo, pipelineFiles, envs)
	if err != nil {
		msg := fmt.Sprintf("failure to CreateBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	build, err = Start(ctx, store, build, user, repo, buildItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return build, nil
}

func persistBuildConfigs(store store.Store, configs []*model.Config, buildID int64) error {
	for _, conf := range configs {
		buildConfig := &model.BuildConfig{
			ConfigID: conf.ID,
			BuildID:  buildID,
		}
		err := store.BuildConfigCreate(buildConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

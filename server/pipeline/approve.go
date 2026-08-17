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

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// Approve update the status to pending for a blocked pipeline so it can be executed.
func Approve(ctx context.Context, store store.Store, currentPipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	if currentPipeline.Status != model.StatusBlocked {
		return nil, ErrBadRequest{Msg: fmt.Sprintf("cannot approve a pipeline with status %s", currentPipeline.Status)}
	}

	forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		msg := fmt.Sprintf("failure to load forge for repo '%s'", repo.FullName)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, errors.New(msg)
	}

	// fetch the pipeline file from the database
	configs, err := store.ConfigsForPipeline(currentPipeline.ID)
	if err != nil {
		msg := fmt.Sprintf("failure to get pipeline config for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, ErrNotFound{Msg: msg}
	}
	var yamls []*forge_types.FileMeta
	for _, y := range configs {
		yamls = append(yamls, &forge_types.FileMeta{Data: y.Data, Name: y.Name})
	}

	// Release the gate before building workflows: saveWorkflowsFromPipelineBuilder
	// derives workflow and step state from the pipeline status, so the status
	// must already be pending when the new workflows are persisted.
	currentPipeline.Status = model.StatusPending

	currentPipeline, pipelineItems, parseErr, err := createPipelineItems(ctx, forge, store, currentPipeline, user, repo, yamls, nil, true)
	if handleParseErrors(currentPipeline, parseErr) {
		if err := updatePipelineWithErr(ctx, forge, store, currentPipeline, repo, user, parseErr); err != nil {
			log.Error().Err(err).Msgf("error setting error status of pipeline for %s#%d after approval", repo.FullName, currentPipeline.Number)
		}
		msg := fmt.Sprintf("failure to parse pipeline config for %s", repo.FullName)
		log.Error().Err(parseErr).Msg(msg)
		return nil, errors.New(msg)
	}
	if err != nil {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error persisting new steps for %s#%d after approval", repo.FullName, currentPipeline.Number)
		return nil, err
	}

	if currentPipeline, err = UpdateToStatusPending(store, *currentPipeline, user.Login); err != nil {
		return nil, fmt.Errorf("error updating pipeline. %w", err)
	}

	publishPipeline(ctx, forge, currentPipeline, repo, user)

	currentPipeline, err = start(ctx, forge, store, currentPipeline, user, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start pipeline for %s: %v", repo.FullName, err)
		log.Error().Err(err).Msg(msg)
		return nil, errors.New(msg)
	}

	return currentPipeline, nil
}

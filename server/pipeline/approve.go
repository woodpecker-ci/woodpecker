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

	"github.com/woodpecker-ci/woodpecker/server/model"
	remote_types "github.com/woodpecker-ci/woodpecker/server/remote/types"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Approve update the status to pending for blocked pipeline because of a gated repo
// and start them afterwards
func Approve(ctx context.Context, store store.Store, currentPipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	if currentPipeline.Status != model.StatusBlocked {
		return nil, ErrBadRequest{Msg: fmt.Sprintf("cannot decline a pipeline with status %s", currentPipeline.Status)}
	}

	// fetch the pipeline file from the database
	configs, err := store.ConfigsForPipeline(currentPipeline.ID)
	if err != nil {
		msg := fmt.Sprintf("failure to get pipeline config for %s. %s", repo.FullName, err)
		log.Error().Msg(msg)
		return nil, ErrNotFound{Msg: msg}
	}

	if currentPipeline, err = UpdateToStatusPending(store, *currentPipeline, user.Login); err != nil {
		return nil, fmt.Errorf("error updating pipeline. %s", err)
	}

	var yamls []*remote_types.FileMeta
	for _, y := range configs {
		yamls = append(yamls, &remote_types.FileMeta{Data: y.Data, Name: y.Name})
	}

	currentPipeline, pipelineItems, err := createPipelineItems(ctx, store, currentPipeline, user, repo, yamls, nil)
	if err != nil {
		msg := fmt.Sprintf("failure to createBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, err
	}

	currentPipeline, err = start(ctx, store, currentPipeline, user, repo, pipelineItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start pipeline for %s: %v", repo.FullName, err)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return currentPipeline, nil
}

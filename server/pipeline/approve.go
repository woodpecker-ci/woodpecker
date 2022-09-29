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
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Approve update the status to pending for blocked build because of a gated repo
// and start them afterwards
func Approve(ctx context.Context, store store.Store, pipeline *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	if pipeline.Status != model.StatusBlocked {
		return nil, ErrBadRequest{Msg: fmt.Sprintf("cannot decline a pipeline with status %s", pipeline.Status)}
	}

	// fetch the build file from the database
	configs, err := store.ConfigsForBuild(pipeline.ID)
	if err != nil {
		msg := fmt.Sprintf("failure to get pipeline config for %s. %s", repo.FullName, err)
		log.Error().Msg(msg)
		return nil, ErrNotFound{Msg: msg}
	}

	if pipeline, err = shared.UpdateToStatusPending(store, *pipeline, user.Login); err != nil {
		return nil, fmt.Errorf("error updating pipeline. %s", err)
	}

	var yamls []*remote.FileMeta
	for _, y := range configs {
		yamls = append(yamls, &remote.FileMeta{Data: y.Data, Name: y.Name})
	}

	pipeline, buildItems, err := createPipelineItems(ctx, store, pipeline, user, repo, yamls, nil)
	if err != nil {
		msg := fmt.Sprintf("failure to createBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, err
	}

	pipeline, err = start(ctx, store, pipeline, user, repo, buildItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start pipeline for %s: %v", repo.FullName, err)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return pipeline, nil
}

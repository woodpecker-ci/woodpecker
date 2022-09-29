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
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Decline update the status to declined for blocked build because of a gated repo
func Decline(ctx context.Context, store store.Store, build *model.Pipeline, user *model.User, repo *model.Repo) (*model.Pipeline, error) {
	if build.Status != model.StatusBlocked {
		return nil, fmt.Errorf("cannot decline a build with status %s", build.Status)
	}

	_, err := shared.UpdateToStatusDeclined(store, *build, user.Login)
	if err != nil {
		return nil, fmt.Errorf("error updating build. %s", err)
	}

	if build.Procs, err = store.ProcList(build); err != nil {
		log.Error().Err(err).Msg("can not get proc list from store")
	}
	if build.Procs, err = model.Tree(build.Procs); err != nil {
		log.Error().Err(err).Msg("can not build tree from proc list")
	}

	if err := updateBuildStatus(ctx, build, repo, user); err != nil {
		log.Error().Err(err).Msg("updateBuildStatus")
	}

	if err := publishToTopic(ctx, build, repo); err != nil {
		log.Error().Err(err).Msg("publishToTopic")
	}

	return build, nil
}

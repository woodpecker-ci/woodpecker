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

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// Create a new build and start it
func Create(ctx context.Context, _store store.Store, repo *model.Repo, build *model.Build) (*model.Build, error) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	// if the remote has a refresh token, the current access token
	// may be stale. Therefore, we should refresh prior to dispatching
	// the build.
	if refresher, ok := server.Config.Services.Remote.(remote.Refresher); ok {
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

	// fetch the build file from the remote
	remoteYamlConfigs, err := server.Config.Extensions.Config().FetchConfig(ctx, repoUser, repo, build)
	if err != nil {
		msg := fmt.Sprintf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, build.Ref, repoUser.Login)
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, ErrNotFound{Msg: msg}
	}

	filtered, err := branchFiltered(build, remoteYamlConfigs)
	if err != nil {
		msg := "failure to parse yaml from hook"
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		return nil, ErrBadRequest{Msg: msg}
	}
	if filtered {
		err := ErrFiltered{Msg: "branch does not match restrictions defined in yaml"}
		log.Debug().Str("repo", repo.FullName).Msgf("%v", err)
		return nil, err
	}

	if zeroSteps(build, remoteYamlConfigs) {
		err := ErrFiltered{Msg: "step conditions yield zero runnable steps"}
		log.Debug().Str("repo", repo.FullName).Msgf("%v", err)
		return nil, err
	}

	// update some build fields
	build.RepoID = repo.ID
	build.Verified = true
	build.Status = model.StatusPending

	// TODO(336) extend gated feature with an allow/block List
	if repo.IsGated {
		build.Status = model.StatusBlocked
	}

	err = _store.CreateBuild(build, build.Procs...)
	if err != nil {
		msg := fmt.Sprintf("failure to save build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	// persist the build config for historical correctness, restarts, etc
	for _, remoteYamlConfig := range remoteYamlConfigs {
		_, err := findOrPersistPipelineConfig(_store, build, remoteYamlConfig)
		if err != nil {
			msg := fmt.Sprintf("failure to find or persist pipeline config for %s", repo.FullName)
			log.Error().Err(err).Msg(msg)
			return nil, fmt.Errorf(msg)
		}
	}

	build, buildItems, err := createBuildItems(ctx, _store, build, repoUser, repo, remoteYamlConfigs, nil)
	if err != nil {
		msg := fmt.Sprintf("failure to createBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	if build.Status == model.StatusBlocked {
		if err := publishToTopic(ctx, build, repo); err != nil {
			log.Error().Err(err).Msg("publishToTopic")
		}

		if err := updateBuildStatus(ctx, build, repo, repoUser); err != nil {
			log.Error().Err(err).Msg("updateBuildStatus")
		}

		return build, nil
	}

	build, err = start(ctx, _store, build, repoUser, repo, buildItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return nil, fmt.Errorf(msg)
	}

	return build, nil
}

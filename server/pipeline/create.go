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
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// TODO: create error types instead of return string or interface
func Create(ctx context.Context, _store store.Store, repo *model.Repo, build *model.Build) (string, interface{}, int) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		msg := fmt.Sprintf("failure to find repo owner via id '%d'", repo.UserID)
		log.Error().Err(err).Str("repo", repo.FullName).Msg(msg)
		return msg, nil, http.StatusInternalServerError
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
	configFetcher := shared.NewConfigFetcher(server.Config.Services.Remote, server.Config.Services.ConfigService, repoUser, repo, build)
	remoteYamlConfigs, err := configFetcher.Fetch(ctx)
	if err != nil {
		msg := fmt.Sprintf("cannot find config '%s' in '%s' with user: '%s'", repo.Config, build.Ref, repoUser.Login)
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		return msg, nil, http.StatusNotFound
	}

	filtered, err := branchFiltered(build, remoteYamlConfigs)
	if err != nil {
		msg := "failure to parse yaml from hook"
		log.Debug().Err(err).Str("repo", repo.FullName).Msg(msg)
		return msg, nil, http.StatusBadRequest
	}
	if filtered {
		msg := "ignoring hook: branch does not match restrictions defined in yaml"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		return msg, nil, http.StatusOK
	}

	if zeroSteps(build, remoteYamlConfigs) {
		msg := "ignoring hook: step conditions yield zero runnable steps"
		log.Debug().Str("repo", repo.FullName).Msg(msg)
		return msg, nil, http.StatusOK
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
		return msg, nil, http.StatusInternalServerError
	}

	// persist the build config for historical correctness, restarts, etc
	for _, remoteYamlConfig := range remoteYamlConfigs {
		_, err := findOrPersistPipelineConfig(_store, build, remoteYamlConfig)
		if err != nil {
			msg := fmt.Sprintf("failure to find or persist pipeline config for %s", repo.FullName)
			log.Error().Err(err).Msg(msg)
			return msg, nil, http.StatusInternalServerError
		}
	}

	build, buildItems, err := CreateBuildItems(ctx, _store, build, repoUser, repo, remoteYamlConfigs, nil)
	if err != nil {
		msg := fmt.Sprintf("failure to CreateBuildItems for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return msg, nil, http.StatusInternalServerError
	}

	if build.Status == model.StatusBlocked {
		if err := PublishToTopic(ctx, build, repo); err != nil {
			log.Error().Err(err).Msg("PublishToTopic")
		}

		if err := UpdateBuildStatus(ctx, build, repo, repoUser); err != nil {
			log.Error().Err(err).Msg("UpdateBuildStatus")
		}

		return "", build, http.StatusOK
	}

	build, err = Start(ctx, _store, build, repoUser, repo, buildItems)
	if err != nil {
		msg := fmt.Sprintf("failure to start build for %s", repo.FullName)
		log.Error().Err(err).Msg(msg)
		return msg, nil, http.StatusInternalServerError
	}

	return "", build, http.StatusOK
}

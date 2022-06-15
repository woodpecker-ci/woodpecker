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
	"database/sql"
	"errors"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/shared"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func createBuildItems(ctx context.Context, store store.Store, build *model.Build, user *model.User, repo *model.Repo, yamls []*remote.FileMeta, envs map[string]string) (*model.Build, []*shared.BuildItem, error) {
	netrc, err := server.Config.Services.Remote.Netrc(user, repo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate netrc file")
	}

	// get the previous build so that we can send status change notifications
	last, err := store.GetBuildLastBefore(repo, build.Branch, build.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("Error getting last build before build number '%d'", build.Number)
	}

	secs, err := server.Config.Services.Secrets.SecretListBuild(repo, build)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting secrets for %s#%d", repo.FullName, build.Number)
	}

	regs, err := server.Config.Services.Registries.RegistryList(repo)
	if err != nil {
		log.Error().Err(err).Msgf("Error getting registry credentials for %s#%d", repo.FullName, build.Number)
	}

	if envs == nil {
		envs = map[string]string{}
	}
	if server.Config.Services.Environ != nil {
		globals, _ := server.Config.Services.Environ.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	b := shared.ProcBuilder{
		Repo:  repo,
		Curr:  build,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Envs:  envs,
		Link:  server.Config.Server.Host,
		Yamls: yamls,
	}
	buildItems, err := b.Build()
	if err != nil {
		if _, err := shared.UpdateToStatusError(store, *build, err); err != nil {
			log.Error().Err(err).Msgf("Error setting error status of build for %s#%d", repo.FullName, build.Number)
		}
		return nil, nil, err
	}

	build = shared.SetBuildStepsOnBuild(b.Curr, buildItems)

	return build, buildItems, nil
}

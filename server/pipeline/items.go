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

	"github.com/woodpecker-ci/woodpecker/pipeline"
	"github.com/woodpecker-ci/woodpecker/server"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

func createPipelineItems(c context.Context, store store.Store,
	currentPipeline *model.Pipeline, user *model.User, repo *model.Repo,
	yamls []*forge_types.FileMeta, envs map[string]string,
) (*model.Pipeline, []*pipeline.Item, error) {
	netrc, err := server.Config.Services.Forge.Netrc(user, repo)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate netrc file")
	}

	// get the previous pipeline so that we can send status change notifications
	last, err := store.GetPipelineLastBefore(repo, currentPipeline.Branch, currentPipeline.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("Error getting last pipeline before pipeline number '%d'", currentPipeline.Number)
	}

	secs, err := server.Config.Services.Secrets.SecretListPipeline(repo, currentPipeline, &model.ListOptions{All: true})
	if err != nil {
		log.Error().Err(err).Msgf("Error getting secrets for %s#%d", repo.FullName, currentPipeline.Number)
	}

	regs, err := server.Config.Services.Registries.RegistryList(repo, &model.ListOptions{All: true})
	if err != nil {
		log.Error().Err(err).Msgf("Error getting registry credentials for %s#%d", repo.FullName, currentPipeline.Number)
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

	for k, v := range currentPipeline.AdditionalVariables {
		envs[k] = v
	}

	b := pipeline.StepBuilder{
		Repo:  repo,
		Curr:  currentPipeline,
		Last:  last,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Envs:  envs,
		Link:  server.Config.Server.Host,
		Yamls: yamls,
		Forge: server.Config.Services.Forge,
	}
	pipelineItems, err := b.Build()
	if err != nil {
		currentPipeline, uerr := UpdateToStatusError(store, *currentPipeline, err)
		if uerr != nil {
			log.Error().Err(err).Msgf("Error setting error status of pipeline for %s#%d", repo.FullName, currentPipeline.Number)
		} else {
			updatePipelineStatus(c, currentPipeline, repo, user)
		}
		return currentPipeline, nil, err
	}

	currentPipeline = pipeline.SetPipelineStepsOnPipeline(b.Curr, pipelineItems)

	return currentPipeline, pipelineItems, nil
}

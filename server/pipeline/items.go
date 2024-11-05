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

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/pipeline/stepbuilder"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

func parsePipeline(forge forge.Forge, store store.Store, currentPipeline *model.Pipeline, user *model.User, repo *model.Repo, yamls []*forge_types.FileMeta, envs map[string]string) ([]*stepbuilder.Item, error) {
	netrc, err := forge.Netrc(user, repo)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate netrc file")
	}

	// get the previous pipeline so that we can send status change notifications
	prev, err := store.GetPipelineLastBefore(repo, currentPipeline.Branch, currentPipeline.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error getting last pipeline before pipeline number '%d'", currentPipeline.Number)
	}

	secretService := server.Config.Services.Manager.SecretServiceFromRepo(repo)
	secs, err := secretService.SecretListPipeline(repo, currentPipeline)
	if err != nil {
		log.Error().Err(err).Msgf("error getting secrets for %s#%d", repo.FullName, currentPipeline.Number)
	}

	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	regs, err := registryService.RegistryListPipeline(repo, currentPipeline)
	if err != nil {
		log.Error().Err(err).Msgf("error getting registry credentials for %s#%d", repo.FullName, currentPipeline.Number)
	}

	if envs == nil {
		envs = map[string]string{}
	}

	environmentService := server.Config.Services.Manager.EnvironmentService()
	if environmentService != nil {
		globals, _ := environmentService.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	for k, v := range currentPipeline.AdditionalVariables {
		envs[k] = v
	}

	b := stepbuilder.StepBuilder{
		Repo:  repo,
		Curr:  currentPipeline,
		Prev:  prev,
		Netrc: netrc,
		Secs:  secs,
		Regs:  regs,
		Envs:  envs,
		Host:  server.Config.Server.Host,
		Yamls: yamls,
		Forge: forge,
		ProxyOpts: compiler.ProxyOptions{
			NoProxy:    server.Config.Pipeline.Proxy.No,
			HTTPProxy:  server.Config.Pipeline.Proxy.HTTP,
			HTTPSProxy: server.Config.Pipeline.Proxy.HTTPS,
		},
	}
	return b.Build()
}

func createPipelineItems(c context.Context, forge forge.Forge, store store.Store,
	currentPipeline *model.Pipeline, user *model.User, repo *model.Repo,
	yamls []*forge_types.FileMeta, envs map[string]string,
) (*model.Pipeline, []*stepbuilder.Item, error) {
	pipelineItems, err := parsePipeline(forge, store, currentPipeline, user, repo, yamls, envs)
	if pipeline_errors.HasBlockingErrors(err) {
		currentPipeline, uErr := UpdateToStatusError(store, *currentPipeline, err)
		if uErr != nil {
			log.Error().Err(uErr).Msgf("error setting error status of pipeline for %s#%d", repo.FullName, currentPipeline.Number)
		} else {
			updatePipelineStatus(c, forge, currentPipeline, repo, user)
		}

		return currentPipeline, nil, err
	} else if err != nil {
		currentPipeline.Errors = pipeline_errors.GetPipelineErrors(err)
		err = updatePipelinePending(c, forge, store, currentPipeline, repo, user)
	}

	currentPipeline = setPipelineStepsOnPipeline(currentPipeline, pipelineItems)

	return currentPipeline, pipelineItems, err
}

// setPipelineStepsOnPipeline is the link between pipeline representation in "pipeline package" and server
// to be specific this func currently is used to convert the pipeline.Item list (crafted by StepBuilder.Build()) into
// a pipeline that can be stored in the database by the server.
func setPipelineStepsOnPipeline(pipeline *model.Pipeline, pipelineItems []*stepbuilder.Item) *model.Pipeline {
	var pidSequence int
	for _, item := range pipelineItems {
		if pidSequence < item.Workflow.PID {
			pidSequence = item.Workflow.PID
		}
	}

	// the workflows in the pipeline should be empty as only we do populate them,
	// but if a pipeline was already loaded form database it might contain things, so we just clean it
	pipeline.Workflows = nil
	for _, item := range pipelineItems {
		for _, stage := range item.Config.Stages {
			for _, step := range stage.Steps {
				pidSequence++
				step := &model.Step{
					Name:       step.Name,
					UUID:       step.UUID,
					PipelineID: pipeline.ID,
					PID:        pidSequence,
					PPID:       item.Workflow.PID,
					State:      model.StatusPending,
					Failure:    step.Failure,
					Type:       model.StepType(step.Type),
				}
				if item.Workflow.State == model.StatusSkipped {
					step.State = model.StatusSkipped
				}
				if pipeline.Status == model.StatusBlocked {
					step.State = model.StatusBlocked
				}
				item.Workflow.Children = append(item.Workflow.Children, step)
			}
		}
		if pipeline.Status == model.StatusBlocked {
			item.Workflow.State = model.StatusBlocked
		}
		item.Workflow.PipelineID = pipeline.ID
		pipeline.Workflows = append(pipeline.Workflows, item.Workflow)
	}

	return pipeline
}

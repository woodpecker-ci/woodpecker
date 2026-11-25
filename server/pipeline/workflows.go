// Copyright 2026 Woodpecker Authors
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

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	pipeline_metadata "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// ListWorkflows fetches the config for the given, not persisted, pipeline and
// returns the workflows it would run. Nothing is stored.
func ListWorkflows(ctx context.Context, _forge forge.Forge, _store store.Store, repo *model.Repo, pipeline *model.Pipeline) ([]*model.WorkflowInfo, error) {
	repoUser, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, fmt.Errorf("failure to find repo owner via id '%d': %w", repo.UserID, err)
	}

	forge.Refresh(ctx, _forge, _store, repoUser)

	configService := server.Config.Services.Manager.ConfigServiceFromRepo(repo)
	forgeYamlConfigs, configFetchErr := configService.Fetch(ctx, _forge, repoUser, repo, pipeline, nil, false)
	switch {
	case errors.Is(configFetchErr, &forge_types.ErrConfigNotFound{}):
		return []*model.WorkflowInfo{}, nil
	case configFetchErr != nil && forgeYamlConfigs != nil:
		// unexpected status code from config endpoint - using previous config as fallback, same as Create
		log.Warn().Str("repo", repo.FullName).Err(configFetchErr).Msgf("error while fetching config '%s' in '%s' with user: '%s', will fallback to old config", repo.Config, pipeline.Ref, repoUser.Login)
	case configFetchErr != nil:
		return nil, fmt.Errorf("could not load config from forge: %w", configFetchErr)
	}

	// Substitute and parse like a run would, without linting or compiling.
	envs := make(map[string]string)
	if environmentService := server.Config.Services.Manager.EnvironmentService(); environmentService != nil {
		globals, err := environmentService.EnvironList(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to list global environment for repo %s: %w", repo.FullName, err)
		}
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	serverMetadata := metadata.NewServerMetadata(_forge, repo, pipeline, nil, server.Config.Server.Host)

	yamls := make([]*builder.YamlFile, 0, len(forgeYamlConfigs))
	for _, forgeYaml := range forgeYamlConfigs {
		yamls = append(yamls, &builder.YamlFile{
			Name: forgeYaml.Name,
			Data: forgeYaml.Data,
		})
	}

	b := builder.PipelineBuilder{
		GetWorkflowMetadata: serverMetadata.GetWorkflowMetadata,
		Envs:                envs,
		AdditionalEnvs:      pipeline.AdditionalVariables,
		Yamls:               yamls,
		TrustedClonePlugins: append(repo.NetrcTrustedPlugins, server.Config.Pipeline.TrustedClonePlugins...),
		PrivilegedPlugins:   server.Config.Pipeline.PrivilegedPlugins,
		RepoTrusted: &pipeline_metadata.TrustedConfiguration{
			Network:  repo.Trusted.Network,
			Volumes:  repo.Trusted.Volumes,
			Security: repo.Trusted.Security,
		},
		ListOnly: true,
	}

	items, err := b.Build()
	if pipeline_errors.HasBlockingErrors(err) {
		log.Debug().Str("repo", repo.FullName).Err(err).Msg("failed to list workflows")
		return nil, &ErrBadRequest{Msg: fmt.Sprintf("could not parse pipeline config: %s", err)}
	}

	// matrix workflows produce one item per axis, list each name once
	workflows := make([]*model.WorkflowInfo, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		if seen[item.Workflow.Name] {
			continue
		}
		seen[item.Workflow.Name] = true
		workflows = append(workflows, &model.WorkflowInfo{Name: item.Workflow.Name, DependsOn: item.DependsOn.Names()})
	}
	return workflows, nil
}

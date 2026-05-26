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
	"fmt"
	"maps"

	"github.com/rs/zerolog/log"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	pipeline_metadata "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/compiler"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

func parsePipeline(ctx context.Context, forge forge.Forge, store store.Store, currentPipeline *model.Pipeline, user *model.User, repo *model.Repo, forgeYamls []*forge_types.FileMeta, envs map[string]string) ([]*builder.Item, error) {
	netrc, err := forge.Netrc(user, repo)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate netrc file")
		netrc = &model.Netrc{}
	}

	// get the previous pipeline so that we can send status change notifications
	prev, err := store.GetPipelineLastBefore(repo, currentPipeline.Branch, currentPipeline.ID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Str("repo", repo.FullName).Msgf("error getting last pipeline before pipeline number '%d'", currentPipeline.Number)
	}

	secretService := server.Config.Services.Manager.SecretServiceFromRepo(repo)
	secs, err := secretService.SecretListPipeline(ctx, repo, currentPipeline, netrc)
	if err != nil {
		return nil, fmt.Errorf("error getting secrets for %s#%d: %w", repo.FullName, currentPipeline.Number, err)
	}

	var secrets []compiler.Secret
	for _, sec := range secs {
		var events []pipeline_metadata.Event
		for _, event := range sec.Events {
			events = append(events, pipeline_metadata.Event(event))
		}

		secrets = append(secrets, compiler.Secret{
			Name:           sec.Name,
			Value:          sec.Value,
			AllowedPlugins: sec.Images,
			Events:         events,
		})
	}

	registryService := server.Config.Services.Manager.RegistryServiceFromRepo(repo)
	regs, err := registryService.RegistryListPipeline(ctx, repo, currentPipeline, netrc)
	if err != nil {
		return nil, fmt.Errorf("error getting registry credentials for %s#%d: %w", repo.FullName, currentPipeline.Number, err)
	}

	var registries []compiler.Registry
	for _, reg := range regs {
		registries = append(registries, compiler.Registry{
			Hostname: reg.Address,
			Username: reg.Username,
			Password: reg.Password,
		})
	}

	if envs == nil {
		envs = map[string]string{}
	}

	environmentService := server.Config.Services.Manager.EnvironmentService()
	if environmentService != nil {
		globals, err := environmentService.EnvironList(repo)
		if err != nil {
			return nil, fmt.Errorf("failed to list global environment for repo %s: %w", repo.FullName, err)
		}
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	maps.Copy(envs, currentPipeline.AdditionalVariables)

	serverMetadata := metadata.NewServerMetadata(forge, repo, currentPipeline, prev, server.Config.Server.Host)

	yamls := make([]*builder.YamlFile, 0, len(forgeYamls))
	for _, forgeYaml := range forgeYamls {
		yamls = append(yamls, &builder.YamlFile{
			Name: forgeYaml.Name,
			Data: forgeYaml.Data,
		})
	}

	b := builder.PipelineBuilder{
		GetWorkflowMetadata: serverMetadata.GetWorkflowMetadata,
		Envs:                envs,
		Yamls:               yamls,
		TrustedClonePlugins: append(repo.NetrcTrustedPlugins, server.Config.Pipeline.TrustedClonePlugins...),
		PrivilegedPlugins:   server.Config.Pipeline.PrivilegedPlugins,
		RepoTrusted: &pipeline_metadata.TrustedConfiguration{
			Network:  repo.Trusted.Network,
			Volumes:  repo.Trusted.Volumes,
			Security: repo.Trusted.Security,
		},
		DefaultLabels: server.Config.Pipeline.DefaultWorkflowLabels,
		CompilerOptions: []compiler.Option{
			compiler.WithLocal(false),
			compiler.WithRegistry(registries...),
			compiler.WithSecret(secrets...),
			compiler.WithProxy(compiler.ProxyOptions{
				NoProxy:    server.Config.Pipeline.Proxy.No,
				HTTPProxy:  server.Config.Pipeline.Proxy.HTTP,
				HTTPSProxy: server.Config.Pipeline.Proxy.HTTPS,
			}),
			compiler.WithVolumes(server.Config.Pipeline.Volumes...),
			compiler.WithNetworks(server.Config.Pipeline.Networks...),
			compiler.WithOption(
				compiler.WithNetrc(
					netrc.Login,
					netrc.Password,
					netrc.Machine,
				),
				repo.IsSCMPrivate || server.Config.Pipeline.AuthenticatePublicRepos,
			),
			compiler.WithDefaultClonePlugin(server.Config.Pipeline.DefaultClonePlugin),
			compiler.WithWorkspaceFromURL(compiler.DefaultWorkspaceBase, repo.ForgeURL),
		},
	}

	// TODO: remove with version 4.x
	if server.Config.Pipeline.ForceIgnoreServiceFailure {
		b.CompilerOptions = append(b.CompilerOptions, compiler.WithForceIgnoreServiceFailure())
	}

	return b.Build()
}

// handleParseErrors classifies the error returned by parsePipeline. Blocking
// errors abort the run, so true is returned and the caller decides how to
// report and persist the failure. Non-blocking errors are recorded on the
// pipeline so they surface to the user without stopping the run.
func handleParseErrors(pipeline *model.Pipeline, parseErr error) (blocking bool) {
	if pipeline_errors.HasBlockingErrors(parseErr) {
		return true
	}
	if parseErr != nil {
		pipeline.Errors = pipeline_errors.GetPipelineErrors(parseErr)
	}
	return false
}

// The createPipelineItems parses the pipeline config and persists the resulting
// workflows. It is the shared core of Create, Approve, and Restart.
//
// Returns two errors: parseErr carries pipeline config diagnostics, which
// callers classify with handleParseErrors and report as a blocking failure in
// their own way. The second error, err, signals a hard failure (e.g. persisting
// workflows) that always aborts the run. When the pipeline already has
// persisted workflows (a gated pipeline being approved), setting replaceExisting
// swaps them out for the freshly built ones.
func createPipelineItems(ctx context.Context, forge forge.Forge, store store.Store,
	currentPipeline *model.Pipeline, user *model.User, repo *model.Repo,
	yamls []*forge_types.FileMeta, envs map[string]string, replaceExisting bool,
) (pipeline *model.Pipeline, items []*builder.Item, parseErr, err error) {
	pipelineItems, parseErr := parsePipeline(ctx, forge, store, currentPipeline, user, repo, yamls, envs)
	if pipeline_errors.HasBlockingErrors(parseErr) {
		return currentPipeline, nil, parseErr, nil
	}

	// An empty pipeline (e.g. everything filtered out) has no workflows to
	// persist. Return early so the caller can filter it without us touching
	// the store.
	if len(pipelineItems) == 0 {
		return currentPipeline, pipelineItems, parseErr, nil
	}

	enrichPipelineItemSteps(pipelineItems, repo)
	currentPipeline, err = saveWorkflowsFromPipelineBuilder(store, currentPipeline, pipelineItems, replaceExisting)
	if err != nil {
		return currentPipeline, nil, parseErr, err
	}

	return currentPipeline, pipelineItems, parseErr, nil
}

// enrichPipelineItemSteps stamps server-side fields onto the backend step
// definitions inside each item's compiled config.
//
// TODO(6444): OrgID and WorkflowLabels on backend/types.Step are Kubernetes-specific
// and should be moved to step.BackendOptions so that generic step types carry
// no backend-specific fields.
func enrichPipelineItemSteps(items []*builder.Item, repo *model.Repo) {
	for _, item := range items {
		for stageI := range item.Config.Stages {
			for stepI := range item.Config.Stages[stageI].Steps {
				item.Config.Stages[stageI].Steps[stepI].WorkflowLabels = item.Labels
				item.Config.Stages[stageI].Steps[stepI].OrgID = repo.OrgID
			}
		}
	}
}

// saveWorkflowsFromPipelineBuilder converts the pipeline.Item list crafted by
// PipelineBuilder.Build() into model workflows and persists them.
//
// A freshly created pipeline has no workflows yet, so they are inserted. A
// gated pipeline already persisted its workflows when it was created, so on
// approval the stored workflows must be swapped for the freshly built ones:
// pass replaceExisting to delete the old workflows and steps before inserting.
func saveWorkflowsFromPipelineBuilder(store store.Store, pipeline *model.Pipeline, pipelineItems []*builder.Item, replaceExisting bool) (*model.Pipeline, error) {
	if pipeline.Workflows != nil && !replaceExisting {
		return nil, errors.New("cannot save new workflows from pipeline builder: pipeline already has workflows loaded")
	}

	workflows := workflowsFromPipelineBuilder(pipeline, pipelineItems)

	if replaceExisting {
		if err := store.WorkflowsReplace(pipeline, workflows); err != nil {
			return nil, err
		}
	} else if err := store.WorkflowsCreate(workflows); err != nil {
		return nil, err
	}

	pipeline.Workflows = workflows
	setPipelineItemWorkflowIDs(pipelineItems, pipeline.Workflows)

	return pipeline, nil
}

func workflowsFromPipelineBuilder(pipeline *model.Pipeline, pipelineItems []*builder.Item) []*model.Workflow {
	var pidSequence int
	for _, item := range pipelineItems {
		if pidSequence < item.Workflow.PID {
			pidSequence = item.Workflow.PID
		}
	}

	workflows := make([]*model.Workflow, 0, len(pipelineItems))

	for _, item := range pipelineItems {
		workflow := &model.Workflow{
			ID:         item.Workflow.ID,
			Name:       item.Workflow.Name,
			PID:        item.Workflow.PID,
			PipelineID: pipeline.ID,
			State:      model.StatusPending,
			Environ:    item.Workflow.Environ,
			AxisID:     item.Workflow.AxisID,
		}

		if pipeline.Status == model.StatusBlocked {
			workflow.State = model.StatusBlocked
		}

		// gather all workflow steps through stages as flat list
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

				if pipeline.Status == model.StatusBlocked {
					step.State = model.StatusBlocked
				}
				workflow.Children = append(workflow.Children, step)
			}
		}

		workflows = append(workflows, workflow)
	}

	return workflows
}

func setPipelineItemWorkflowIDs(pipelineItems []*builder.Item, workflows []*model.Workflow) {
	for i, wf := range workflows {
		pipelineItems[i].Workflow.ID = wf.ID
	}
}

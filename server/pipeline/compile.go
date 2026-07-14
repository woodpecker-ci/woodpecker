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
	"fmt"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// CompileWorkflow fully compiles a persisted workflow into the payload an
// agent executes. It is meant to be called at agent fetch time, so the
// embedded credentials (netrc / oauth token, secrets, registries) are fresh at
// the moment the workflow starts instead of frozen at pipeline creation time.
//
// The function is stateless and may be called any number of times for the same
// workflow: a workflow that is re-scheduled (e.g. because its agent stopped
// extending the lease) is simply compiled again. Step identity is taken from
// the persisted step rows, so every compilation reports state against the same
// steps.
func CompileWorkflow(ctx context.Context, _store store.Store, workflowID int64) (*rpc.Workflow, error) {
	workflow, err := _store.WorkflowLoad(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to load workflow %d: %w", workflowID, err)
	}

	currentPipeline, err := _store.GetPipeline(workflow.PipelineID)
	if err != nil {
		return nil, fmt.Errorf("failed to load pipeline %d of workflow %d: %w", workflow.PipelineID, workflowID, err)
	}

	repo, err := _store.GetRepo(currentPipeline.RepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to load repo %d of pipeline %d: %w", currentPipeline.RepoID, currentPipeline.ID, err)
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to load repo owner %d of repo %s: %w", repo.UserID, repo.FullName, err)
	}

	_forge, err := server.Config.Services.Manager.ForgeFromRepo(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to load forge %d of repo %s: %w", repo.ForgeID, repo.FullName, err)
	}

	// make sure the oauth token embedded via netrc is fresh
	forge.Refresh(ctx, _forge, _store, user)

	configs, err := _store.ConfigsForPipeline(currentPipeline.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load configs of pipeline %d: %w", currentPipeline.ID, err)
	}
	yamls := make([]*forge_types.FileMeta, 0, len(configs))
	for _, config := range configs {
		yamls = append(yamls, &forge_types.FileMeta{Name: config.Name, Data: config.Data})
	}

	items, parseErr := parsePipeline(ctx, _forge, _store, currentPipeline, user, repo, yamls, nil)
	if pipeline_errors.HasBlockingErrors(parseErr) {
		return nil, fmt.Errorf("failed to compile workflow %d: %w", workflowID, parseErr)
	}

	item := findItemForWorkflow(items, workflow.PID, workflow.Name)
	if item == nil {
		return nil, fmt.Errorf("compiling pipeline %d no longer produces workflow %s (pid %d)", currentPipeline.ID, workflow.Name, workflow.PID)
	}
	item.Workflow.ID = workflow.ID

	enrichPipelineItemSteps([]*builder.Item{item}, repo)

	if err := applyPersistedStepIdentity(_store, workflow, item); err != nil {
		return nil, err
	}

	return &rpc.Workflow{
		ID:      fmt.Sprint(workflow.ID),
		Config:  item.Config,
		Timeout: repo.Timeout,
	}, nil
}

// findItemForWorkflow selects the compiled item matching the persisted
// workflow. The PID mapping is stable because yaml files are sorted by name
// and the matrix axis order is deterministic.
func findItemForWorkflow(items []*builder.Item, pid int, name string) *builder.Item {
	for _, item := range items {
		if item.Workflow.PID == pid && item.Workflow.Name == name {
			return item
		}
	}
	return nil
}

// applyPersistedStepIdentity stamps the UUIDs of the persisted step rows onto
// the freshly compiled config. Agents report step state by UUID, so every
// compilation of a workflow has to hand out the identities that were persisted
// when the pipeline was created. The persisted steps are ordered by their
// positional id, which matches the flattened stage/step order of the compiled
// config; any structural mismatch is an error.
func applyPersistedStepIdentity(_store store.Store, workflow *model.Workflow, item *builder.Item) error {
	persisted, err := _store.StepListFromWorkflowFind(workflow)
	if err != nil {
		return fmt.Errorf("failed to load steps of workflow %d: %w", workflow.ID, err)
	}

	i := 0
	for _, stage := range item.Config.Stages {
		for _, step := range stage.Steps {
			if i >= len(persisted) {
				return fmt.Errorf("workflow %d compiled to more steps than persisted (%d)", workflow.ID, len(persisted))
			}
			if persisted[i].Name != step.Name {
				return fmt.Errorf("workflow %d step %d changed between compilations: persisted %q, compiled %q", workflow.ID, i, persisted[i].Name, step.Name)
			}
			step.UUID = persisted[i].UUID
			i++
		}
	}
	if i != len(persisted) {
		return fmt.Errorf("workflow %d compiled to %d steps but %d are persisted", workflow.ID, i, len(persisted))
	}

	return nil
}

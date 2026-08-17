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
	"slices"
	"time"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/compile"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// ErrCompilePhaseNotDone is returned by CompilePhaseDone when the merge is not
// this caller's to run: either a compile workflow is still going, or another
// Done call already claimed it. It is not a failure.
var ErrCompilePhaseNotDone = errors.New("compile phase is not this caller's to finish")

// CompilePhaseDone advances a pipeline whose compile phase has just finished.
//
// It is called every time a compile workflow reports Done. Exactly one call
// gets past the compile state swap; that one merges what the compile workflows
// emitted into the pipeline's source configs, persists the result, and appends
// and queues the run workflows.
func CompilePhaseDone(ctx context.Context, _forge forge.Forge, _store store.Store,
	currentPipeline *model.Pipeline, repo *model.Repo, user *model.User,
) error {
	workflows, err := _store.WorkflowGetTree(currentPipeline)
	if err != nil {
		return fmt.Errorf("could not list workflows: %w", err)
	}

	for _, workflow := range workflows {
		if workflow.Phase == model.WorkflowPhaseCompile && workflow.Running() {
			return ErrCompilePhaseNotDone
		}
	}

	// Compile workflows finish on different agents at unpredictable times, so
	// several Done calls can reach this point together. Only the one that moves
	// the state owns the merge.
	claimed, err := _store.PipelineCompileStateCompareAndSwap(
		currentPipeline.ID, model.CompileStateCompiling, model.CompileStateMerging,
	)
	if err != nil {
		return fmt.Errorf("could not claim the compile phase: %w", err)
	}
	if !claimed {
		return ErrCompilePhaseNotDone
	}
	currentPipeline.CompileState = model.CompileStateMerging

	emitted, err := collectCompileResults(workflows)
	if err != nil {
		return failCompilePhase(ctx, _forge, _store, currentPipeline, repo, user, err)
	}

	if err := compile.LintEmitted(emitted); err != nil {
		return failCompilePhase(ctx, _forge, _store, currentPipeline, repo, user, err)
	}

	sourceConfigs, err := _store.SourceConfigsForPipeline(currentPipeline.ID)
	if err != nil {
		return fmt.Errorf("could not load source configs: %w", err)
	}

	source := make([]compile.Config, 0, len(sourceConfigs))
	for _, config := range sourceConfigs {
		source = append(source, compile.Config{Name: config.Name, Data: config.Data})
	}

	merged, err := compile.Merge(source, emitted)
	if err != nil {
		return failCompilePhase(ctx, _forge, _store, currentPipeline, repo, user, err)
	}

	effective, err := persistCompiledConfigs(_store, currentPipeline, repo, merged)
	if err != nil {
		return fmt.Errorf("could not persist compiled configs: %w", err)
	}

	if len(merged) == 0 {
		// The compile phase removed everything. That is a deliberate outcome,
		// not a failure: the pipeline simply has nothing left to run.
		log.Debug().Str("repo", repo.FullName).Msg("compile phase left no configuration to run")
		return finishCompiledPipeline(ctx, _forge, _store, currentPipeline, repo, user)
	}

	return startRunPhase(ctx, _forge, _store, currentPipeline, repo, user, effective, workflows)
}

// collectCompileResults gathers what every compile workflow emitted, in
// workflow order so the merge is deterministic.
func collectCompileResults(workflows []*model.Workflow) ([]compile.Config, error) {
	var emitted []compile.Config

	for _, workflow := range workflows {
		if workflow.Phase != model.WorkflowPhaseCompile {
			continue
		}

		if workflow.Failing() {
			return nil, fmt.Errorf("compile workflow %q failed", workflow.Name)
		}

		switch {
		case workflow.CompileResult == nil:
			return nil, fmt.Errorf("compile workflow %q reported no config response", workflow.Name)
		case workflow.CompileResult.Error != "":
			return nil, fmt.Errorf("compile workflow %q: %s", workflow.Name, workflow.CompileResult.Error)
		}

		for _, config := range workflow.CompileResult.Configs {
			emitted = append(emitted, compile.Config{Name: config.Name, Data: config.Data})
		}
	}

	return emitted, nil
}

// persistCompiledConfigs stores the merged configuration and records it as what
// the pipeline runs, so it can be inspected after the run.
func persistCompiledConfigs(_store store.Store, currentPipeline *model.Pipeline, repo *model.Repo,
	merged []compile.Config,
) ([]*model.Config, error) {
	configs := make([]*model.Config, 0, len(merged))

	for _, config := range merged {
		persisted, err := _store.ConfigPersist(&model.Config{
			RepoID: repo.ID,
			Name:   config.Name,
			Data:   config.Data,
		})
		if err != nil {
			return nil, err
		}
		configs = append(configs, persisted)
	}

	return configs, _store.PipelineConfigsSetEffective(currentPipeline.ID, configs)
}

// startRunPhase builds the run workflows from the merged configuration, appends
// them to the pipeline and queues them.
func startRunPhase(ctx context.Context, _forge forge.Forge, _store store.Store,
	currentPipeline *model.Pipeline, repo *model.Repo, user *model.User,
	merged []*model.Config, existing []*model.Workflow,
) error {
	yamls := make([]*forge_types.FileMeta, 0, len(merged))
	for _, config := range merged {
		yamls = append(yamls, &forge_types.FileMeta{Name: config.Name, Data: config.Data})
	}

	plan, parseErr := parsePipeline(ctx, _forge, _store, currentPipeline, user, repo, yamls, nil)
	if handleParseErrors(currentPipeline, parseErr) {
		return failCompilePhase(ctx, _forge, _store, currentPipeline, repo, user, parseErr)
	}

	// LintEmitted already rejected a nested compile section, so reaching this
	// would mean the two disagree.
	if len(plan.Compile) > 0 {
		return failCompilePhase(ctx, _forge, _store, currentPipeline, repo, user,
			errors.New("the compiled configuration declares a compile section"))
	}

	if len(plan.Run) == 0 {
		log.Debug().Str("repo", repo.FullName).Msg("compile phase produced no workflows to run")
		return finishCompiledPipeline(ctx, _forge, _store, currentPipeline, repo, user)
	}

	// The compile workflows are already persisted and carry forge commit
	// statuses of their own, so the run workflows are appended rather than
	// replacing them: WorkflowsReplace deletes workflows and their steps, which
	// would destroy the history needed to diagnose what a generator did.
	shiftPIDsPast(plan.Run, existing)
	enrichPipelineItemSteps(plan.Run, repo)

	workflows := workflowsFromPipelineBuilder(currentPipeline, plan.Run)
	if err := _store.WorkflowsCreate(workflows); err != nil {
		return fmt.Errorf("could not persist run workflows: %w", err)
	}
	setPipelineItemWorkflowIDs(plan.Run, workflows)
	// A fresh slice, not append onto existing: that would write into the
	// caller's backing array whenever it had spare capacity.
	currentPipeline.Workflows = slices.Concat(existing, workflows)

	currentPipeline.CompileState = model.CompileStateCompiled
	if err := _store.UpdatePipeline(currentPipeline); err != nil {
		return fmt.Errorf("could not update pipeline: %w", err)
	}

	tasks, err := pipelineTasks(repo, currentPipeline, plan.Run)
	if err != nil {
		return fmt.Errorf("could not build tasks for run workflows: %w", err)
	}

	if err := server.Config.Services.Scheduler.StartPipeline(ctx, repo, currentPipeline, tasks); err != nil {
		// The workflows are persisted but will never be picked up, so mark them
		// rather than leaving the pipeline running forever.
		for _, workflow := range workflows {
			workflow.State = model.StatusError
			workflow.Error = "could not queue compiled workflow"
			if updateErr := _store.WorkflowUpdate(workflow); updateErr != nil {
				log.Error().Err(updateErr).Msg("cannot mark unqueued run workflow")
			}
		}
		return fmt.Errorf("could not queue run workflows: %w", err)
	}

	publishPipeline(ctx, _forge, currentPipeline, repo, user)

	return nil
}

// shiftPIDsPast moves a set of items past the positional ids already in use, so
// run workflows continue numbering rather than colliding with the compile
// phase.
func shiftPIDsPast(items []*builder.Item, existing []*model.Workflow) {
	maxPID := 0
	for _, workflow := range existing {
		maxPID = max(maxPID, workflow.PID)
		for _, step := range workflow.Children {
			maxPID = max(maxPID, step.PID)
		}
	}

	for _, item := range items {
		item.Workflow.PID += maxPID
	}
}

// finishCompiledPipeline closes a pipeline whose compile phase left nothing to
// run.
func finishCompiledPipeline(ctx context.Context, _forge forge.Forge, _store store.Store,
	currentPipeline *model.Pipeline, repo *model.Repo, user *model.User,
) error {
	currentPipeline.CompileState = model.CompileStateCompiled
	if err := _store.UpdatePipeline(currentPipeline); err != nil {
		return err
	}

	workflows, err := _store.WorkflowGetTree(currentPipeline)
	if err != nil {
		return err
	}
	currentPipeline.Workflows = workflows

	finished, err := UpdateStatusToDone(_store, *currentPipeline, PipelineStatus(workflows), time.Now().Unix())
	if err != nil {
		return err
	}
	*currentPipeline = *finished

	publishPipeline(ctx, _forge, currentPipeline, repo, user)

	return nil
}

// failCompilePhase reports a compile phase that produced something unusable.
func failCompilePhase(ctx context.Context, _forge forge.Forge, _store store.Store,
	currentPipeline *model.Pipeline, repo *model.Repo, user *model.User, cause error,
) error {
	log.Error().Err(cause).Str("repo", repo.FullName).Msg("compile phase failed")

	currentPipeline.CompileState = model.CompileStateCompiled

	return updatePipelineWithErr(ctx, _forge, _store, currentPipeline, repo, user, cause)
}

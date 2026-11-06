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
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// maxRestartChain bounds how far a chain of partial restarts is followed. Real
// chains are a handful of pipelines long, this only guards against a loop.
// Keep in sync with the web UI (useReusedWorkflows.ts).
const maxRestartChain = 50

// ranWorkflow is a workflow together with the number of the pipeline it ran in.
type ranWorkflow struct {
	*model.Workflow
	PipelineNumber int64
}

// effectiveWorkflows returns the workflows that make up the outcome of a
// pipeline. For a regular pipeline these are its own workflows. A partial
// restart only contains the workflows it reran, so the ones it took over are
// collected from its ancestors, up to the first pipeline that ran in full.
func effectiveWorkflows(_store store.Store, repo *model.Repo, pipeline *model.Pipeline) ([]ranWorkflow, error) {
	var effective []ranWorkflow
	covered := make(map[string]struct{})

	current := pipeline
	for depth := 0; ; depth++ {
		if depth == maxRestartChain {
			log.Warn().Msgf("stopped following the parents of partial restart %s#%d after %d pipelines", repo.FullName, pipeline.Number, maxRestartChain)
			break
		}
		workflows, err := _store.WorkflowGetTree(current)
		if err != nil {
			return nil, err
		}
		for _, wf := range workflows {
			key := workflowKey(wf.Name, wf.Environ)
			if _, ok := covered[key]; ok {
				continue
			}
			covered[key] = struct{}{}
			effective = append(effective, ranWorkflow{Workflow: wf, PipelineNumber: current.Number})
		}

		if !current.PartialRestart || current.Parent == 0 {
			break
		}
		parent := current.Parent
		current, err = _store.GetPipelineNumber(repo, parent)
		if errors.Is(err, types.ErrRecordNotExist) {
			// The parent got deleted. Its workflows are gone with it, so they
			// count as never run.
			log.Warn().Msgf("parent pipeline %s#%d of a partial restart does not exist anymore", repo.FullName, parent)
			break
		}
		if err != nil {
			return nil, fmt.Errorf("get parent pipeline #%d: %w", parent, err)
		}
	}

	return effective, nil
}

// workflowKey identifies a workflow, or one variant of a matrix workflow,
// across the pipelines of a restart chain. Maps print with sorted keys, and a
// nil axis prints like an empty one.
func workflowKey(name string, environ map[string]string) string {
	return fmt.Sprintf("%s%v", name, environ)
}

// workflowSelection holds the workflows of a pipeline that a partial restart
// reruns, see RestartFailed.
type workflowSelection struct {
	// old are the effective workflows of the pipeline the restart is made from.
	old []ranWorkflow
	// oldKeys are the keys of the old workflows.
	oldKeys map[string]struct{}
	// failed are the keys of the workflows that did not pass: the ones that
	// failed, and the skipped ones, as canceling a pipeline skips all
	// workflows that did not start yet. Matrix variants are matched one by
	// one, so only the variants that did not pass are rerun.
	failed map[string]struct{}
	// dependents are the names of the workflows that depend on a failed one,
	// directly or transitively. All of their variants are rerun, as depends_on
	// refers to a workflow by name.
	dependents map[string]struct{}
	// depStatus is filled by filterItems: per rerun workflow name, the state
	// of the dependencies that are not rerun, keyed by their old workflow ID.
	depStatus map[string]map[string]model.StatusValue
}

// selectFailedWorkflows returns the workflows that did not pass and the
// workflows that depend on them, or nil if there is nothing to restart: a
// pipeline needs a failed workflow, or to be canceled, for that.
func selectFailedWorkflows(workflows []ranWorkflow, pipelineStatus model.StatusValue) *workflowSelection {
	s := &workflowSelection{
		old:        workflows,
		oldKeys:    make(map[string]struct{}),
		failed:     make(map[string]struct{}),
		dependents: make(map[string]struct{}),
	}

	restartable := pipelineStatus == model.StatusKilled || pipelineStatus == model.StatusCanceled
	selected := make(map[string]struct{})
	for _, wf := range workflows {
		s.oldKeys[workflowKey(wf.Name, wf.Environ)] = struct{}{}
		if wf.Failing() {
			restartable = true
		}
		if wf.Failing() || wf.State == model.StatusSkipped {
			s.failed[workflowKey(wf.Name, wf.Environ)] = struct{}{}
			selected[wf.Name] = struct{}{}
		}
	}
	// A canceled pipeline may still have nothing left to restart, if the
	// cancel came in as the last workflow finished.
	if !restartable || len(s.failed) == 0 {
		return nil
	}

	for changed := true; changed; {
		changed = false
		for _, wf := range workflows {
			if _, ok := s.dependents[wf.Name]; ok {
				continue
			}
			for _, dep := range wf.DependsOn {
				if _, ok := selected[dep]; ok {
					s.dependents[wf.Name] = struct{}{}
					selected[wf.Name] = struct{}{}
					changed = true
					break
				}
			}
		}
	}

	return s
}

// rerun tells whether the workflow with the given name and matrix axis runs
// again. Besides the selected workflows, this includes the ones that did not
// exist when the old pipeline ran, as they never passed.
func (s *workflowSelection) rerun(name string, environ map[string]string) bool {
	if _, ok := s.dependents[name]; ok {
		return true
	}
	key := workflowKey(name, environ)
	if _, ok := s.failed[key]; ok {
		return true
	}
	_, existed := s.oldKeys[key]
	return !existed
}

// reused returns the workflows that are not rerun with the number of the
// pipeline they ran in. A variant of a matrix workflow is named like its forge
// status, "name/axis" with the position of the variant in the matrix, as its
// variants may have run in different pipelines. It is based on the old
// dependency graph, see filterItems.
func (s *workflowSelection) reused() map[string]int64 {
	reused := make(map[string]int64)
	for _, wf := range s.old {
		if s.rerun(wf.Name, wf.Environ) {
			continue
		}
		name := wf.Name
		if wf.AxisID > 0 {
			name = fmt.Sprintf("%s/%d", wf.Name, wf.AxisID)
		}
		reused[name] = wf.PipelineNumber
	}
	return reused
}

// filterItems keeps the items of the workflows that are rerun. The state the
// workflows that are not rerun finished with is recorded in depStatus for the
// kept items depending on them, so that the rerun is gated on them just like
// the original run was. The queue only links tasks of the kept items, so the
// dependencies themselves can stay as they are.
func (s *workflowSelection) filterItems(items []*builder.Item) []*builder.Item {
	// The dependents were found in the dependency graph the old pipeline ran
	// with. A config extension may serve a different config on restart, so
	// also follow the graph of the config that runs now.
	for changed := true; changed; {
		changed = false
		for _, item := range items {
			if _, ok := s.dependents[item.Workflow.Name]; ok {
				continue
			}
			for _, dep := range item.DependsOn {
				if s.rerunName(dep.Name, items) {
					s.dependents[item.Workflow.Name] = struct{}{}
					changed = true
					break
				}
			}
		}
	}

	kept := make([]*builder.Item, 0, len(items))
	for _, item := range items {
		if s.rerun(item.Workflow.Name, item.Workflow.Environ) {
			kept = append(kept, item)
		}
	}

	s.depStatus = make(map[string]map[string]model.StatusValue)
	for _, item := range kept {
		for _, dep := range item.DependsOn {
			// The variants of the dependency that are not rerun finished in
			// an earlier pipeline.
			for _, wf := range s.old {
				if wf.Name != dep.Name || s.rerun(wf.Name, wf.Environ) {
					continue
				}
				if s.depStatus[item.Workflow.Name] == nil {
					s.depStatus[item.Workflow.Name] = make(map[string]model.StatusValue)
				}
				s.depStatus[item.Workflow.Name][fmt.Sprint(wf.ID)] = wf.State
			}
		}
	}

	return kept
}

// rerunName tells whether any variant of the named workflow is rerun.
func (s *workflowSelection) rerunName(name string, items []*builder.Item) bool {
	for _, item := range items {
		if item.Workflow.Name == name && s.rerun(item.Workflow.Name, item.Workflow.Environ) {
			return true
		}
	}
	return false
}

// Copyright 2025 Woodpecker Authors
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

package builder

import (
	"fmt"
	"path/filepath"
	"strings"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
)

func SanitizePath(path string) string {
	path = filepath.Base(path)
	path = strings.TrimSuffix(path, ".yml")
	path = strings.TrimSuffix(path, ".yaml")
	path = strings.TrimPrefix(path, ".")
	return path
}

// filterMissingDependencies drops items with missing required deps and
// drops missing optional deps from items that survive. Loops until stable
// so a transitive removal doesn't kill an optional consumer.
func filterMissingDependencies(items []*Item) []*Item {
	for {
		kept := make([]*Item, 0, len(items))
		changed := false
		for _, item := range items {
			var resolved constraint.DependsOn
			missingRequired := false
			for _, dep := range item.DependsOn {
				if ContainsItemWithName(dep.Name, items) {
					resolved = append(resolved, dep)
					continue
				}
				if dep.Optional {
					changed = true
					continue
				}
				missingRequired = true
				break
			}
			if missingRequired {
				changed = true
				continue
			}
			item.DependsOn = resolved
			kept = append(kept, item)
		}
		items = kept
		if !changed {
			break
		}
	}

	// surviving deps are all present; flag is no longer relevant
	for _, item := range items {
		for i := range item.DependsOn {
			item.DependsOn[i].Optional = false
		}
	}
	return items
}

func ContainsItemWithName(name string, items []*Item) bool {
	for _, item := range items {
		if name == item.Workflow.Name {
			return true
		}
	}
	return false
}

// validateWorkflowDependencies checks the cross-workflow step dependencies
// (Step.WaitFor, produced by the compiler from 'workflow:' entries in a
// step's depends_on) against the full set of built workflows — the only
// place all workflows of a pipeline are visible at once. It rejects
// self-references, missing required targets (a target absent from items is
// either an unknown name or a workflow filtered out by its when conditions),
// missing target steps, and cycles; missing optional targets are dropped.
func validateWorkflowDependencies(items []*Item) error {
	for _, item := range items {
		for _, stage := range item.Config.Stages {
			for _, step := range stage.Steps {
				if len(step.WaitFor) == 0 {
					continue
				}
				kept := step.WaitFor[:0]
				for _, dep := range step.WaitFor {
					if dep.Workflow == item.Workflow.Name {
						return &pipeline_errors.PipelineError{
							Message: fmt.Sprintf(
								"step '%s' in workflow '%s' depends on its own workflow; use a plain step dependency instead",
								step.Name, item.Workflow.Name),
							Type: pipeline_errors.PipelineErrorTypeCompiler,
						}
					}
					if !ContainsItemWithName(dep.Workflow, items) {
						if dep.Optional {
							continue
						}
						return &pipeline_errors.PipelineError{
							Message: fmt.Sprintf(
								"step '%s' in workflow '%s' depends on unknown workflow '%s' (it does not exist or was filtered out by its conditions)",
								step.Name, item.Workflow.Name, dep.Workflow),
							Type: pipeline_errors.PipelineErrorTypeCompiler,
						}
					}
					// the wait fans out to every matrix item with that
					// workflow name, so the step must exist in all of them
					if dep.Step != "" && !allItemsContainStep(dep.Step, dep.Workflow, items) {
						if dep.Optional {
							continue
						}
						return &pipeline_errors.PipelineError{
							Message: fmt.Sprintf(
								"step '%s' in workflow '%s' depends on unknown step '%s' of workflow '%s'",
								step.Name, item.Workflow.Name, dep.Step, dep.Workflow),
							Type: pipeline_errors.PipelineErrorTypeCompiler,
						}
					}
					kept = append(kept, dep)
				}
				step.WaitFor = kept
			}
		}
	}

	return detectWorkflowDependencyCycles(items)
}

func containsStepWithName(name string, item *Item) bool {
	for _, stage := range item.Config.Stages {
		for _, step := range stage.Steps {
			if step.Name == name {
				return true
			}
		}
	}
	return false
}

func allItemsContainStep(stepName, workflowName string, items []*Item) bool {
	for _, item := range items {
		if item.Workflow.Name == workflowName && !containsStepWithName(stepName, item) {
			return false
		}
	}
	return true
}

// detectWorkflowDependencyCycles rejects cycles on the workflow-collapsed
// dependency graph: nodes are workflow names, edges are workflow-level
// depends_on entries plus step-level cross-workflow dependencies collapsed
// to workflow→workflow. Collapsing is deliberately conservative — a
// step-granular topology that is theoretically schedulable but whose
// workflow-level projection is cyclic is still rejected, because a waiting
// step only unblocks at target (step) completion and such topologies can
// deadlock agents.
func detectWorkflowDependencyCycles(items []*Item) error {
	edges := make(map[string][]string)
	for _, item := range items {
		name := item.Workflow.Name
		for _, dep := range item.DependsOn {
			edges[name] = append(edges[name], dep.Name)
		}
		for _, stage := range item.Config.Stages {
			for _, step := range stage.Steps {
				for _, dep := range step.WaitFor {
					edges[name] = append(edges[name], dep.Workflow)
				}
			}
		}
	}

	const (
		unvisited = iota
		visiting
		done
	)
	state := make(map[string]int)
	var visit func(name string, path []string) error
	visit = func(name string, path []string) error {
		switch state[name] {
		case visiting:
			return &pipeline_errors.PipelineError{
				Message: fmt.Sprintf(
					"cyclic workflow dependency detected: %s -> %s",
					strings.Join(path, " -> "), name),
				Type: pipeline_errors.PipelineErrorTypeCompiler,
			}
		case done:
			return nil
		}
		state[name] = visiting
		for _, dep := range edges[name] {
			if err := visit(dep, append(path, name)); err != nil {
				return err
			}
		}
		state[name] = done
		return nil
	}
	for _, item := range items {
		if err := visit(item.Workflow.Name, nil); err != nil {
			return err
		}
	}
	return nil
}

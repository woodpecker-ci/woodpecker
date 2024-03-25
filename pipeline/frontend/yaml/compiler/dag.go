// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package compiler

import (
	"sort"

	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

type dagCompilerStep struct {
	step      *backend_types.Step
	position  int
	name      string
	group     string
	dependsOn []string
	uses      []string
}

type dagCompiler struct {
	steps    []*dagCompilerStep
	services []*dagCompilerStep
}

func newDAGCompiler(steps []*dagCompilerStep, services []*dagCompilerStep) dagCompiler {
	return dagCompiler{
		steps:    steps,
		services: services,
	}
}

func (c dagCompiler) isDAG() bool {
	for _, v := range c.steps {
		if v.dependsOn != nil {
			return true
		}
	}
	return false
}

func (c dagCompiler) compile() ([]*backend_types.Stage, error) {
	if c.isDAG() {
		return c.compileByDependsOn()
	}
	return c.compileByGroup()
}

func (c dagCompiler) compileByGroup() ([]*backend_types.Stage, error) {
	var stages []*backend_types.Stage

	serviceStage := new(backend_types.Stage)
	for _, srv := range c.services {
		serviceStage.Steps = append(serviceStage.Steps, srv.step)
	}
	stages = append(stages, serviceStage)

	var currentStage *backend_types.Stage
	var currentGroup string

	for _, s := range c.steps {
		// create a new stage if current step is in a new group compared to last one
		if currentStage == nil || currentGroup != s.group || s.group == "" {
			currentGroup = s.group

			currentStage = new(backend_types.Stage)
			stages = append(stages, currentStage)
		}

		// add step to current stage
		currentStage.Steps = append(currentStage.Steps, s.step)
	}

	return stages, nil
}

func (c dagCompiler) compileByDependsOn() ([]*backend_types.Stage, error) {
	stepMap := make(map[string]*dagCompilerStep, len(c.steps))
	for _, s := range c.steps {
		stepMap[s.name] = s
	}
	serviceMap := make(map[string]*dagCompilerStep, len(c.steps))
	for _, s := range c.services {
		stepMap[s.name] = s
	}
	return convertDAGToStages(stepMap, serviceMap)
}

func dfsVisit(steps map[string]*dagCompilerStep, name string, visited map[string]struct{}, path []string) error {
	if _, ok := visited[name]; ok {
		return &ErrStepDependencyCycle{path: path}
	}

	visited[name] = struct{}{}
	path = append(path, name)

	for _, dep := range steps[name].dependsOn {
		if err := dfsVisit(steps, dep, visited, path); err != nil {
			return err
		}
	}

	delete(visited, name)

	return nil
}

func convertDAGToStages(steps map[string]*dagCompilerStep, services map[string]*dagCompilerStep) ([]*backend_types.Stage, error) {
	addedSteps := make(map[string]struct{})
	addedServices := make(map[string]struct{})
	stages := make([]*backend_types.Stage, 0)

	// TODO remove in next major
	if !hasServiceUses(steps, services) {
		stage := new(backend_types.Stage)
		for _, srv := range services {
			stage.Steps = append(stage.Steps, srv.step)
		}
		stages = append(stages, stage)
	}

	for name, step := range steps {
		// check if all depends_on are valid
		for _, dep := range step.dependsOn {
			if _, ok := steps[dep]; !ok {
				return nil, &ErrStepMissingDependency{name: name, dep: dep}
			}
		}

		// check if all uses are valid
		for _, dep := range step.uses {
			if _, ok := services[dep]; !ok {
				return nil, &ErrStepMissingDependency{name: name, dep: dep}
			}
		}

		// check if there are cycles
		visited := make(map[string]struct{})
		if err := dfsVisit(steps, name, visited, []string{}); err != nil {
			return nil, err
		}
	}

	for len(steps) > 0 {
		stage := new(backend_types.Stage)

		var stepsToAdd []*dagCompilerStep
		for name, step := range steps {
			if allDependenciesSatisfied(step, addedSteps) {
				stepsToAdd = append(stepsToAdd, step)
				delete(steps, name)
			}
		}

		var servicesToAdd []*dagCompilerStep
		for _, step := range stepsToAdd {
			for _, u := range step.uses {
				if _, has := addedServices[u]; !has {
					servicesToAdd = append(servicesToAdd, services[u])
					addedServices[u] = struct{}{}
				}
			}
		}
		sort.Slice(servicesToAdd, func(i, j int) bool {
			return servicesToAdd[i].position < servicesToAdd[j].position
		})
		var serviceSteps []*backend_types.Step
		for _, step := range servicesToAdd {
			serviceSteps = append(serviceSteps, step.step)
		}
		if len(stages) == 0 {
			stages = append(stages, &backend_types.Stage{Steps: serviceSteps})
		} else {
			stages[len(stages)-1].Steps = append(stages[len(stages)-1].Steps, serviceSteps...)
		}

		// as steps are from a map that has no deterministic order,
		// we sort the steps by original config position to make the order similar between pipelines
		sort.Slice(stepsToAdd, func(i, j int) bool {
			return stepsToAdd[i].position < stepsToAdd[j].position
		})

		for _, step := range stepsToAdd {
			stage.Steps = append(stage.Steps, step.step)
		}

		for _, step := range stepsToAdd {
			addedSteps[step.name] = struct{}{}
		}

		stages = append(stages, stage)
	}

	return stages, nil
}

func hasServiceUses(steps map[string]*dagCompilerStep, services map[string]*dagCompilerStep) bool {
	for _, step := range steps {
		if len(step.uses) > 0 {
			return true
		}
	}
	for _, step := range services {
		if len(step.uses) > 0 {
			return true
		}
	}
	return false
}

func allDependenciesSatisfied(step *dagCompilerStep, addedSteps map[string]struct{}) bool {
	for _, childName := range step.dependsOn {
		_, ok := addedSteps[childName]
		if !ok {
			return false
		}
	}
	return true
}

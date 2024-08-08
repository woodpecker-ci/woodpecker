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
	needs     []string
}

type dagCompiler struct {
	steps    []*dagCompilerStep
	services []*dagCompilerStep
}

func newDAGCompiler(steps, services []*dagCompilerStep) dagCompiler {
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
	for _, v := range c.services {
		if v.dependsOn != nil {
			return true
		}
	}
	return c.hasNeeds()
}

func (c dagCompiler) hasNeeds() bool {
	for _, v := range c.steps {
		if v.needs != nil {
			return true
		}
	}
	for _, v := range c.services {
		if v.needs != nil {
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
	stages := make([]*backend_types.Stage, 0, len(c.steps))

	if len(c.services) > 0 {
		servicesStage := new(backend_types.Stage)
		for _, s := range c.services {
			servicesStage.Steps = append(servicesStage.Steps, s.step)
		}
		stages = append(stages, servicesStage)
	}

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
	servicesMap := make(map[string]*dagCompilerStep, len(c.services))
	// if no needs -> call with empty services map
	if c.hasNeeds() {
		for _, s := range c.services {
			servicesMap[s.name] = s
		}
	}
	stages, err := convertDAGToStages(stepMap, servicesMap)
	if err != nil {
		return nil, err
	}
	if !c.hasNeeds() && len(c.services) > 0 {
		// add services before steps
		stage := new(backend_types.Stage)
		for _, s := range c.services {
			stage.Steps = append(stage.Steps, s.step)
		}
		stages = append([]*backend_types.Stage{stage}, stages...)
	}
	return stages, nil
}

func dfsVisit(steps, services map[string]*dagCompilerStep, name string, visited, visitedServices map[string]struct{}, path []string, isService bool) error {
	if _, ok := visited[name]; ok && !isService {
		return &ErrStepDependencyCycle{path: path}
	}
	if _, ok := visitedServices[name]; ok && isService {
		return &ErrStepDependencyCycle{path: path}
	}

	var step *dagCompilerStep
	if isService {
		visitedServices[name] = struct{}{}
		step = services[name]
	} else {
		visited[name] = struct{}{}
		step = steps[name]
	}

	path = append(path, name)

	for _, dep := range step.dependsOn {
		if err := dfsVisit(steps, services, dep, visited, visitedServices, path, false); err != nil {
			return err
		}
	}

	for _, dep := range step.needs {
		if err := dfsVisit(steps, services, dep, visited, visitedServices, path, true); err != nil {
			return err
		}
	}

	delete(visited, name)

	return nil
}

func convertDAGToStages(steps, services map[string]*dagCompilerStep) ([]*backend_types.Stage, error) {
	for name, step := range steps {
		// check if all depends_on are valid
		for _, dep := range step.dependsOn {
			if _, ok := steps[dep]; !ok {
				return nil, &ErrStepMissingDependency{name: name, dep: dep}
			}
		}
		for _, dep := range step.needs {
			if _, ok := services[dep]; !ok {
				return nil, &ErrStepMissingDependency{name: name, dep: dep}
			}
		}

		// check if there are cycles
		visited := make(map[string]struct{})
		visitedServices := make(map[string]struct{})
		if err := dfsVisit(steps, services, name, visited, visitedServices, []string{}, false); err != nil {
			return nil, err
		}
	}

	for name, step := range services {
		// check if all depends_on are valid
		for _, dep := range step.dependsOn {
			if _, ok := steps[dep]; !ok {
				return nil, &ErrStepMissingDependency{name: name, dep: dep}
			}
		}
		for _, dep := range step.needs {
			if _, ok := services[dep]; !ok {
				return nil, &ErrStepMissingDependency{name: name, dep: dep}
			}
		}

		// check if there are cycles
		visited := make(map[string]struct{})
		visitedServices := make(map[string]struct{})
		if err := dfsVisit(steps, services, name, visited, visitedServices, []string{}, true); err != nil {
			return nil, err
		}
	}

	addedSteps := make(map[string]struct{})
	addedServices := make(map[string]struct{})
	stages := make([]*backend_types.Stage, 0)

	for len(steps) > 0 || len(services) > 0 {
		addedNodesThisLevel := make(map[string]struct{})
		addedServicesThisLevel := make(map[string]struct{})
		stage := new(backend_types.Stage)

		var stepsToAdd []*dagCompilerStep
		for name, step := range steps {
			if allDependenciesSatisfied(step, addedSteps, addedServices) {
				stepsToAdd = append(stepsToAdd, step)
				addedNodesThisLevel[name] = struct{}{}
				delete(steps, name)
			}
		}
		for name, step := range services {
			if allDependenciesSatisfied(step, addedSteps, addedServices) {
				stepsToAdd = append(stepsToAdd, step)
				addedServicesThisLevel[name] = struct{}{}
				delete(services, name)
			}
		}

		// as steps are from a map that has no deterministic order,
		// we sort the steps by original config position to make the order similar between pipelines
		// Services should appear on top of steps.
		sort.Slice(stepsToAdd, func(i, j int) bool {
			return stepsToAdd[i].position < stepsToAdd[j].position
		})

		for i := range stepsToAdd {
			stage.Steps = append(stage.Steps, stepsToAdd[i].step)
		}

		for name := range addedNodesThisLevel {
			addedSteps[name] = struct{}{}
		}

		for name := range addedServicesThisLevel {
			addedServices[name] = struct{}{}
		}

		stages = append(stages, stage)
	}

	return stages, nil
}

func allDependenciesSatisfied(step *dagCompilerStep, addedSteps, addedServices map[string]struct{}) bool {
	for _, childName := range step.dependsOn {
		_, ok := addedSteps[childName]
		if !ok {
			return false
		}
	}
	for _, childName := range step.needs {
		_, ok := addedServices[childName]
		if !ok {
			return false
		}
	}
	return true
}

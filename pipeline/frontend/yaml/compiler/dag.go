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
	"fmt"

	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

type dagCompilerStep struct {
	step      *backend_types.Step
	position  int
	name      string
	group     string
	dependsOn []string
}

type dagCompiler struct {
	dcs    []*dagCompilerStep
	prefix string
}

func newDAGCompiler(dcs []*dagCompilerStep, prefix string) dagCompiler {
	return dagCompiler{
		dcs:    dcs,
		prefix: prefix,
	}
}

type ErrStepMissingDependency struct {
	name,
	dep string
}

func (err *ErrStepMissingDependency) Error() string {
	return fmt.Sprintf("step %s depends on unknown step %s", err.name, err.dep)
}

func (*ErrStepMissingDependency) Is(target error) bool {
	_, ok := target.(*ErrStepMissingDependency) //nolint:errorlint
	return ok
}

type ErrStepDependencyCycle struct {
	path string
}

func (err *ErrStepDependencyCycle) Error() string {
	return fmt.Sprintf("cycle detected: %v", err.path)
}

func (*ErrStepDependencyCycle) Is(target error) bool {
	_, ok := target.(*ErrStepDependencyCycle) //nolint:errorlint
	return ok
}

func (c dagCompiler) isDAG() bool {
	for _, v := range c.dcs {
		if len(v.dependsOn) != 0 {
			return true
		}
	}
	return false
}

func (dsc dagCompiler) compile() ([]*backend_types.Stage, error) {
	if dsc.isDAG() {
		return dsc.compileByDependsOn()
	}

	return dsc.compileByGroup()
}

func (c dagCompiler) compileByGroup() ([]*backend_types.Stage, error) {
	stages := make([]*backend_types.Stage, 0, len(c.dcs))
	var curStage *backend_types.Stage
	var curGroup string

	for _, s := range c.dcs {
		// create a new stage if current step is in a new group compared to last one
		if curStage == nil || curGroup != s.group || s.group == "" {
			curGroup = s.group

			curStage = new(backend_types.Stage)
			curStage.Name = fmt.Sprintf("%s_stage_%v", c.prefix, s.position)
			stages = append(stages, curStage)
		}

		// add step to curStage
		curStage.Steps = append(curStage.Steps, s.step)
	}

	return stages, nil
}

func (c dagCompiler) compileByDependsOn() ([]*backend_types.Stage, error) {
	stepMap := make(map[string]*dagCompilerStep, len(c.dcs))
	for _, s := range c.dcs {
		stepMap[s.name] = s
	}
	return convertDAGToStages(stepMap, c.prefix)
}

func dfsVisit(steps map[string]*dagCompilerStep, name string, visited map[string]struct{}, path []string) error {
	if _, ok := visited[name]; ok {
		return fmt.Errorf("cycle detected: %v", path)
	}

	visited[name] = struct{}{}
	path = append(path, name)

	for _, dep := range steps[name].dependsOn {
		if err := dfsVisit(steps, dep, visited, path); err != nil {
			return err
		}
	}

	return nil
}

func convertDAGToStages(steps map[string]*dagCompilerStep, prefix string) ([]*backend_types.Stage, error) {
	addedSteps := make(map[string]struct{})
	stages := make([]*backend_types.Stage, 0)

	for name, step := range steps {
		// check if all depends_on are valid
		for _, dep := range step.dependsOn {
			if _, ok := steps[dep]; !ok {
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
		addedNodesThisLevel := make(map[string]struct{})
		stage := &backend_types.Stage{
			Name:  fmt.Sprintf("stage_%d", prefix, len(stages)),
			Alias: fmt.Sprintf("stage_%d", prefix, len(stages)),
		}

		for name, step := range steps {
			if allDependenciesSatisfied(step, addedSteps) {
				stage.Steps = append(stage.Steps, step.step)
				addedNodesThisLevel[name] = struct{}{}
				delete(steps, name)
			}
		}

		for name := range addedNodesThisLevel {
			addedSteps[name] = struct{}{}
		}

		stages = append(stages, stage)
	}

	return stages, nil
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

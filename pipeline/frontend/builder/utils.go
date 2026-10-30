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
	"path/filepath"
	"strings"

	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/matrix"
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
	return items
}

// selectYamls returns the file of the named workflow and, transitively, the
// files of the workflows it hard depends on.
func (b *PipelineBuilder) selectYamls(name string) ([]*YamlFile, error) {
	byName := make(map[string]*YamlFile, len(b.Yamls))
	for _, y := range b.Yamls {
		byName[SanitizePath(y.Name)] = y
	}
	if _, ok := byName[name]; !ok {
		return nil, nil
	}

	wanted := map[string]struct{}{name: {}}
	queue := []string{name}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		dependencies, err := b.workflowDependencies(byName[current])
		if err != nil {
			return nil, err
		}
		for _, dep := range dependencies {
			if dep.Optional {
				continue
			}
			if _, ok := byName[dep.Name]; !ok {
				continue
			}
			if _, ok := wanted[dep.Name]; !ok {
				wanted[dep.Name] = struct{}{}
				queue = append(queue, dep.Name)
			}
		}
	}

	kept := make([]*YamlFile, 0, len(wanted))
	for _, y := range b.Yamls {
		if _, ok := wanted[SanitizePath(y.Name)]; ok {
			kept = append(kept, y)
		}
	}
	return kept, nil
}

// workflowDependencies resolves hard and optional dependencies for every matrix
// axis using the same variable substitution as a workflow build.
func (b *PipelineBuilder) workflowDependencies(file *YamlFile) (constraint.DependsOn, error) {
	axes, err := matrix.Parse(file.Data)
	if err != nil {
		return nil, err
	}
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}

	var dependencies constraint.DependsOn
	for _, axis := range axes {
		workflow := &Workflow{Name: SanitizePath(file.Name), Environ: axis}
		environ := b.environmentVariables(b.GetWorkflowMetadata(workflow), axis)
		substituted, err := metadata.EnvVarSubst(string(file.Data), environ)
		if err != nil {
			return nil, err
		}
		parsed, err := yaml.ParseString(substituted)
		if err != nil {
			return nil, &pipeline_errors.PipelineError{Message: err.Error(), Type: pipeline_errors.PipelineErrorTypeCompiler}
		}
		dependencies = append(dependencies, parsed.DependsOn...)
	}
	return dependencies, nil
}

func ContainsItemWithName(name string, items []*Item) bool {
	for _, item := range items {
		if name == item.Workflow.Name {
			return true
		}
	}
	return false
}

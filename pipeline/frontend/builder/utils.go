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

	"go.uber.org/multierr"

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

// resolveDependencies drops optional deps whose target is not part of the
// pipeline, and reports an error for every required dep that cannot be
// resolved.
//
// A required dep naming a workflow that does not exist, or one that was
// filtered out by its own when:, is a configuration mistake. Dropping the
// dependent workflow instead leaves the user with a pipeline that silently does
// less than it says, and takes everything behind it with it. Use
// `optional: true` to tolerate absence.
func resolveDependencies(items []*Item) error {
	var errs error

	for _, item := range items {
		var resolved constraint.DependsOn

		for _, dep := range item.DependsOn {
			if !ContainsItemWithName(dep.Name, items) {
				if !dep.Optional {
					errs = multierr.Append(errs, &pipeline_errors.PipelineError{
						Type: pipeline_errors.PipelineErrorTypeCompiler,
						Message: fmt.Sprintf(
							"workflow %q depends on %q, which does not exist or was filtered out; mark the dependency optional to tolerate its absence",
							item.Workflow.Name, dep.Name,
						),
					})
				}
				continue
			}

			// the target is present, so the flag has no further meaning
			dep.Optional = false
			resolved = append(resolved, dep)
		}

		item.DependsOn = resolved
	}

	return errs
}

func ContainsItemWithName(name string, items []*Item) bool {
	for _, item := range items {
		if name == item.Workflow.Name {
			return true
		}
	}
	return false
}

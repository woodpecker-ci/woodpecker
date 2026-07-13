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

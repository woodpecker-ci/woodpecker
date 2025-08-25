// Copyright 2025 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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

package stepbuilder

import (
	"path/filepath"
	"strings"
)

func SanitizePath(path string) string {
	path = filepath.Base(path)
	path = strings.TrimSuffix(path, ".yml")
	path = strings.TrimSuffix(path, ".yaml")
	path = strings.TrimPrefix(path, ".")
	return path
}

func stepListContainsItemsToRun(items []*Item) bool {
	for i := range items {
		if items[i].Pending {
			return true
		}
	}
	return false
}

func filterItemsWithMissingDependencies(items []*Item) []*Item {
	itemsToRemove := make([]*Item, 0)

	for _, item := range items {
		for _, dep := range item.DependsOn {
			if !ContainsItemWithName(dep, items) {
				itemsToRemove = append(itemsToRemove, item)
			}
		}
	}

	if len(itemsToRemove) > 0 {
		filtered := make([]*Item, 0)
		for _, item := range items {
			if !ContainsItemWithName(item.Workflow.Name, itemsToRemove) {
				filtered = append(filtered, item)
			}
		}
		// Recursive to handle transitive deps
		return filterItemsWithMissingDependencies(filtered)
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

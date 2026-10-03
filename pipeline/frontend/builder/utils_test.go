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

package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
)

func itemNames(items []*Item) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Workflow.Name)
	}
	return names
}

func TestFilterMissingDependenciesDropsUnignoredMissingRequired(t *testing.T) {
	// unchanged default behavior: a `when` filter pruning "build" still drops
	// "deploy", which hard-depends on it, exactly as before this field existed.
	items := []*Item{
		{Workflow: &Workflow{Name: "deploy"}, DependsOn: constraint.DependsOn{{Name: "build"}}},
	}

	got := filterMissingDependencies(items, nil)

	assert.Empty(t, itemNames(got))
}

func TestFilterMissingDependenciesIgnoresSelectionExcludedDependency(t *testing.T) {
	// "build" is missing because it was left out of a workflow selection, not
	// because a `when` filter pruned it, so "deploy" runs without waiting for it.
	items := []*Item{
		{Workflow: &Workflow{Name: "deploy"}, DependsOn: constraint.DependsOn{{Name: "build"}}},
	}

	got := filterMissingDependencies(items, map[string]bool{"build": true})

	assert.Equal(t, []string{"deploy"}, itemNames(got))
	assert.Empty(t, got[0].DependsOn)
}

func TestFilterMissingDependenciesStillDropsWhenFilteredDependencyAlongsideIgnored(t *testing.T) {
	// "build" is genuinely absent (pruned by `when`, not by selection) so
	// "deploy" is still dropped, even though an unrelated name is ignorable.
	items := []*Item{
		{Workflow: &Workflow{Name: "deploy"}, DependsOn: constraint.DependsOn{{Name: "build"}}},
	}

	got := filterMissingDependencies(items, map[string]bool{"other": true})

	assert.Empty(t, itemNames(got))
}

func TestFilterMissingDependenciesIgnoredDependencyDoesNotMaskOtherRequiredDeps(t *testing.T) {
	items := []*Item{
		{Workflow: &Workflow{Name: "deploy"}, DependsOn: constraint.DependsOn{{Name: "build"}, {Name: "test"}}},
		{Workflow: &Workflow{Name: "test"}},
	}

	got := filterMissingDependencies(items, map[string]bool{"build": true})

	assert.ElementsMatch(t, []string{"deploy", "test"}, itemNames(got))
}

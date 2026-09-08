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
	"fmt"
	"slices"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
)

// ValidateWorkflowSelection reports whether a selection can be run against the
// given configs, without building anything. Triggers that store a selection for
// later (cron jobs) use this at save time, so a selection that could never run
// is refused while the user is still looking at the form, rather than failing
// on every tick with nothing but a server log to show for it.
func ValidateWorkflowSelection(configs []*forge_types.FileMeta, selected []string) error {
	_, err := filterConfigsByWorkflows(configs, selected)
	return err
}

// filterConfigsByWorkflows narrows the configs fetched from the forge down to
// the ones the trigger asked for. An empty selection means "run everything",
// which is the behavior of every trigger that cannot select (webhooks).
//
// A selected name matches a config either by its full path as stored on the
// forge (".woodpecker/deploy.yaml") or by the sanitized workflow name shown in
// the UI and used by depends_on ("deploy"). Selecting a name that matches no
// config is an error rather than a silent no-op: running zero workflows because
// of a typo is the worst possible outcome for the person who pressed the
// button.
func filterConfigsByWorkflows(configs []*forge_types.FileMeta, selected []string) ([]*forge_types.FileMeta, error) {
	if len(selected) == 0 {
		return configs, nil
	}

	matched := make([]*forge_types.FileMeta, 0, len(selected))
	var unknown []string

	for _, name := range selected {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		config := findConfigByName(configs, name)
		if config == nil {
			unknown = append(unknown, name)
			continue
		}
		// tolerate the same workflow being selected twice
		if !slices.Contains(matched, config) {
			matched = append(matched, config)
		}
	}

	if len(unknown) > 0 {
		return nil, &ErrBadRequest{Msg: fmt.Sprintf(
			"unknown workflow(s) %s, this repo defines %s",
			quoteList(unknown), quoteList(WorkflowNames(configs)),
		)}
	}

	if len(matched) == 0 {
		return nil, &ErrBadRequest{Msg: "no workflow selected"}
	}

	if err := checkSelectedDependencies(matched); err != nil {
		return nil, err
	}

	return matched, nil
}

// findConfigByName resolves a selected name to a config, accepting either the
// forge path or the sanitized workflow name.
func findConfigByName(configs []*forge_types.FileMeta, name string) *forge_types.FileMeta {
	for _, config := range configs {
		if config.Name == name || builder.SanitizePath(config.Name) == name {
			return config
		}
	}
	return nil
}

// WorkflowNames returns the sanitized workflow name of every config, in the
// order the forge returned them. These are the names a trigger can select.
func WorkflowNames(configs []*forge_types.FileMeta) []string {
	names := make([]string, 0, len(configs))
	for _, config := range configs {
		names = append(names, builder.SanitizePath(config.Name))
	}
	return names
}

// WorkflowInfo describes one selectable workflow.
type WorkflowInfo struct {
	Name string `json:"name"`
	// DependsOn lists the workflows that must be selected alongside this one.
	// Optional dependencies are left out: they are dropped when absent, so they
	// never make a selection invalid.
	DependsOn []string `json:"depends_on,omitempty"`
} //	@name	WorkflowInfo

// WorkflowInfos returns the selectable workflows together with their required
// dependencies, so a caller can offer a selection that is valid by
// construction rather than discovering the constraint through an error.
//
// A config that fails to parse still appears, with no dependencies. Its real
// problem is reported when a pipeline is built from it, with a better message
// than anything this could produce.
func WorkflowInfos(configs []*forge_types.FileMeta) []WorkflowInfo {
	infos := make([]WorkflowInfo, 0, len(configs))
	for _, config := range configs {
		info := WorkflowInfo{Name: builder.SanitizePath(config.Name)}
		if parsed, err := yaml.ParseBytes(config.Data); err == nil {
			info.DependsOn = parsed.DependsOn.RequiredNames()
		}
		infos = append(infos, info)
	}
	return infos
}

// checkSelectedDependencies rejects a selection that leaves a required
// depends_on unsatisfied.
//
// Without this check the builder's filterMissingDependencies silently drops the
// dependent workflow, and a selection of a single dependent workflow collapses
// into an empty pipeline that the caller reports as "filtered" with no
// explanation of what went wrong.
//
// Configs that fail to parse are skipped: the builder reports the parse error
// properly a moment later, and guessing at dependencies here would only
// duplicate that failure with a worse message.
func checkSelectedDependencies(selected []*forge_types.FileMeta) error {
	present := WorkflowNames(selected)

	var missing []string
	for _, config := range selected {
		parsed, err := yaml.ParseBytes(config.Data)
		if err != nil {
			continue
		}
		for _, dep := range parsed.DependsOn.RequiredNames() {
			if !slices.Contains(present, dep) && !slices.Contains(missing, dep) {
				missing = append(missing, dep)
			}
		}
	}

	if len(missing) > 0 {
		return &ErrBadRequest{Msg: fmt.Sprintf(
			"selected workflows depend on %s, which %s not in the selection",
			quoteList(missing), plural(len(missing), "is", "are"),
		)}
	}

	return nil
}

func quoteList(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("%q", name))
	}
	return strings.Join(quoted, ", ")
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}

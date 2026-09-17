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

		candidates := findConfigsByName(configs, name)
		if len(candidates) == 0 {
			unknown = append(unknown, name)
			continue
		}
		if len(candidates) > 1 {
			paths := make([]string, 0, len(candidates))
			for _, candidate := range candidates {
				paths = append(paths, candidate.Name)
			}
			return nil, &ErrBadRequest{Msg: fmt.Sprintf(
				"ambiguous workflow %q matches %s, select it by path", name, quoteList(paths),
			)}
		}
		config := candidates[0]
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

	if err := checkRequiredDependenciesExist(configs, matched); err != nil {
		return nil, err
	}

	return matched, nil
}

// checkRequiredDependenciesExist refuses a selected workflow whose required
// depends_on names a workflow that exists nowhere in the repo. Without a
// selection the builder drops such a workflow on its own, so a typo has
// always meant "this never runs". With a selection active the builder is
// told to tolerate dependencies the selection left out, and it cannot tell a
// left-out name from a misspelled one: the typo would make the workflow run
// standalone instead. Catching it here, against the full config list, keeps
// that distinction honest. Optional dependencies are dropped when absent
// either way, so an unknown optional name is not an error.
func checkRequiredDependenciesExist(all, selected []*forge_types.FileMeta) error {
	defined := make(map[string]bool, len(all))
	for _, config := range all {
		defined[builder.SanitizePath(config.Name)] = true
	}

	for _, config := range selected {
		parsed, err := yaml.ParseBytes(config.Data)
		if err != nil {
			// the builder reports the parse error with a better message
			continue
		}
		for _, dep := range parsed.DependsOn {
			if dep.Optional || defined[dep.Name] {
				continue
			}
			return &ErrBadRequest{Msg: fmt.Sprintf(
				"workflow %q depends on unknown workflow %q, this repo defines %s",
				builder.SanitizePath(config.Name), dep.Name, quoteList(WorkflowNames(all)),
			)}
		}
	}
	return nil
}

// findConfigsByName resolves a selected name to the configs it could mean. A
// forge path names exactly one config; a sanitized workflow name can match
// several when e.g. both a.yml and a.yaml exist, since SanitizePath drops the
// extension. The caller decides what to do with more than one.
func findConfigsByName(configs []*forge_types.FileMeta, name string) []*forge_types.FileMeta {
	for _, config := range configs {
		if config.Name == name {
			return []*forge_types.FileMeta{config}
		}
	}
	var matched []*forge_types.FileMeta
	for _, config := range configs {
		if builder.SanitizePath(config.Name) == name {
			matched = append(matched, config)
		}
	}
	return matched
}

// WorkflowNames returns the name a trigger can select for every config, in the
// order the forge returned them. That is the sanitized workflow name, except
// when two configs sanitize to the same one: those are listed by their full
// path instead, which is the only name that still tells them apart.
func WorkflowNames(configs []*forge_types.FileMeta) []string {
	names := make([]string, 0, len(configs))
	for _, config := range configs {
		names = append(names, selectableName(configs, config))
	}
	return names
}

func selectableName(configs []*forge_types.FileMeta, config *forge_types.FileMeta) string {
	name := builder.SanitizePath(config.Name)
	for _, other := range configs {
		if other != config && builder.SanitizePath(other.Name) == name {
			return config.Name
		}
	}
	return name
}

// WorkflowInfo describes one selectable workflow.
type WorkflowInfo struct {
	Name string `json:"name"`
	// DependsOn lists every workflow named in this one's depends_on, required
	// or optional. A dependency left out of a selection no longer makes the
	// selection invalid (see PipelineBuilder.IgnoreMissingDependencies), so
	// the distinction stops mattering here: this is purely informational, for
	// a caller that wants to show the dependency before the user picks.
	DependsOn []string `json:"depends_on,omitempty"`
} //	@name	WorkflowInfo

// WorkflowInfos returns the selectable workflows together with the
// dependencies each one names, so a caller can show them before the user
// picks rather than the user discovering them by trial and error.
//
// A config that fails to parse still appears, with no dependencies. Its real
// problem is reported when a pipeline is built from it, with a better message
// than anything this could produce.
func WorkflowInfos(configs []*forge_types.FileMeta) []WorkflowInfo {
	infos := make([]WorkflowInfo, 0, len(configs))
	for _, config := range configs {
		info := WorkflowInfo{Name: selectableName(configs, config)}
		if parsed, err := yaml.ParseBytes(config.Data); err == nil {
			info.DependsOn = parsed.DependsOn.Names()
		}
		infos = append(infos, info)
	}
	return infos
}

// namesExcludedBySelection returns the names any of the given yamls' depends_on
// reference that are not among the yamls themselves — i.e. names a workflow
// selection left out. PipelineBuilder.IgnoreMissingDependencies uses this to
// tell "left out on purpose by the trigger's selection" apart from "pruned by
// a when filter for this run", which must still drop its dependents.
//
// The selected argument being empty means every workflow is running, so
// there is nothing to be lenient about: nil is returned and the builder enforces depends_on
// exactly as it always has. A selection is validated by
// filterConfigsByWorkflows/ValidateWorkflowSelection before this ever runs,
// and that validation rejects a selected workflow whose required depends_on
// names nothing in the repo (checkRequiredDependenciesExist), so a name found
// here really was excluded by choice, not misspelled. An unparsable config is
// a separate, pre-existing problem the builder reports on its own a moment
// later.
func namesExcludedBySelection(selected []string, yamls []*forge_types.FileMeta) map[string]bool {
	if len(selected) == 0 {
		return nil
	}

	present := make(map[string]bool, len(yamls))
	for _, y := range yamls {
		present[builder.SanitizePath(y.Name)] = true
	}

	excluded := make(map[string]bool)
	for _, y := range yamls {
		parsed, err := yaml.ParseBytes(y.Data)
		if err != nil {
			continue
		}
		for _, dep := range parsed.DependsOn.Names() {
			if !present[dep] {
				excluded[dep] = true
			}
		}
	}
	return excluded
}

func quoteList(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, fmt.Sprintf("%q", name))
	}
	return strings.Join(quoted, ", ")
}

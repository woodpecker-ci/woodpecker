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

package linter

import (
	"fmt"

	"go.uber.org/multierr"

	"github.com/woodpecker-ci/woodpecker/pipeline/errors"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/linter/schema"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

// A Linter lints a pipeline configuration.
type Linter struct {
	trusted bool
}

// New creates a new Linter with options.
func New(opts ...Option) *Linter {
	linter := new(Linter)
	for _, opt := range opts {
		opt(linter)
	}
	return linter
}

// Lint lints the configuration.
func (l *Linter) Lint(rawConfig string, c *types.Workflow) error {
	var linterErr error

	if len(c.Steps.ContainerList) == 0 {
		linterErr = multierr.Append(linterErr, newLinterError("Invalid or missing steps section", "steps", false))
	}

	if err := l.lintContainers(c.Clone.ContainerList); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintContainers(c.Steps.ContainerList); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintContainers(c.Services.ContainerList); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}

	if err := l.lintSchema(rawConfig); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintDeprecations(c); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintBadHabits(c); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}

	return linterErr
}

func (l *Linter) lintContainers(containers []*types.Container) error {
	var linterErr error

	for _, container := range containers {
		if err := l.lintImage(container); err != nil {
			linterErr = multierr.Append(linterErr, err)
		}
		if !l.trusted {
			if err := l.lintTrusted(container); err != nil {
				linterErr = multierr.Append(linterErr, err)
			}
		}
		if err := l.lintCommands(container); err != nil {
			linterErr = multierr.Append(linterErr, err)
		}
	}

	return linterErr
}

func (l *Linter) lintImage(c *types.Container) error {
	if len(c.Image) == 0 {
		return newLinterError("Invalid or missing image", fmt.Sprintf("steps.%s", c.Name), false)
	}
	return nil
}

func (l *Linter) lintCommands(c *types.Container) error {
	if len(c.Commands) == 0 {
		return nil
	}
	if len(c.Settings) != 0 {
		var keys []string
		for key := range c.Settings {
			keys = append(keys, key)
		}
		return newLinterError(fmt.Sprintf("Cannot configure both commands and custom attributes %v", keys), fmt.Sprintf("steps.%s", c.Name), false)
	}
	return nil
}

func (l *Linter) lintTrusted(c *types.Container) error {
	yamlPath := fmt.Sprintf("steps.%s", c.Name)
	if c.Privileged {
		return newLinterError("Insufficient privileges to use privileged mode", yamlPath, false)
	}
	if c.ShmSize != 0 {
		return newLinterError("Insufficient privileges to override shm_size", yamlPath, false)
	}
	if len(c.DNS) != 0 {
		return newLinterError("Insufficient privileges to use custom dns", yamlPath, false)
	}
	if len(c.DNSSearch) != 0 {
		return newLinterError("Insufficient privileges to use dns_search", yamlPath, false)
	}
	if len(c.Devices) != 0 {
		return newLinterError("Insufficient privileges to use devices", yamlPath, false)
	}
	if len(c.ExtraHosts) != 0 {
		return newLinterError("Insufficient privileges to use extra_hosts", yamlPath, false)
	}
	if len(c.NetworkMode) != 0 {
		return newLinterError("Insufficient privileges to use network_mode", yamlPath, false)
	}
	if len(c.IpcMode) != 0 {
		return newLinterError("Insufficient privileges to use ipc_mode", yamlPath, false)
	}
	if len(c.Sysctls) != 0 {
		return newLinterError("Insufficient privileges to use sysctls", yamlPath, false)
	}
	if c.Networks.Networks != nil && len(c.Networks.Networks) != 0 {
		return newLinterError("Insufficient privileges to use networks", yamlPath, false)
	}
	if c.Volumes.Volumes != nil && len(c.Volumes.Volumes) != 0 {
		return newLinterError("Insufficient privileges to use volumes", yamlPath, false)
	}
	if len(c.Tmpfs) != 0 {
		return newLinterError("Insufficient privileges to use tmpfs", yamlPath, false)
	}
	return nil
}

func (l *Linter) lintSchema(rawConfig string) error {
	var linterErr error
	schemaErrors, err := schema.LintString(rawConfig)
	if err != nil {
		for _, schemaError := range schemaErrors {
			linterErr = multierr.Append(linterErr, newLinterError(
				schemaError.Description(),
				schemaError.Field(),
				true, // TODO: let pipelines fail if the schema is invalid
			))
		}
	}
	return linterErr
}

func (l *Linter) lintDeprecations(workflow *types.Workflow) error {
	if workflow.PipelineDontUseIt.ContainerList != nil {
		return &errors.PipelineError{
			Type:    errors.PipelineErrorTypeDeprecation,
			Message: "Please use 'steps:' instead of deprecated 'pipeline:' list",
			Data: errors.DeprecationErrorData{
				Docs: "https://woodpecker-ci.org/docs/next/migrations#next-200",
			},
			IsWarning: true,
		}
	}

	if workflow.PlatformDontUseIt != "" {
		return &errors.PipelineError{
			Type:    errors.PipelineErrorTypeDeprecation,
			Message: "Please use labels instead of deprecated 'platform' filters",
			Data: errors.DeprecationErrorData{
				Docs: "https://woodpecker-ci.org/docs/next/migrations#next-200",
			},
			IsWarning: true,
		}
	}

	if !workflow.BranchesDontUseIt.IsEmpty() {
		return &errors.PipelineError{
			Type:    errors.PipelineErrorTypeDeprecation,
			Message: "Please use global when instead of deprecated 'branches' filter",
			Data: errors.DeprecationErrorData{
				Docs: "https://woodpecker-ci.org/docs/next/migrations#next-200",
			},
			IsWarning: true,
		}
	}

	return nil
}

func (l *Linter) lintBadHabits(_ *types.Workflow) error {
	// TODO: add bad habit warnings
	return nil
}

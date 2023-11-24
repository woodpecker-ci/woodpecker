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

	"codeberg.org/6543/xyaml"
	"go.uber.org/multierr"

	"go.woodpecker-ci.org/woodpecker/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/linter/schema"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/types"
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

type WorkflowConfig struct {
	// File is the path to the configuration file.
	File string

	// RawConfig is the raw configuration.
	RawConfig string

	// Config is the parsed configuration.
	Workflow *types.Workflow
}

// Lint lints the configuration.
func (l *Linter) Lint(configs []*WorkflowConfig) error {
	var linterErr error

	for _, config := range configs {
		if err := l.lintFile(config); err != nil {
			linterErr = multierr.Append(linterErr, err)
		}
	}

	return linterErr
}

func (l *Linter) lintFile(config *WorkflowConfig) error {
	var linterErr error

	if len(config.Workflow.Steps.ContainerList) == 0 {
		linterErr = multierr.Append(linterErr, newLinterError("Invalid or missing steps section", config.File, "steps", false))
	}

	if err := l.lintContainers(config, "clone"); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintContainers(config, "steps"); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintContainers(config, "services"); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}

	if err := l.lintSchema(config); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintDeprecations(config); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lintBadHabits(config); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}

	return linterErr
}

func (l *Linter) lintContainers(config *WorkflowConfig, area string) error {
	var linterErr error

	var containers []*types.Container

	switch area {
	case "clone":
		containers = config.Workflow.Clone.ContainerList
	case "steps":
		containers = config.Workflow.Steps.ContainerList
	case "services":
		containers = config.Workflow.Services.ContainerList
	}

	for _, container := range containers {
		if err := l.lintImage(config, container, area); err != nil {
			linterErr = multierr.Append(linterErr, err)
		}
		if !l.trusted {
			if err := l.lintTrusted(config, container, area); err != nil {
				linterErr = multierr.Append(linterErr, err)
			}
		}
		if err := l.lintCommands(config, container, area); err != nil {
			linterErr = multierr.Append(linterErr, err)
		}
	}

	return linterErr
}

func (l *Linter) lintImage(config *WorkflowConfig, c *types.Container, area string) error {
	if len(c.Image) == 0 {
		return newLinterError("Invalid or missing image", config.File, fmt.Sprintf("%s.%s", area, c.Name), false)
	}
	return nil
}

func (l *Linter) lintCommands(config *WorkflowConfig, c *types.Container, field string) error {
	if len(c.Commands) == 0 {
		return nil
	}
	if len(c.Settings) != 0 {
		var keys []string
		for key := range c.Settings {
			keys = append(keys, key)
		}
		return newLinterError(fmt.Sprintf("Cannot configure both commands and custom attributes %v", keys), config.File, fmt.Sprintf("%s.%s", field, c.Name), false)
	}
	return nil
}

func (l *Linter) lintTrusted(config *WorkflowConfig, c *types.Container, area string) error {
	yamlPath := fmt.Sprintf("%s.%s", area, c.Name)
	err := ""
	if c.Privileged {
		err = "Insufficient privileges to use privileged mode"
	}
	if c.ShmSize != 0 {
		err = "Insufficient privileges to override shm_size"
	}
	if len(c.DNS) != 0 {
		err = "Insufficient privileges to use custom dns"
	}
	if len(c.DNSSearch) != 0 {
		err = "Insufficient privileges to use dns_search"
	}
	if len(c.Devices) != 0 {
		err = "Insufficient privileges to use devices"
	}
	if len(c.ExtraHosts) != 0 {
		err = "Insufficient privileges to use extra_hosts"
	}
	if len(c.NetworkMode) != 0 {
		err = "Insufficient privileges to use network_mode"
	}
	if len(c.IpcMode) != 0 {
		err = "Insufficient privileges to use ipc_mode"
	}
	if len(c.Sysctls) != 0 {
		err = "Insufficient privileges to use sysctls"
	}
	if c.Networks.Networks != nil && len(c.Networks.Networks) != 0 {
		err = "Insufficient privileges to use networks"
	}
	if c.Volumes.Volumes != nil && len(c.Volumes.Volumes) != 0 {
		err = "Insufficient privileges to use volumes"
	}
	if len(c.Tmpfs) != 0 {
		err = "Insufficient privileges to use tmpfs"
	}

	if len(err) != 0 {
		return newLinterError(err, config.File, yamlPath, false)
	}

	return nil
}

func (l *Linter) lintSchema(config *WorkflowConfig) error {
	var linterErr error
	schemaErrors, err := schema.LintString(config.RawConfig)
	if err != nil {
		for _, schemaError := range schemaErrors {
			linterErr = multierr.Append(linterErr, newLinterError(
				schemaError.Description(),
				config.File,
				schemaError.Field(),
				true, // TODO: let pipelines fail if the schema is invalid
			))
		}
	}
	return linterErr
}

func (l *Linter) lintDeprecations(config *WorkflowConfig) (err error) {
	parsed := new(types.Workflow)
	err = xyaml.Unmarshal([]byte(config.RawConfig), parsed)
	if err != nil {
		return err
	}

	if parsed.PipelineDontUseIt.ContainerList != nil {
		err = multierr.Append(err, &errors.PipelineError{
			Type:    errors.PipelineErrorTypeDeprecation,
			Message: "Please use 'steps:' instead of deprecated 'pipeline:' list",
			Data: errors.DeprecationErrorData{
				File:  config.File,
				Field: "pipeline",
				Docs:  "https://woodpecker-ci.org/docs/next/migrations#next-200",
			},
			IsWarning: true,
		})
	}

	if parsed.PlatformDontUseIt != "" {
		err = multierr.Append(err, &errors.PipelineError{
			Type:    errors.PipelineErrorTypeDeprecation,
			Message: "Please use labels instead of deprecated 'platform' filters",
			Data: errors.DeprecationErrorData{
				File:  config.File,
				Field: "platform",
				Docs:  "https://woodpecker-ci.org/docs/next/migrations#next-200",
			},
			IsWarning: true,
		})
	}

	if parsed.BranchesDontUseIt != nil {
		err = multierr.Append(err, &errors.PipelineError{
			Type:    errors.PipelineErrorTypeDeprecation,
			Message: "Please use global when instead of deprecated 'branches' filter",
			Data: errors.DeprecationErrorData{
				File:  config.File,
				Field: "branches",
				Docs:  "https://woodpecker-ci.org/docs/next/migrations#next-200",
			},
			IsWarning: true,
		})
	}

	for _, step := range parsed.Steps.ContainerList {
		if step.Group != "" {
			err = multierr.Append(err, &errors.PipelineError{
				Type:    errors.PipelineErrorTypeDeprecation,
				Message: "Please use depends_on instead of deprecated 'group' setting",
				Data: errors.DeprecationErrorData{
					File:  config.File,
					Field: "steps." + step.Name + ".group",
					Docs:  "https://woodpecker-ci.org/docs/next/usage/workflow-syntax#depends_on",
				},
				IsWarning: true,
			})
		}
	}

	return err
}

func (l *Linter) lintBadHabits(_ *WorkflowConfig) error {
	// TODO: add bad habit warnings
	return nil
}

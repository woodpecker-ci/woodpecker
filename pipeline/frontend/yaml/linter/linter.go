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

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/errors"
	errorTypes "go.woodpecker-ci.org/woodpecker/v2/pipeline/errors/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/linter/schema"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/utils"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

// A Linter lints a pipeline configuration.
type Linter struct {
	trusted             bool
	privilegedPlugins   *[]string
	trustedClonePlugins *[]string
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

	if err := l.lintCloneSteps(config); err != nil {
		linterErr = multierr.Append(linterErr, err)
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

func (l *Linter) lintCloneSteps(config *WorkflowConfig) error {
	if len(config.Workflow.Clone.ContainerList) == 0 {
		return nil
	}

	trustedClonePlugins := constant.TrustedClonePlugins
	if l.trustedClonePlugins != nil {
		trustedClonePlugins = *l.trustedClonePlugins
	}

	var linterErr error
	for _, container := range config.Workflow.Clone.ContainerList {
		if !utils.MatchImageDynamic(container.Image, trustedClonePlugins...) {
			linterErr = multierr.Append(linterErr,
				newLinterError(
					"Specified clone image does not match allow list, netrc will not be injected",
					config.File, fmt.Sprintf("clone.%s", container.Name), true),
			)
		}
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
		if err := l.lintSettings(config, container, area); err != nil {
			linterErr = multierr.Append(linterErr, err)
		}
		if err := l.lintPrivilegedPlugins(config, container, area); err != nil {
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

func (l *Linter) lintPrivilegedPlugins(config *WorkflowConfig, c *types.Container, area string) error {
	// lint for conflicts of https://github.com/woodpecker-ci/woodpecker/pull/3918
	if utils.MatchImage(c.Image, "plugins/docker", "plugins/gcr", "plugins/ecr", "woodpeckerci/plugin-docker-buildx") {
		msg := fmt.Sprintf("Cannot use once by default privileged plugin '%s', if needed add it too WOODPECKER_PLUGINS_PRIVILEGED", c.Image)
		// check first if user did not add them back
		if l.privilegedPlugins != nil && !utils.MatchImageDynamic(c.Image, *l.privilegedPlugins...) {
			return newLinterError(msg, config.File, fmt.Sprintf("%s.%s", area, c.Name), false)
		} else if l.privilegedPlugins == nil {
			// if linter has no info of current privileged plugins, it's just a warning
			return newLinterError(msg, config.File, fmt.Sprintf("%s.%s", area, c.Name), true)
		}
	}

	return nil
}

func (l *Linter) lintSettings(config *WorkflowConfig, c *types.Container, field string) error {
	if len(c.Settings) == 0 {
		return nil
	}
	if len(c.Commands) != 0 {
		return newLinterError("Cannot configure both commands and settings", config.File, fmt.Sprintf("%s.%s", field, c.Name), false)
	}
	if len(c.Entrypoint) != 0 {
		return newLinterError("Cannot configure both entrypoint and settings", config.File, fmt.Sprintf("%s.%s", field, c.Name), false)
	}
	if len(c.Environment) != 0 {
		return newLinterError("Should not configure both environment and settings", config.File, fmt.Sprintf("%s.%s", field, c.Name), true)
	}
	if len(c.Secrets) != 0 {
		return newLinterError("Should not configure both secrets and settings", config.File, fmt.Sprintf("%s.%s", field, c.Name), true)
	}
	return nil
}

func (l *Linter) lintTrusted(config *WorkflowConfig, c *types.Container, area string) error {
	yamlPath := fmt.Sprintf("%s.%s", area, c.Name)
	errors := []string{}
	if c.Privileged {
		errors = append(errors, "Insufficient privileges to use privileged mode")
	}
	if c.ShmSize != 0 {
		errors = append(errors, "Insufficient privileges to override shm_size")
	}
	if len(c.DNS) != 0 {
		errors = append(errors, "Insufficient privileges to use custom dns")
	}
	if len(c.DNSSearch) != 0 {
		errors = append(errors, "Insufficient privileges to use dns_search")
	}
	if len(c.Devices) != 0 {
		errors = append(errors, "Insufficient privileges to use devices")
	}
	if len(c.ExtraHosts) != 0 {
		errors = append(errors, "Insufficient privileges to use extra_hosts")
	}
	if len(c.NetworkMode) != 0 {
		errors = append(errors, "Insufficient privileges to use network_mode")
	}
	if len(c.Volumes.Volumes) != 0 {
		errors = append(errors, "Insufficient privileges to use volumes")
	}
	if len(c.Tmpfs) != 0 {
		errors = append(errors, "Insufficient privileges to use tmpfs")
	}

	if len(errors) > 0 {
		var err error

		for _, e := range errors {
			err = multierr.Append(err, newLinterError(e, config.File, yamlPath, false))
		}

		return err
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

	return nil
}

func (l *Linter) lintBadHabits(config *WorkflowConfig) (err error) {
	parsed := new(types.Workflow)
	err = xyaml.Unmarshal([]byte(config.RawConfig), parsed)
	if err != nil {
		return err
	}

	rootEventFilters := len(parsed.When.Constraints) > 0
	for _, c := range parsed.When.Constraints {
		if len(c.Event) == 0 {
			rootEventFilters = false
			break
		}
	}
	if !rootEventFilters {
		// root whens do not necessarily have an event filter, check steps
		for _, step := range parsed.Steps.ContainerList {
			var field string
			if len(step.When.Constraints) == 0 {
				field = fmt.Sprintf("steps.%s", step.Name)
			} else {
				stepEventIndex := -1
				for i, c := range step.When.Constraints {
					if len(c.Event) == 0 {
						stepEventIndex = i
						break
					}
				}
				if stepEventIndex > -1 {
					field = fmt.Sprintf("steps.%s.when[%d]", step.Name, stepEventIndex)
				}
			}
			if field != "" {
				err = multierr.Append(err, &errorTypes.PipelineError{
					Type:    errorTypes.PipelineErrorTypeBadHabit,
					Message: "Please set an event filter for all steps or the whole workflow on all items of the when block",
					Data: errors.BadHabitErrorData{
						File:  config.File,
						Field: field,
						Docs:  "https://woodpecker-ci.org/docs/usage/linter#event-filter-for-all-steps",
					},
					IsWarning: true,
				})
			}
		}
	}

	return
}

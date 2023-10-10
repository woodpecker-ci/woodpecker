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
		linterErr = multierr.Append(linterErr, &LinterError{
			Message: "Invalid or missing pipeline section",
			Field:   "pipeline",
		})
	}

	if err := l.lint(c.Clone.ContainerList); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lint(c.Steps.ContainerList); err != nil {
		linterErr = multierr.Append(linterErr, err)
	}
	if err := l.lint(c.Services.ContainerList); err != nil {
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

func (l *Linter) lint(containers []*types.Container) error {
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
		return &LinterError{Message: "Invalid or missing image", Field: fmt.Sprintf("pipeline.%s", c.Name)}
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
		return &LinterError{Message: fmt.Sprintf("Cannot configure both commands and custom attributes %v", keys), Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	return nil
}

func (l *Linter) lintTrusted(c *types.Container) error {
	if c.Privileged {
		return &LinterError{Message: "Insufficient privileges to use privileged mode", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if c.ShmSize != 0 {
		return &LinterError{Message: "Insufficient privileges to override shm_size", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.DNS) != 0 {
		return &LinterError{Message: "Insufficient privileges to use custom dns", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.DNSSearch) != 0 {
		return &LinterError{Message: "Insufficient privileges to use dns_search", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.Devices) != 0 {
		return &LinterError{Message: "Insufficient privileges to use devices", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.ExtraHosts) != 0 {
		return &LinterError{Message: "Insufficient privileges to use extra_hosts", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.NetworkMode) != 0 {
		return &LinterError{Message: "Insufficient privileges to use network_mode", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.IpcMode) != 0 {
		return &LinterError{Message: "Insufficient privileges to use ipc_mode", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.Sysctls) != 0 {
		return &LinterError{Message: "Insufficient privileges to use sysctls", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if c.Networks.Networks != nil && len(c.Networks.Networks) != 0 {
		return &LinterError{Message: "Insufficient privileges to use networks", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if c.Volumes.Volumes != nil && len(c.Volumes.Volumes) != 0 {
		return &LinterError{Message: "Insufficient privileges to use volumes", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	if len(c.Tmpfs) != 0 {
		return &LinterError{Message: "Insufficient privileges to use tmpfs", Field: fmt.Sprintf("pipeline.%s", c.Name)}
	}
	return nil
}

func (l *Linter) lintSchema(rawConfig string) error {
	var linterErr error
	schemaErrors, err := schema.LintString(rawConfig)
	if err != nil {
		for _, schemaError := range schemaErrors {
			linterErr = multierr.Append(linterErr, &LinterError{
				Message: schemaError.Description(),
				Field:   schemaError.Field(),
				Warning: true, // TODO: let pipelines fail if the schema is invalid
			})
		}
	}
	return linterErr
}

func (l *Linter) lintDeprecations(_ *types.Workflow) error {
	// TODO: add deprecation warnings
	return nil
}

func (l *Linter) lintBadHabits(_ *types.Workflow) error {
	// TODO: add bad habit warnings
	return nil
}

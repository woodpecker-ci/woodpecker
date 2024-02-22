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

package compiler

import (
	"fmt"
	"path"

	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	yaml_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/utils"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

const (
	defaultCloneName = "clone"
)

// Registry represents registry credentials
type Registry struct {
	Hostname string
	Username string
	Password string
	Token    string
}

type Secret struct {
	Name           string
	Value          string
	AllowedPlugins []string
	Events         []string
}

func (s *Secret) Available(event string, container *yaml_types.Container) error {
	onlyAllowSecretForPlugins := len(s.AllowedPlugins) > 0
	if onlyAllowSecretForPlugins && !container.IsPlugin() {
		return fmt.Errorf("secret %q only allowed to be used by plugins by step %q", s.Name, container.Name)
	}

	if onlyAllowSecretForPlugins && !utils.MatchImage(container.Image, s.AllowedPlugins...) {
		return fmt.Errorf("secret %q is not allowed to be used with image %q by step %q", s.Name, container.Image, container.Name)
	}

	if !s.Match(event) {
		return fmt.Errorf("secret %q is not allowed to be used with pipeline event %q", s.Name, event)
	}

	return nil
}

// Match returns true if an image and event match the restricted list.
// Note that EventPullClosed are treated as EventPull.
func (s *Secret) Match(event string) bool {
	// if there is no filter set secret matches all webhook events
	if len(s.Events) == 0 {
		return true
	}
	// tread all pull events the same way
	if event == "pull_request_closed" {
		event = "pull_request"
	}
	// one match is enough
	for _, e := range s.Events {
		if e == event {
			return true
		}
	}
	// a filter is set but the webhook did not match it
	return false
}

type secretMap map[string]Secret

type ResourceLimit struct {
	MemSwapLimit int64
	MemLimit     int64
	ShmSize      int64
	CPUQuota     int64
	CPUShares    int64
	CPUSet       string
}

// Compiler compiles the yaml
type Compiler struct {
	local             bool
	escalated         []string
	prefix            string
	volumes           []string
	networks          []string
	env               map[string]string
	cloneEnv          map[string]string
	base              string
	path              string
	metadata          metadata.Metadata
	registries        []Registry
	secrets           secretMap
	cacher            Cacher
	reslimit          ResourceLimit
	defaultCloneImage string
	trustedPipeline   bool
	netrcOnlyTrusted  bool
}

// New creates a new Compiler with options.
func New(opts ...Option) *Compiler {
	compiler := &Compiler{
		env:      map[string]string{},
		cloneEnv: map[string]string{},
		secrets:  map[string]Secret{},
	}
	for _, opt := range opts {
		opt(compiler)
	}
	return compiler
}

// Compile compiles the YAML configuration to the pipeline intermediate
// representation configuration format.
func (c *Compiler) Compile(conf *yaml_types.Workflow) (*backend_types.Config, error) {
	config := new(backend_types.Config)

	if match, err := conf.When.Match(c.metadata, true, c.env); !match && err == nil {
		// This pipeline does not match the configured filter so return an empty config and stop further compilation.
		// An empty pipeline will just be skipped completely.
		return config, nil
	} else if err != nil {
		return nil, err
	}

	// create a default volume
	config.Volumes = append(config.Volumes, &backend_types.Volume{
		Name: fmt.Sprintf("%s_default", c.prefix),
	})

	// create a default network
	config.Networks = append(config.Networks, &backend_types.Network{
		Name: fmt.Sprintf("%s_default", c.prefix),
	})

	// create secrets for mask
	for _, sec := range c.secrets {
		config.Secrets = append(config.Secrets, &backend_types.Secret{
			Name:  sec.Name,
			Value: sec.Value,
		})
	}

	// overrides the default workspace paths when specified
	// in the YAML file.
	if len(conf.Workspace.Base) != 0 {
		c.base = conf.Workspace.Base
	}
	if len(conf.Workspace.Path) != 0 {
		c.path = conf.Workspace.Path
	}

	cloneImage := constant.DefaultCloneImage
	if len(c.defaultCloneImage) > 0 {
		cloneImage = c.defaultCloneImage
	}

	// add default clone step
	if !c.local && len(conf.Clone.ContainerList) == 0 && !conf.SkipClone {
		cloneSettings := map[string]any{"depth": "0"}
		if c.metadata.Curr.Event == metadata.EventTag {
			cloneSettings["tags"] = "true"
		}
		container := &yaml_types.Container{
			Name:     defaultCloneName,
			Image:    cloneImage,
			Settings: cloneSettings,
		}
		for k, v := range c.cloneEnv {
			container.Environment[k] = v
		}
		step, err := c.createProcess(container, backend_types.StepTypeClone)
		if err != nil {
			return nil, err
		}

		stage := new(backend_types.Stage)
		stage.Steps = append(stage.Steps, step)

		config.Stages = append(config.Stages, stage)
	} else if !c.local && !conf.SkipClone {
		for _, container := range conf.Clone.ContainerList {
			if match, err := container.When.Match(c.metadata, false, c.env); !match && err == nil {
				continue
			} else if err != nil {
				return nil, err
			}

			stage := new(backend_types.Stage)

			step, err := c.createProcess(container, backend_types.StepTypeClone)
			if err != nil {
				return nil, err
			}

			// only inject netrc if it's a trusted repo or a trusted plugin
			if !c.netrcOnlyTrusted || c.trustedPipeline || (container.IsPlugin() && container.IsTrustedCloneImage()) {
				for k, v := range c.cloneEnv {
					step.Environment[k] = v
				}
			}

			stage.Steps = append(stage.Steps, step)

			config.Stages = append(config.Stages, stage)
		}
	}

	err := c.setupCache(conf, config)
	if err != nil {
		return nil, err
	}

	// add services steps
	if len(conf.Services.ContainerList) != 0 {
		stage := new(backend_types.Stage)

		for _, container := range conf.Services.ContainerList {
			if match, err := container.When.Match(c.metadata, false, c.env); !match && err == nil {
				continue
			} else if err != nil {
				return nil, err
			}

			step, err := c.createProcess(container, backend_types.StepTypeService)
			if err != nil {
				return nil, err
			}

			stage.Steps = append(stage.Steps, step)
		}
		config.Stages = append(config.Stages, stage)
	}

	// add pipeline steps
	steps := make([]*dagCompilerStep, 0, len(conf.Steps.ContainerList))
	for pos, container := range conf.Steps.ContainerList {
		// Skip if local and should not run local
		if c.local && !container.When.IsLocal() {
			continue
		}

		if match, err := container.When.Match(c.metadata, false, c.env); !match && err == nil {
			continue
		} else if err != nil {
			return nil, err
		}

		stepType := backend_types.StepTypeCommands
		if container.IsPlugin() {
			stepType = backend_types.StepTypePlugin
		}
		step, err := c.createProcess(container, stepType)
		if err != nil {
			return nil, err
		}

		// inject netrc if it's a trusted repo or a trusted clone-plugin
		if c.trustedPipeline || (container.IsPlugin() && container.IsTrustedCloneImage()) {
			for k, v := range c.cloneEnv {
				step.Environment[k] = v
			}
		}

		steps = append(steps, &dagCompilerStep{
			step:      step,
			position:  pos,
			name:      container.Name,
			group:     container.Group,
			dependsOn: container.DependsOn,
		})
	}

	// generate stages out of steps
	stepStages, err := newDAGCompiler(steps).compile()
	if err != nil {
		return nil, err
	}

	config.Stages = append(config.Stages, stepStages...)

	err = c.setupCacheRebuild(conf, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Compiler) setupCache(conf *yaml_types.Workflow, ir *backend_types.Config) error {
	if c.local || len(conf.Cache) == 0 || c.cacher == nil {
		return nil
	}

	container := c.cacher.Restore(path.Join(c.metadata.Repo.Owner, c.metadata.Repo.Name), c.metadata.Curr.Commit.Branch, conf.Cache)
	step, err := c.createProcess(container, backend_types.StepTypeCache)
	if err != nil {
		return err
	}

	stage := new(backend_types.Stage)
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)

	return nil
}

func (c *Compiler) setupCacheRebuild(conf *yaml_types.Workflow, ir *backend_types.Config) error {
	if c.local || len(conf.Cache) == 0 || c.metadata.Curr.Event != metadata.EventPush || c.cacher == nil {
		return nil
	}
	container := c.cacher.Rebuild(path.Join(c.metadata.Repo.Owner, c.metadata.Repo.Name), c.metadata.Curr.Commit.Branch, conf.Cache)

	step, err := c.createProcess(container, backend_types.StepTypeCache)
	if err != nil {
		return err
	}

	stage := new(backend_types.Stage)
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)

	return nil
}

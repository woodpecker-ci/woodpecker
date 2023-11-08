package compiler

import (
	"fmt"
	"path"

	backend_types "go.woodpecker-ci.org/woodpecker/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/metadata"
	yaml_types "go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/types"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/utils"
	"go.woodpecker-ci.org/woodpecker/shared/constant"
)

const (
	defaultCloneName = "clone"

	nameServices = "services"
)

// Registry represents registry credentials
type Registry struct {
	Hostname string
	Username string
	Password string
	Email    string
	Token    string
}

type Secret struct {
	Name       string
	Value      string
	Match      []string
	PluginOnly bool
}

func (s *Secret) Available(container *yaml_types.Container) bool {
	return (len(s.Match) == 0 || utils.MatchImage(container.Image, s.Match...)) && (!s.PluginOnly || container.IsPlugin())
}

type secretMap map[string]Secret

func (sm secretMap) toStringMap() map[string]string {
	m := make(map[string]string, len(sm))
	for k, v := range sm {
		m[k] = v.Value
	}
	return m
}

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
			Mask:  true,
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
		cloneSettings := map[string]interface{}{"depth": "0"}
		if c.metadata.Curr.Event == metadata.EventTag {
			cloneSettings["tags"] = "true"
		}
		container := &yaml_types.Container{
			Name:        defaultCloneName,
			Image:       cloneImage,
			Settings:    cloneSettings,
			Environment: c.cloneEnv,
		}
		name := fmt.Sprintf("%s_clone", c.prefix)
		step := c.createProcess(name, container, backend_types.StepTypeClone)

		stage := new(backend_types.Stage)
		stage.Name = name
		stage.Alias = defaultCloneName
		stage.Steps = append(stage.Steps, step)

		config.Stages = append(config.Stages, stage)
	} else if !c.local && !conf.SkipClone {
		for i, container := range conf.Clone.ContainerList {
			if match, err := container.When.Match(c.metadata, false, c.env); !match && err == nil {
				continue
			} else if err != nil {
				return nil, err
			}

			stage := new(backend_types.Stage)
			stage.Name = fmt.Sprintf("%s_clone_%v", c.prefix, i)
			stage.Alias = container.Name

			name := fmt.Sprintf("%s_clone_%d", c.prefix, i)
			step := c.createProcess(name, container, backend_types.StepTypeClone)

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

	c.setupCache(conf, config)

	// add services steps
	if len(conf.Services.ContainerList) != 0 {
		stage := new(backend_types.Stage)
		stage.Name = fmt.Sprintf("%s_%s", c.prefix, nameServices)
		stage.Alias = nameServices

		for i, container := range conf.Services.ContainerList {
			if match, err := container.When.Match(c.metadata, false, c.env); !match && err == nil {
				continue
			} else if err != nil {
				return nil, err
			}

			name := fmt.Sprintf("%s_%s_%d", c.prefix, nameServices, i)
			step := c.createProcess(name, container, backend_types.StepTypeService)
			stage.Steps = append(stage.Steps, step)
		}
		config.Stages = append(config.Stages, stage)
	}

	// add pipeline steps. 1 pipeline step per stage, at the moment
	var stage *backend_types.Stage
	var group string
	for i, container := range conf.Steps.ContainerList {
		// Skip if local and should not run local
		if c.local && !container.When.IsLocal() {
			continue
		}

		if match, err := container.When.Match(c.metadata, false, c.env); !match && err == nil {
			continue
		} else if err != nil {
			return nil, err
		}

		if stage == nil || group != container.Group || container.Group == "" {
			group = container.Group

			stage = new(backend_types.Stage)
			stage.Name = fmt.Sprintf("%s_stage_%v", c.prefix, i)
			stage.Alias = container.Name
			config.Stages = append(config.Stages, stage)
		}

		name := fmt.Sprintf("%s_step_%d", c.prefix, i)
		stepType := backend_types.StepTypeCommands
		if container.IsPlugin() {
			stepType = backend_types.StepTypePlugin
		}
		step := c.createProcess(name, container, stepType)
		stage.Steps = append(stage.Steps, step)
	}

	c.setupCacheRebuild(conf, config)

	return config, nil
}

func (c *Compiler) setupCache(conf *yaml_types.Workflow, ir *backend_types.Config) {
	if c.local || len(conf.Cache) == 0 || c.cacher == nil {
		return
	}

	container := c.cacher.Restore(path.Join(c.metadata.Repo.Owner, c.metadata.Repo.Name), c.metadata.Curr.Commit.Branch, conf.Cache)
	name := fmt.Sprintf("%s_restore_cache", c.prefix)
	step := c.createProcess(name, container, backend_types.StepTypeCache)

	stage := new(backend_types.Stage)
	stage.Name = name
	stage.Alias = "restore_cache"
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)
}

func (c *Compiler) setupCacheRebuild(conf *yaml_types.Workflow, ir *backend_types.Config) {
	if c.local || len(conf.Cache) == 0 || c.metadata.Curr.Event != metadata.EventPush || c.cacher == nil {
		return
	}
	container := c.cacher.Rebuild(path.Join(c.metadata.Repo.Owner, c.metadata.Repo.Name), c.metadata.Curr.Commit.Branch, conf.Cache)

	name := fmt.Sprintf("%s_rebuild_cache", c.prefix)
	step := c.createProcess(name, container, backend_types.StepTypeCache)

	stage := new(backend_types.Stage)
	stage.Name = name
	stage.Alias = "rebuild_cache"
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)
}

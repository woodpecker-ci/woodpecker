package compiler

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
)

// TODO(bradrydzewski) compiler should handle user-defined volumes from YAML
// TODO(bradrydzewski) compiler should handle user-defined networks from YAML

const (
	windowsPrefix = "windows/"

	defaultCloneImage = "woodpeckerci/plugin-git:latest"
	defaultCloneName  = "clone"

	networkDriverNAT    = "nat"
	networkDriverBridge = "bridge"
)

type Registry struct {
	Hostname string
	Username string
	Password string
	Email    string
	Token    string
}

type Secret struct {
	Name  string
	Value string
	Match []string
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
	metadata          frontend.Metadata
	registries        []Registry
	secrets           map[string]Secret
	cacher            Cacher
	reslimit          ResourceLimit
	defaultCloneImage string
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
func (c *Compiler) Compile(conf *yaml.Config) *backend.Config {
	config := new(backend.Config)

	if !conf.MatchConstraints(c.metadata) {
		// This pipeline does not match the configured filter so return an empty config and stop further compilation.
		// An empty pipeline will just be skipped and wont be shown in the UI as well.
		return config
	}

	// create a default volume
	config.Volumes = append(config.Volumes, &backend.Volume{
		Name:   fmt.Sprintf("%s_default", c.prefix),
		Driver: "local",
	})

	// create a default network
	if strings.HasPrefix(c.metadata.Sys.Arch, windowsPrefix) {
		config.Networks = append(config.Networks, &backend.Network{
			Name:   fmt.Sprintf("%s_default", c.prefix),
			Driver: networkDriverNAT,
		})
	} else {
		config.Networks = append(config.Networks, &backend.Network{
			Name:   fmt.Sprintf("%s_default", c.prefix),
			Driver: networkDriverBridge,
		})
	}

	// create secrets for mask
	for _, sec := range c.secrets {
		config.Secrets = append(config.Secrets, &backend.Secret{
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

	// add default clone step
	if !c.local && len(conf.Clone.Containers) == 0 && !conf.SkipClone {
		cloneImage := defaultCloneImage
		if len(c.defaultCloneImage) > 0 {
			cloneImage = c.defaultCloneImage
		}
		container := &yaml.Container{
			Name:        defaultCloneName,
			Image:       cloneImage,
			Settings:    map[string]interface{}{"depth": "0"},
			Environment: c.cloneEnv,
		}
		name := fmt.Sprintf("%s_clone", c.prefix)
		step := c.createProcess(name, container, defaultCloneName)

		stage := new(backend.Stage)
		stage.Name = name
		stage.Alias = defaultCloneName
		stage.Steps = append(stage.Steps, step)
		log.Debug().Msg("Added step")
	} else if !c.local && !conf.SkipClone {
		for i, container := range conf.Clone.Containers {
			if !container.MatchConstraints(c.metadata) {
				continue
			}
			stage := new(backend.Stage)
			stage.Name = fmt.Sprintf("%s_clone_%v", c.prefix, i)
			stage.Alias = container.Name

			name := fmt.Sprintf("%s_clone_%d", c.prefix, i)
			step := c.createProcess(name, container, defaultCloneName)
			for k, v := range c.cloneEnv {
				step.Environment[k] = v
			}
			stage.Steps = append(stage.Steps, step)

			config.Stages = append(config.Stages, stage)
		}
	}

	c.setupCache(conf, config)

	// add services steps
	if len(conf.Services.Containers) != 0 {
		stage := new(backend.Stage)
		stage.Name = fmt.Sprintf("%s_services", c.prefix)
		stage.Alias = "services"

		for i, container := range conf.Services.Containers {
			if !container.MatchConstraints(c.metadata) {
				continue
			}

			name := fmt.Sprintf("%s_services_%d", c.prefix, i)
			step := c.createProcess(name, container, "services")
			stage.Steps = append(stage.Steps, step)
		}
		config.Stages = append(config.Stages, stage)
	}

	// add pipeline steps. 1 pipeline step per stage, at the moment
	var stage *backend.Stage
	var group string
	for i, container := range conf.Pipeline.Containers {
		// Skip if local and should not run local
		if c.local && !container.Constraints.IsLocal() {
			continue
		}

		if !container.MatchConstraints(c.metadata) {
			continue
		}

		if stage == nil || group != container.Group || container.Group == "" {
			group = container.Group

			stage = new(backend.Stage)
			stage.Name = fmt.Sprintf("%s_stage_%v", c.prefix, i)
			stage.Alias = container.Name
			config.Stages = append(config.Stages, stage)
		}

		name := fmt.Sprintf("%s_step_%d", c.prefix, i)
		step := c.createProcess(name, container, "pipeline")
		stage.Steps = append(stage.Steps, step)
	}

	c.setupCacheRebuild(conf, config)

	return config
}

func (c *Compiler) setupCache(conf *yaml.Config, ir *backend.Config) {
	if c.local || len(conf.Cache) == 0 || c.cacher == nil {
		return
	}

	container := c.cacher.Restore(c.metadata.Repo.Name, c.metadata.Curr.Commit.Branch, conf.Cache)
	name := fmt.Sprintf("%s_restore_cache", c.prefix)
	step := c.createProcess(name, container, "cache")

	stage := new(backend.Stage)
	stage.Name = name
	stage.Alias = "restore_cache"
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)
}

func (c *Compiler) setupCacheRebuild(conf *yaml.Config, ir *backend.Config) {
	if c.local || len(conf.Cache) == 0 || c.metadata.Curr.Event != "push" || c.cacher == nil {
		return
	}
	container := c.cacher.Rebuild(c.metadata.Repo.Name, c.metadata.Curr.Commit.Branch, conf.Cache)

	name := fmt.Sprintf("%s_rebuild_cache", c.prefix)
	step := c.createProcess(name, container, "cache")

	stage := new(backend.Stage)
	stage.Name = name
	stage.Alias = "rebuild_cache"
	stage.Steps = append(stage.Steps, step)

	ir.Stages = append(ir.Stages, stage)
}

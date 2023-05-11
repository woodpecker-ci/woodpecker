package compiler

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"

	backend "github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/compiler/settings"
)

func (c *Compiler) createProcess(name string, container *yaml.Container, section string) (*backend.Step, error) {
	var (
		detached   bool
		workingdir string

		workspace   = fmt.Sprintf("%s_default:%s", c.prefix, c.base)
		privileged  = container.Privileged
		networkMode = container.NetworkMode
		ipcMode     = container.IpcMode
		// network    = container.Network
	)

	networks := []backend.Conn{
		{
			Name:    fmt.Sprintf("%s_default", c.prefix),
			Aliases: []string{container.Name},
		},
	}
	for _, network := range c.networks {
		networks = append(networks, backend.Conn{
			Name: network,
		})
	}

	var volumes []string
	if !c.local {
		volumes = append(volumes, workspace)
	}
	volumes = append(volumes, c.volumes...)
	for _, volume := range container.Volumes.Volumes {
		volumes = append(volumes, volume.String())
	}

	// append default environment variables
	environment := map[string]string{}
	for k, v := range container.Environment {
		environment[k] = v
	}
	for k, v := range c.env {
		switch v {
		case "", "0", "false":
			continue
		default:
			environment[k] = v
		}
	}

	environment["CI_WORKSPACE"] = path.Join(c.base, c.path)
	environment["CI_STEP_NAME"] = name

	if section == "services" || container.Detached {
		detached = true
	}

	if !detached || len(container.Commands) != 0 {
		workingdir = c.stepWorkdir(container)
	}

	if !detached {
		pluginSecrets := secretMap{}
		for name, secret := range c.secrets {
			if secret.Available(container) {
				pluginSecrets[name] = secret
			}
		}

		if err := settings.ParamsToEnv(container.Settings, environment, pluginSecrets.toStringMap()); err != nil {
			return nil, err
		}
	}

	if matchImage(container.Image, c.escalated...) && container.IsPlugin() {
		privileged = true
	}

	authConfig := backend.Auth{
		Username: container.AuthConfig.Username,
		Password: container.AuthConfig.Password,
		Email:    container.AuthConfig.Email,
	}
	for _, registry := range c.registries {
		if matchHostname(container.Image, registry.Hostname) {
			authConfig.Username = registry.Username
			authConfig.Password = registry.Password
			authConfig.Email = registry.Email
			break
		}
	}

	for _, requested := range container.Secrets.Secrets {
		secret, ok := c.secrets[strings.ToLower(requested.Source)]
		if ok && secret.Available(container) {
			environment[strings.ToUpper(requested.Target)] = secret.Value
		}
	}

	memSwapLimit := int64(container.MemSwapLimit)
	if c.reslimit.MemSwapLimit != 0 {
		memSwapLimit = c.reslimit.MemSwapLimit
	}
	memLimit := int64(container.MemLimit)
	if c.reslimit.MemLimit != 0 {
		memLimit = c.reslimit.MemLimit
	}
	shmSize := int64(container.ShmSize)
	if c.reslimit.ShmSize != 0 {
		shmSize = c.reslimit.ShmSize
	}
	cpuQuota := int64(container.CPUQuota)
	if c.reslimit.CPUQuota != 0 {
		cpuQuota = c.reslimit.CPUQuota
	}
	cpuShares := int64(container.CPUShares)
	if c.reslimit.CPUShares != 0 {
		cpuShares = c.reslimit.CPUShares
	}
	cpuSet := container.CPUSet
	if c.reslimit.CPUSet != "" {
		cpuSet = c.reslimit.CPUSet
	}

	// at least one constraint contain status success, or all constraints have no status set
	onSuccess := container.When.IncludesStatusSuccess()
	// at least one constraint must include the status failure.
	onFailure := container.When.IncludesStatusFailure()

	failure := container.Failure
	if container.Failure == "" {
		failure = frontend.FailureFail
	}

	return &backend.Step{
		Name:         name,
		Alias:        container.Name,
		Image:        container.Image,
		Pull:         container.Pull,
		Detached:     detached,
		Privileged:   privileged,
		WorkingDir:   workingdir,
		Environment:  environment,
		Labels:       container.Labels,
		Commands:     container.Commands,
		ExtraHosts:   container.ExtraHosts,
		Volumes:      volumes,
		Tmpfs:        container.Tmpfs,
		Devices:      container.Devices,
		Networks:     networks,
		DNS:          container.DNS,
		DNSSearch:    container.DNSSearch,
		MemSwapLimit: memSwapLimit,
		MemLimit:     memLimit,
		ShmSize:      shmSize,
		Sysctls:      container.Sysctls,
		CPUQuota:     cpuQuota,
		CPUShares:    cpuShares,
		CPUSet:       cpuSet,
		AuthConfig:   authConfig,
		OnSuccess:    onSuccess,
		OnFailure:    onFailure,
		Failure:      failure,
		NetworkMode:  networkMode,
		IpcMode:      ipcMode,
	}, nil
}

func (c *Compiler) stepWorkdir(container *yaml.Container) string {
	if filepath.IsAbs(container.Directory) {
		return container.Directory
	}
	return filepath.Join(c.base, c.path, container.Directory)
}

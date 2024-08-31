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
	"maps"
	"path"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"

	backend_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/compiler/settings"
	yaml_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/utils"
)

const (
	// The pluginWorkspaceBase should not be changed, only if you are sure what you do.
	pluginWorkspaceBase = "/woodpecker"
	// DefaultWorkspaceBase is set if not altered by the user.
	DefaultWorkspaceBase = pluginWorkspaceBase
)

func (c *Compiler) createProcess(container *yaml_types.Container, stepType backend_types.StepType) (*backend_types.Step, error) {
	var (
		uuid = ulid.Make()

		detached   bool
		workingDir string

		privileged  = container.Privileged
		networkMode = container.NetworkMode
	)

	workspaceBase := c.workspaceBase
	if container.IsPlugin() {
		// plugins have a predefined workspace base to not tamper with entrypoint executables
		workspaceBase = pluginWorkspaceBase
	}
	workspaceVolume := fmt.Sprintf("%s_default:%s", c.prefix, workspaceBase)

	networks := []backend_types.Conn{
		{
			Name:    fmt.Sprintf("%s_default", c.prefix),
			Aliases: []string{container.Name},
		},
	}
	for _, network := range c.networks {
		networks = append(networks, backend_types.Conn{
			Name: network,
		})
	}

	extraHosts := make([]backend_types.HostAlias, len(container.ExtraHosts))
	for i, extraHost := range container.ExtraHosts {
		name, ip, ok := strings.Cut(extraHost, ":")
		if !ok {
			return nil, &ErrExtraHostFormat{host: extraHost}
		}
		extraHosts[i].Name = name
		extraHosts[i].IP = ip
	}

	var volumes []string
	if !c.local {
		volumes = append(volumes, workspaceVolume)
	}
	volumes = append(volumes, c.volumes...)
	for _, volume := range container.Volumes.Volumes {
		volumes = append(volumes, volume.String())
	}

	// append default environment variables
	environment := map[string]string{}
	maps.Copy(environment, c.env)

	environment["CI_WORKSPACE"] = path.Join(workspaceBase, c.workspacePath)

	if stepType == backend_types.StepTypeService || container.Detached {
		detached = true
	}

	workingDir = c.stepWorkingDir(container)

	getSecretValue := func(name string) (string, error) {
		name = strings.ToLower(name)
		secret, ok := c.secrets[name]
		if !ok {
			return "", fmt.Errorf("secret %q not found", name)
		}

		event := c.metadata.Curr.Event
		err := secret.Available(event, container)
		if err != nil {
			return "", err
		}

		return secret.Value, nil
	}

	// TODO: why don't we pass secrets to detached steps?
	if !detached {
		if err := settings.ParamsToEnv(container.Settings, environment, "PLUGIN_", true, getSecretValue); err != nil {
			return nil, err
		}
	}

	if err := settings.ParamsToEnv(container.Environment, environment, "", false, getSecretValue); err != nil {
		return nil, err
	}

	for _, requested := range container.Secrets {
		secretValue, err := getSecretValue(requested)
		if err != nil {
			return nil, err
		}

		if !environmentAllowed(requested, stepType) {
			continue
		}

		environment[requested] = secretValue
	}

	if utils.MatchImageDynamic(container.Image, c.escalated...) && container.IsPlugin() {
		privileged = true
	}

	authConfig := backend_types.Auth{}
	for _, registry := range c.registries {
		if utils.MatchHostname(container.Image, registry.Hostname) {
			authConfig.Username = registry.Username
			authConfig.Password = registry.Password
			break
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

	var ports []backend_types.Port
	for _, portDef := range container.Ports {
		port, err := convertPort(portDef)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}

	// at least one constraint contain status success, or all constraints have no status set
	onSuccess := container.When.IncludesStatusSuccess()
	// at least one constraint must include the status failure.
	onFailure := container.When.IncludesStatusFailure()

	failure := container.Failure
	if container.Failure == "" {
		failure = metadata.FailureFail
	}

	return &backend_types.Step{
		Name:           container.Name,
		UUID:           uuid.String(),
		Type:           stepType,
		Image:          container.Image,
		Pull:           container.Pull,
		Detached:       detached,
		Privileged:     privileged,
		WorkingDir:     workingDir,
		Environment:    environment,
		Commands:       container.Commands,
		Entrypoint:     container.Entrypoint,
		ExtraHosts:     extraHosts,
		Volumes:        volumes,
		Tmpfs:          container.Tmpfs,
		Devices:        container.Devices,
		Networks:       networks,
		DNS:            container.DNS,
		DNSSearch:      container.DNSSearch,
		MemSwapLimit:   memSwapLimit,
		MemLimit:       memLimit,
		ShmSize:        shmSize,
		CPUQuota:       cpuQuota,
		CPUShares:      cpuShares,
		CPUSet:         cpuSet,
		AuthConfig:     authConfig,
		OnSuccess:      onSuccess,
		OnFailure:      onFailure,
		Failure:        failure,
		NetworkMode:    networkMode,
		Ports:          ports,
		BackendOptions: container.BackendOptions,
	}, nil
}

func (c *Compiler) stepWorkingDir(container *yaml_types.Container) string {
	if path.IsAbs(container.Directory) {
		return container.Directory
	}
	base := c.workspaceBase
	if container.IsPlugin() {
		base = pluginWorkspaceBase
	}
	return path.Join(base, c.workspacePath, container.Directory)
}

func convertPort(portDef string) (backend_types.Port, error) {
	var err error
	var port backend_types.Port

	number, protocol, _ := strings.Cut(portDef, "/")
	port.Protocol = protocol

	portNumber, err := strconv.ParseUint(number, 10, 16)
	if err != nil {
		return port, err
	}
	port.Number = uint16(portNumber)

	return port, nil
}

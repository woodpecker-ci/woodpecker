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
	"strings"

	"github.com/google/uuid"

	backend_types "go.woodpecker-ci.org/woodpecker/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/compiler/settings"
	yaml_types "go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/types"
	"go.woodpecker-ci.org/woodpecker/pipeline/frontend/yaml/utils"
)

func (c *Compiler) createProcess(name string, container *yaml_types.Container, stepType backend_types.StepType) (*backend_types.Step, error) {
	var (
		uuid = uuid.New()

		detached   bool
		workingdir string

		workspace   = fmt.Sprintf("%s_default:%s", c.prefix, c.base)
		privileged  = container.Privileged
		networkMode = container.NetworkMode
		ipcMode     = container.IpcMode
		// network    = container.Network
	)

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
	maps.Copy(environment, container.Environment)
	maps.Copy(environment, c.env)

	environment["CI_WORKSPACE"] = path.Join(c.base, c.path)
	environment["CI_STEP_NAME"] = name

	if stepType == backend_types.StepTypeService || container.Detached {
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

	if utils.MatchImage(container.Image, c.escalated...) && container.IsPlugin() {
		privileged = true
	}

	authConfig := backend_types.Auth{}
	for _, registry := range c.registries {
		if utils.MatchHostname(container.Image, registry.Hostname) {
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
		} else {
			return nil, fmt.Errorf("secret %q not found or not allowed to be used", requested.Source)
		}
	}

	// Advanced backend settings
	backendOptions := backend_types.BackendOptions{
		Kubernetes: convertKubernetesBackendOptions(&container.BackendOptions.Kubernetes),
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

	var ports []uint16
	for _, port := range container.Ports {
		ports = append(ports, uint16(port))
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
		Name:           name,
		UUID:           uuid.String(),
		Type:           stepType,
		Alias:          container.Name,
		Image:          container.Image,
		Pull:           container.Pull,
		Detached:       detached,
		Privileged:     privileged,
		WorkingDir:     workingdir,
		Environment:    environment,
		Commands:       container.Commands,
		ExtraHosts:     container.ExtraHosts,
		Volumes:        volumes,
		Tmpfs:          container.Tmpfs,
		Devices:        container.Devices,
		Networks:       networks,
		DNS:            container.DNS,
		DNSSearch:      container.DNSSearch,
		MemSwapLimit:   memSwapLimit,
		MemLimit:       memLimit,
		ShmSize:        shmSize,
		Sysctls:        container.Sysctls,
		CPUQuota:       cpuQuota,
		CPUShares:      cpuShares,
		CPUSet:         cpuSet,
		AuthConfig:     authConfig,
		OnSuccess:      onSuccess,
		OnFailure:      onFailure,
		Failure:        failure,
		NetworkMode:    networkMode,
		IpcMode:        ipcMode,
		Ports:          ports,
		BackendOptions: backendOptions,
	}, nil
}

func (c *Compiler) stepWorkdir(container *yaml_types.Container) string {
	if path.IsAbs(container.Directory) {
		return container.Directory
	}
	return path.Join(c.base, c.path, container.Directory)
}

func convertKubernetesBackendOptions(kubeOpt *yaml_types.KubernetesBackendOptions) backend_types.KubernetesBackendOptions {
	resources := backend_types.Resources{
		Limits:   kubeOpt.Resources.Limits,
		Requests: kubeOpt.Resources.Requests,
	}

	var tolerations []backend_types.Toleration
	for _, t := range kubeOpt.Tolerations {
		tolerations = append(tolerations, backend_types.Toleration{
			Key:               t.Key,
			Operator:          backend_types.TolerationOperator(t.Operator),
			Value:             t.Value,
			Effect:            backend_types.TaintEffect(t.Effect),
			TolerationSeconds: t.TolerationSeconds,
		})
	}

	var securityContext *backend_types.SecurityContext
	if kubeOpt.SecurityContext != nil {
		securityContext = &backend_types.SecurityContext{
			Privileged:   kubeOpt.SecurityContext.Privileged,
			RunAsNonRoot: kubeOpt.SecurityContext.RunAsNonRoot,
			RunAsUser:    kubeOpt.SecurityContext.RunAsUser,
			RunAsGroup:   kubeOpt.SecurityContext.RunAsGroup,
			FSGroup:      kubeOpt.SecurityContext.FSGroup,
		}
	}

	return backend_types.KubernetesBackendOptions{
		Resources:          resources,
		ServiceAccountName: kubeOpt.ServiceAccountName,
		NodeSelector:       kubeOpt.NodeSelector,
		Tolerations:        tolerations,
		SecurityContext:    securityContext,
	}
}

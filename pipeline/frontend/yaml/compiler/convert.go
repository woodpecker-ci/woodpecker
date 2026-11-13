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
	"slices"
	"strconv"
	"strings"

	"github.com/oklog/ulid/v2"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/compiler/settings"
	yaml_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/utils"
)

const (
	// The pluginWorkspaceBase should not be changed, only if you are sure what you do.
	pluginWorkspaceBase = "/woodpecker"
	// DefaultWorkspaceBase is set if not altered by the user.
	DefaultWorkspaceBase = pluginWorkspaceBase
)

func (c *Compiler) createProcess(container *yaml_types.Container, workflow *yaml_types.Workflow, stepType backend_types.StepType) (*backend_types.Step, error) {
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

	// always mount the workspace volume: it references config.Volume, which the
	// backend provisions, so steps mount the volume the backend actually creates
	volumes := []string{workspaceVolume}
	volumes = append(volumes, c.volumes...)
	for _, volume := range container.Volumes.Volumes {
		volumes = append(volumes, volume.String())
	}

	// append default environment variables
	environment := make(map[string]string)
	maps.Copy(environment, c.env)

	if !container.IsPlugin() {
		maps.Copy(environment, c.nonPluginEnv)
	}

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

	secretMapping := make(map[string]string)

	if err := settings.ParamsToEnv(container.Settings, environment, "PLUGIN_", true, getSecretValue, secretMapping); err != nil {
		return nil, err
	}

	if err := settings.ParamsToEnv(container.Environment, environment, "", false, getSecretValue, secretMapping); err != nil {
		return nil, err
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

	var ports []backend_types.Port
	for _, portDef := range container.Ports {
		port, err := convertPort(portDef)
		if err != nil {
			return nil, err
		}
		ports = append(ports, port)
	}

	whenEnv := make(map[string]string)
	maps.Copy(whenEnv, c.env)
	maps.Copy(whenEnv, c.nonPluginEnv)

	// at least one constraint contain status success, or all constraints have no status set
	onSuccess := container.When.IncludesStatusSuccess(c.metadata, false, whenEnv)
	// at least one constraint must include the status failure.
	onFailure := container.When.IncludesStatusFailure(c.metadata, false, whenEnv)

	failure := container.Failure
	if container.Failure == "" {
		failure = string(metadata.FailureFail)
	}

	// TODO: remove with version 4.x
	if c.forceIgnoreServiceFailure && detached {
		failure = string(metadata.FailureIgnore)
	}

	signalCfg, err := convertSignals(container.Signals)
	if err != nil {
		return nil, fmt.Errorf("step %q: %w", container.Name, err)
	}

	return &backend_types.Step{
		Name:            container.Name,
		UUID:            uuid.String(),
		Type:            stepType,
		Image:           container.Image,
		Pull:            container.Pull,
		Detached:        detached,
		Privileged:      privileged,
		WorkingDir:      workingDir,
		WorkspaceBase:   workspaceBase,
		Environment:     environment,
		SecretMapping:   secretMapping,
		Commands:        container.Commands,
		Entrypoint:      container.Entrypoint,
		ExtraHosts:      extraHosts,
		Volumes:         volumes,
		Tmpfs:           container.Tmpfs,
		Devices:         container.Devices,
		Networks:        networks,
		DNS:             container.DNS,
		DNSSearch:       container.DNSSearch,
		AuthConfig:      authConfig,
		OnSuccess:       onSuccess,
		OnFailure:       onFailure,
		Failure:         failure,
		ExitCodeSignals: signalCfg.exitCodeSignals,
		MaxRestarts:     signalCfg.maxRestarts,
		NetworkMode:     networkMode,
		Ports:           ports,
		BackendOptions:  container.BackendOptions,
		WorkflowLabels:  workflow.Labels,
	}, nil
}

// signalConfig is the compiled `signals:` option of a step.
type signalConfig struct {
	exitCodeSignals map[int]backend_types.StepSignal
	maxRestarts     int
}

// convertSignals turns the `signals:` map of a step (action -> exit codes and
// options) into a lookup by exit code plus the signal options. It rejects
// unknown actions, exit codes outside of 1..255, exit codes mapped more than
// once and invalid options.
func convertSignals(signals map[string]yaml_types.SignalConfig) (*signalConfig, error) {
	result := &signalConfig{}
	if len(signals) == 0 {
		return result, nil
	}

	result.exitCodeSignals = make(map[int]backend_types.StepSignal)
	// iterate in a stable order so the error for duplicated codes is deterministic
	for _, action := range slices.Sorted(maps.Keys(signals)) {
		signal := backend_types.StepSignal(action)
		if !signal.IsValid() {
			return nil, fmt.Errorf("unknown signal %q", action)
		}
		cfg := signals[action]

		if cfg.MaxRestarts != nil {
			if signal != backend_types.StepSignalRestart {
				return nil, fmt.Errorf("signal %q: max_restarts is only allowed for the restart signal", action)
			}
			if *cfg.MaxRestarts < 1 || *cfg.MaxRestarts > backend_types.StepSignalRestartLimit {
				return nil, fmt.Errorf("signal %q: max_restarts %d is out of range 1-%d",
					action, *cfg.MaxRestarts, backend_types.StepSignalRestartLimit)
			}
			result.maxRestarts = *cfg.MaxRestarts
		}

		for _, code := range cfg.ExitCodes {
			if code < backend_types.MinSignalExitCode || code > backend_types.MaxSignalExitCode {
				return nil, fmt.Errorf("signal %q: exit code %d is out of range %d-%d",
					action, code, backend_types.MinSignalExitCode, backend_types.MaxSignalExitCode)
			}
			if existing, ok := result.exitCodeSignals[code]; ok {
				return nil, fmt.Errorf("exit code %d is mapped to both signal %q and %q", code, existing, action)
			}
			result.exitCodeSignals[code] = signal
		}
	}

	return result, nil
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

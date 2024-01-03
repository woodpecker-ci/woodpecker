// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

// returns a container configuration.
func (e *docker) toConfig(step *types.Step) *container.Config {
	config := &container.Config{
		Image:        step.Image,
		Labels:       map[string]string{"wp_uuid": step.UUID},
		WorkingDir:   step.WorkingDir,
		AttachStdout: true,
		AttachStderr: true,
	}

	if len(step.Commands) != 0 {
		env, entry, cmd := common.GenerateContainerConf(step.Commands, e.info.OSType)
		for k, v := range env {
			step.Environment[k] = v
		}
		config.Entrypoint = entry
		config.Cmd = cmd
	}

	if len(step.Environment) != 0 {
		config.Env = toEnv(step.Environment)
	}
	if len(step.Volumes) != 0 {
		config.Volumes = toVol(step.Volumes)
	}
	return config
}

func toContainerName(step *types.Step) string {
	return "wp_" + step.UUID
}

// returns a container host configuration.
func toHostConfig(step *types.Step) *container.HostConfig {
	config := &container.HostConfig{
		Resources: container.Resources{
			CPUQuota:   step.CPUQuota,
			CPUShares:  step.CPUShares,
			CpusetCpus: step.CPUSet,
			Memory:     step.MemLimit,
			MemorySwap: step.MemSwapLimit,
		},
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
		Privileged: step.Privileged,
		ShmSize:    step.ShmSize,
		Sysctls:    step.Sysctls,
	}

	if len(step.NetworkMode) != 0 {
		config.NetworkMode = container.NetworkMode(step.NetworkMode)
	}
	if len(step.IpcMode) != 0 {
		config.IpcMode = container.IpcMode(step.IpcMode)
	}
	if len(step.DNS) != 0 {
		config.DNS = step.DNS
	}
	if len(step.DNSSearch) != 0 {
		config.DNSSearch = step.DNSSearch
	}
	extraHosts := []string{}
	for _, hostAlias := range step.ExtraHosts {
		extraHosts = append(extraHosts, hostAlias.Name+":"+hostAlias.IP)
	}
	if len(step.ExtraHosts) != 0 {
		config.ExtraHosts = extraHosts
	}
	if len(step.Devices) != 0 {
		config.Devices = toDev(step.Devices)
	}
	if len(step.Volumes) != 0 {
		config.Binds = step.Volumes
	}
	config.Tmpfs = map[string]string{}
	for _, path := range step.Tmpfs {
		if !strings.Contains(path, ":") {
			config.Tmpfs[path] = ""
			continue
		}
		parts, err := splitVolumeParts(path)
		if err != nil {
			continue
		}
		config.Tmpfs[parts[0]] = parts[1]
	}

	return config
}

// helper function that converts a slice of volume paths to a set of
// unique volume names.
func toVol(paths []string) map[string]struct{} {
	set := map[string]struct{}{}
	for _, path := range paths {
		parts, err := splitVolumeParts(path)
		if err != nil {
			continue
		}
		if len(parts) < 2 {
			continue
		}
		set[parts[1]] = struct{}{}
	}
	return set
}

// helper function that converts a key value map of environment variables to a
// string slice in key=value format.
func toEnv(env map[string]string) []string {
	var envs []string
	for k, v := range env {
		envs = append(envs, k+"="+v)
	}
	return envs
}

// helper function that converts a slice of device paths to a slice of
// container.DeviceMapping.
func toDev(paths []string) []container.DeviceMapping {
	var devices []container.DeviceMapping
	for _, path := range paths {
		parts, err := splitVolumeParts(path)
		if err != nil {
			continue
		}
		if len(parts) < 2 {
			continue
		}
		if strings.HasSuffix(parts[1], ":ro") || strings.HasSuffix(parts[1], ":rw") {
			parts[1] = parts[1][:len(parts[1])-1]
		}
		devices = append(devices, container.DeviceMapping{
			PathOnHost:        parts[0],
			PathInContainer:   parts[1],
			CgroupPermissions: "rwm",
		})
	}
	return devices
}

// helper function that serializes the auth configuration as JSON
// base64 payload.
func encodeAuthToBase64(authConfig types.Auth) (string, error) {
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(buf), nil
}

// helper function that split volume path
func splitVolumeParts(volumeParts string) ([]string, error) {
	pattern := `^((?:[\w]\:)?[^\:]*)\:((?:[\w]\:)?[^\:]*)(?:\:([rwom]*))?`
	r, err := regexp.Compile(pattern)
	if err != nil {
		return []string{}, err
	}
	if r.MatchString(volumeParts) {
		results := r.FindStringSubmatch(volumeParts)[1:]
		var cleanResults []string
		for _, item := range results {
			if item != "" {
				cleanResults = append(cleanResults, item)
			}
		}
		return cleanResults, nil
	}
	return strings.Split(volumeParts, ":"), nil
}

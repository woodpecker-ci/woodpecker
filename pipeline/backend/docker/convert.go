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
	"maps"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
)

// Valid container volumes must have at least two components, source and destination.
const minVolumeComponents = 2

// returns a container configuration.
func (e *docker) toConfig(step *types.Step) *container.Config {
	config := &container.Config{
		Image: step.Image,
		Labels: map[string]string{
			"wp_uuid": step.UUID,
			"wp_step": step.Name,
		},
		WorkingDir:   step.WorkingDir,
		AttachStdout: true,
		AttachStderr: true,
		Volumes:      toVol(step.Volumes),
	}
	configEnv := make(map[string]string)
	maps.Copy(configEnv, step.Environment)

	if len(step.Commands) > 0 {
		env, entry := common.GenerateContainerConf(step.Commands, e.info.OSType)
		for k, v := range env {
			configEnv[k] = v
		}
		config.Entrypoint = entry
	}
	if len(step.Entrypoint) > 0 {
		config.Entrypoint = step.Entrypoint
	}

	if len(configEnv) != 0 {
		config.Env = toEnv(configEnv)
	}
	return config
}

func toContainerName(step *types.Step) string {
	return "wp_" + step.UUID
}

// returns a container host configuration.
func toHostConfig(step *types.Step, conf *config) *container.HostConfig {
	config := &container.HostConfig{
		Resources: container.Resources{
			CPUQuota:   conf.resourceLimit.CPUQuota,
			CPUShares:  conf.resourceLimit.CPUShares,
			CpusetCpus: conf.resourceLimit.CPUSet,
			Memory:     conf.resourceLimit.MemLimit,
			MemorySwap: conf.resourceLimit.MemSwapLimit,
		},
		ShmSize: conf.resourceLimit.ShmSize,
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
		Privileged: step.Privileged,
	}

	if len(step.NetworkMode) != 0 {
		config.NetworkMode = container.NetworkMode(step.NetworkMode)
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
	if len(paths) == 0 {
		return nil
	}
	set := make(map[string]struct{})
	for _, path := range paths {
		parts, err := splitVolumeParts(path)
		if err != nil {
			continue
		}
		if len(parts) < minVolumeComponents {
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
		if k != "" {
			envs = append(envs, k+"="+v)
		}
	}
	return envs
}

// toDev converts a slice of volume paths to a set of device mappings for
// use in a Docker container config. It handles splitting the volume paths
// into host and container paths, and setting default permissions.
func toDev(paths []string) []container.DeviceMapping {
	var devices []container.DeviceMapping

	for _, path := range paths {
		parts, err := splitVolumeParts(path)
		if err != nil {
			continue
		}
		if len(parts) < minVolumeComponents {
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

// splitVolumeParts splits a volume string into its constituent parts.
//
// The parts are:
//
//  1. The path on the host machine
//  2. The path inside the container
//  3. The read/write mode
//
// It handles Windows and Linux style volume paths.
func splitVolumeParts(volumeParts string) ([]string, error) {
	// cspell:disable-next-line
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

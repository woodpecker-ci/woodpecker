package docker

import (
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"

	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
)

// returns a container configuration.
func toConfig(proc *types.Step) *container.Config {
	config := &container.Config{
		Image:        proc.Image,
		Labels:       proc.Labels,
		WorkingDir:   proc.WorkingDir,
		AttachStdout: true,
		AttachStderr: true,
	}
	if len(proc.Environment) != 0 {
		config.Env = toEnv(proc.Environment)
	}
	if len(proc.Command) != 0 {
		config.Cmd = proc.Command
	}
	if len(proc.Entrypoint) != 0 {
		config.Entrypoint = proc.Entrypoint
	}
	if len(proc.Volumes) != 0 {
		config.Volumes = toVol(proc.Volumes)
	}
	return config
}

// returns a container host configuration.
func toHostConfig(proc *types.Step) *container.HostConfig {
	config := &container.HostConfig{
		Resources: container.Resources{
			CPUQuota:   proc.CPUQuota,
			CPUShares:  proc.CPUShares,
			CpusetCpus: proc.CPUSet,
			Memory:     proc.MemLimit,
			MemorySwap: proc.MemSwapLimit,
		},
		LogConfig: container.LogConfig{
			Type: "json-file",
		},
		Privileged: proc.Privileged,
		ShmSize:    proc.ShmSize,
		Sysctls:    proc.Sysctls,
	}

	// if len(proc.VolumesFrom) != 0 {
	// 	config.VolumesFrom = proc.VolumesFrom
	// }
	if len(proc.NetworkMode) != 0 {
		config.NetworkMode = container.NetworkMode(proc.NetworkMode)
	}
	if len(proc.IpcMode) != 0 {
		config.IpcMode = container.IpcMode(proc.IpcMode)
	}
	if len(proc.DNS) != 0 {
		config.DNS = proc.DNS
	}
	if len(proc.DNSSearch) != 0 {
		config.DNSSearch = proc.DNSSearch
	}
	if len(proc.ExtraHosts) != 0 {
		config.ExtraHosts = proc.ExtraHosts
	}
	if len(proc.Devices) != 0 {
		config.Devices = toDev(proc.Devices)
	}
	if len(proc.Volumes) != 0 {
		config.Binds = proc.Volumes
	}
	config.Tmpfs = map[string]string{}
	for _, path := range proc.Tmpfs {
		if strings.Index(path, ":") == -1 {
			config.Tmpfs[path] = ""
			continue
		}
		parts, err := splitVolumeParts(path)
		if err != nil {
			continue
		}
		config.Tmpfs[parts[0]] = parts[1]
	}
	// if proc.OomKillDisable {
	// 	config.OomKillDisable = &proc.OomKillDisable
	// }

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
	} else {
		return strings.Split(volumeParts, ":"), nil
	}
}

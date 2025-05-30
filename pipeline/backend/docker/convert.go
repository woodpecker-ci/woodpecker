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
	"os"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/common"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// Valid container volumes must have at least two components, source and destination.
const minVolumeComponents = 2

// returns a container configuration.
func (e *docker) toConfig(step *types.Step, options BackendOptions) *container.Config {
	e.windowsPathPatch(step)

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
		User:         options.User,
	}
	configEnv := make(map[string]string)
	maps.Copy(configEnv, step.Environment)

	// Inject mirror configurations if any *_MIRROR_URL environment variables are set
	injectMirrorConfigs(configEnv)

	if len(step.Commands) > 0 {
		env, entry := common.GenerateContainerConf(step.Commands, e.info.OSType, step.WorkingDir)
		for k, v := range env {
			configEnv[k] = v
		}
		config.Entrypoint = entry

		// step.WorkingDir will be respected by the generated script
		config.WorkingDir = step.WorkspaceBase
	}
	if len(step.Entrypoint) > 0 {
		config.Entrypoint = step.Entrypoint
	}

	if len(configEnv) != 0 {
		config.Env = toEnv(configEnv)
	}
	return config
}

// injectMirrorConfigs adds mirror configuration environment variables
// based on the presence of *_MIRROR_URL environment variables
func injectMirrorConfigs(configEnv map[string]string) {
	// Get mirror URL from environment (step env takes precedence over system env)
	getMirrorURL := func(envKey string) string {
		if url := configEnv[envKey]; url != "" {
			return url
		}
		if url := os.Getenv(envKey); url != "" {
			return url
		}
		return ""
	}

	// Check which mirrors are configured and apply them
	var enabledServices []string

	// NPM Configuration
	if npmURL := getMirrorURL("NPM_MIRROR_URL"); npmURL != "" {
		enabledServices = append(enabledServices, "npm")
		if _, exists := configEnv["NPM_CONFIG_REGISTRY"]; !exists {
			configEnv["NPM_CONFIG_REGISTRY"] = npmURL
		}
		if _, exists := configEnv["NPM_CONFIG_CACHE"]; !exists {
			configEnv["NPM_CONFIG_CACHE"] = "/tmp/.npm"
		}
		if _, exists := configEnv["NPM_CONFIG_STRICT_SSL"]; !exists {
			configEnv["NPM_CONFIG_STRICT_SSL"] = "false"
		}
	}

	// PNPM Configuration
	if pnpmURL := getMirrorURL("PNPM_MIRROR_URL"); pnpmURL != "" {
		enabledServices = append(enabledServices, "pnpm")
		if _, exists := configEnv["PNPM_REGISTRY"]; !exists {
			configEnv["PNPM_REGISTRY"] = pnpmURL
		}
	}

	// Yarn Configuration
	if yarnURL := getMirrorURL("YARN_MIRROR_URL"); yarnURL != "" {
		enabledServices = append(enabledServices, "yarn")
		if _, exists := configEnv["YARN_REGISTRY"]; !exists {
			configEnv["YARN_REGISTRY"] = yarnURL
		}
	}

	// Alpine Configuration
	if alpineURL := getMirrorURL("ALPINE_MIRROR_URL"); alpineURL != "" {
		enabledServices = append(enabledServices, "alpine")
		if _, exists := configEnv["ALPINE_MIRROR"]; !exists {
			configEnv["ALPINE_MIRROR"] = alpineURL
		}
		if _, exists := configEnv["ALPINE_SETUP_SCRIPT"]; !exists {
			configEnv["ALPINE_SETUP_SCRIPT"] = "sed -i 's|dl-cdn.alpinelinux.org|" + strings.TrimPrefix(alpineURL, "https://") + "|g' /etc/apk/repositories 2>/dev/null || true"
		}
	}

	// Docker Configuration
	if dockerURL := getMirrorURL("DOCKER_MIRROR_URL"); dockerURL != "" {
		enabledServices = append(enabledServices, "docker")
		if _, exists := configEnv["DOCKER_REGISTRY_MIRROR"]; !exists {
			configEnv["DOCKER_REGISTRY_MIRROR"] = dockerURL
		}
	}

	// Python Pip Configuration
	if pipURL := getMirrorURL("PIP_MIRROR_URL"); pipURL != "" {
		enabledServices = append(enabledServices, "pip")
		if _, exists := configEnv["PIP_INDEX_URL"]; !exists {
			configEnv["PIP_INDEX_URL"] = pipURL
		}
		if _, exists := configEnv["PIP_TRUSTED_HOST"]; !exists {
			if strings.Contains(pipURL, "://") {
				parts := strings.Split(pipURL, "://")
				if len(parts) > 1 {
					host := strings.Split(parts[1], "/")[0]
					configEnv["PIP_TRUSTED_HOST"] = host
				}
			}
		}
	}

	// Go Configuration
	if goURL := getMirrorURL("GO_MIRROR_URL"); goURL != "" {
		enabledServices = append(enabledServices, "go")
		if _, exists := configEnv["GOPROXY"]; !exists {
			configEnv["GOPROXY"] = goURL
		}
		if _, exists := configEnv["GOSUMDB"]; !exists {
			configEnv["GOSUMDB"] = "sum.golang.google.cn"
		}
	}

	// Rust Configuration
	if rustURL := getMirrorURL("RUST_MIRROR_URL"); rustURL != "" {
		enabledServices = append(enabledServices, "rust")
		if _, exists := configEnv["CARGO_REGISTRIES_CRATES_IO_INDEX"]; !exists {
			configEnv["CARGO_REGISTRIES_CRATES_IO_INDEX"] = rustURL
		}
	}

	// Composer Configuration
	if composerURL := getMirrorURL("COMPOSER_MIRROR_URL"); composerURL != "" {
		enabledServices = append(enabledServices, "composer")
		if _, exists := configEnv["COMPOSER_REPO_PACKAGIST"]; !exists {
			configEnv["COMPOSER_REPO_PACKAGIST"] = composerURL
		}
	}

	// Maven Configuration
	if mavenURL := getMirrorURL("MAVEN_MIRROR_URL"); mavenURL != "" {
		enabledServices = append(enabledServices, "maven")
		if _, exists := configEnv["MAVEN_MIRROR_URL"]; !exists {
			configEnv["MAVEN_MIRROR_URL"] = mavenURL
		}
	}

	// Gradle Configuration
	if gradleURL := getMirrorURL("GRADLE_MIRROR_URL"); gradleURL != "" {
		enabledServices = append(enabledServices, "gradle")
		if _, exists := configEnv["GRADLE_REPO_OVERRIDE"]; !exists {
			configEnv["GRADLE_REPO_OVERRIDE"] = gradleURL
		}
	}

	// Ruby Configuration
	if rubyURL := getMirrorURL("RUBY_MIRROR_URL"); rubyURL != "" {
		enabledServices = append(enabledServices, "ruby")
		if _, exists := configEnv["BUNDLE_MIRROR__HTTPS___RUBYGEMS__ORG__"]; !exists {
			configEnv["BUNDLE_MIRROR__HTTPS___RUBYGEMS__ORG__"] = rubyURL
		}
	}

	// NuGet Configuration
	if nugetURL := getMirrorURL("NUGET_MIRROR_URL"); nugetURL != "" {
		enabledServices = append(enabledServices, "nuget")
		if _, exists := configEnv["NUGET_PACKAGES"]; !exists {
			configEnv["NUGET_PACKAGES"] = "/tmp/.nuget/packages"
		}
	}

	// Only create setup script if we have enabled services
	if len(enabledServices) > 0 {
		configEnv["MIRROR_SETUP_COMPLETE"] = "true"
		configEnv["ENABLED_MIRRORS"] = strings.Join(enabledServices, ",")
	}
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

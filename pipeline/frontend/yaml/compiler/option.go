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

package compiler

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/metadata"
)

// Option configures a compiler option.
type Option func(*Compiler)

// WithOption configures the compiler with the given option if
// boolean b evaluates to true.
func WithOption(option Option, b bool) Option {
	switch {
	case b:
		return option
	default:
		return func(compiler *Compiler) {}
	}
}

// WithVolumes configures the compiler with default volumes that
// are mounted to each container in the pipeline.
func WithVolumes(volumes ...string) Option {
	return func(compiler *Compiler) {
		compiler.volumes = volumes
	}
}

// WithRegistry configures the compiler with registry credentials
// that should be used to download images.
func WithRegistry(registries ...Registry) Option {
	return func(compiler *Compiler) {
		compiler.registries = registries
	}
}

// WithSecret configures the compiler with external secrets
// to be injected into the container at runtime.
func WithSecret(secrets ...Secret) Option {
	return func(compiler *Compiler) {
		for _, secret := range secrets {
			compiler.secrets[strings.ToLower(secret.Name)] = secret
		}
	}
}

// WithMetadata configures the compiler with the repository, pipeline
// and system metadata. The metadata is used to remove steps from
// the compiled pipeline configuration that should be skipped. The
// metadata is also added to each container as environment variables.
func WithMetadata(metadata metadata.Metadata) Option {
	return func(compiler *Compiler) {
		compiler.metadata = metadata

		for k, v := range metadata.Environ() {
			compiler.env[k] = v
		}
	}
}

// WithNetrc configures the compiler with netrc authentication
// credentials added by default to every container in the pipeline.
func WithNetrc(username, password, machine string) Option {
	return func(compiler *Compiler) {
		compiler.cloneEnv["CI_NETRC_USERNAME"] = username
		compiler.cloneEnv["CI_NETRC_PASSWORD"] = password
		compiler.cloneEnv["CI_NETRC_MACHINE"] = machine
	}
}

// WithWorkspace configures the compiler with the workspace base
// and path. The workspace base is a volume created at runtime and
// mounted into all containers in the pipeline. The base and path
// are joined to provide the working directory for all pipeline and
// plugin steps in the pipeline.
func WithWorkspace(base, path string) Option {
	return func(compiler *Compiler) {
		compiler.base = base
		compiler.path = path
	}
}

// WithWorkspaceFromURL configures the compiler with the workspace
// base and path based on the repository url.
func WithWorkspaceFromURL(base, link string) Option {
	path := "src"
	parsed, err := url.Parse(link)
	if err == nil {
		path = filepath.Join(path, parsed.Hostname(), parsed.Path)
	}
	return WithWorkspace(base, path)
}

// WithEscalated configures the compiler to automatically execute
// images as privileged containers if the match the given list.
func WithEscalated(images ...string) Option {
	return func(compiler *Compiler) {
		compiler.escalated = images
	}
}

// WithPrefix configures the compiler with the prefix. The prefix is
// used to prefix container, volume and network names to avoid
// collision at runtime.
func WithPrefix(prefix string) Option {
	return func(compiler *Compiler) {
		compiler.prefix = prefix
	}
}

// WithLocal configures the compiler with the local flag. The local
// flag indicates the pipeline execution is running in a local development
// environment with a mounted local working directory.
func WithLocal(local bool) Option {
	return func(compiler *Compiler) {
		compiler.local = local
	}
}

// WithEnviron configures the compiler with environment variables
// added by default to every container in the pipeline.
func WithEnviron(env map[string]string) Option {
	return func(compiler *Compiler) {
		for k, v := range env {
			compiler.env[k] = v
		}
	}
}

// WithCacher configures the compiler with default cache settings.
func WithCacher(cacher Cacher) Option {
	return func(compiler *Compiler) {
		compiler.cacher = cacher
	}
}

// WithVolumeCacher configures the compiler with default local volume
// caching enabled.
func WithVolumeCacher(base string) Option {
	return func(compiler *Compiler) {
		compiler.cacher = &volumeCacher{base: base}
	}
}

// WithS3Cacher configures the compiler with default amazon s3
// caching enabled.
func WithS3Cacher(access, secret, region, bucket string) Option {
	return func(compiler *Compiler) {
		compiler.cacher = &s3Cacher{
			access: access,
			secret: secret,
			bucket: bucket,
			region: region,
		}
	}
}

// WithProxy configures the compiler with HTTP_PROXY, HTTPS_PROXY,
// and NO_PROXY environment variables added by default to every
// container in the pipeline.
func WithProxy() Option {
	return WithEnviron(
		map[string]string{
			"no_proxy":    noProxy,
			"NO_PROXY":    noProxy,
			"http_proxy":  httpProxy,
			"HTTP_PROXY":  httpProxy,
			"HTTPS_PROXY": httpsProxy,
			"https_proxy": httpsProxy,
		},
	)
}

// WithNetworks configures the compiler with additional networks
// to be connected to pipeline containers
func WithNetworks(networks ...string) Option {
	return func(compiler *Compiler) {
		compiler.networks = networks
	}
}

// WithResourceLimit configures the compiler with default resource limits that
// are applied each container in the pipeline.
func WithResourceLimit(swap, mem, shmsize, cpuQuota, cpuShares int64, cpuSet string) Option {
	return func(compiler *Compiler) {
		compiler.reslimit = ResourceLimit{
			MemSwapLimit: swap,
			MemLimit:     mem,
			ShmSize:      shmsize,
			CPUQuota:     cpuQuota,
			CPUShares:    cpuShares,
			CPUSet:       cpuSet,
		}
	}
}

func WithDefaultCloneImage(cloneImage string) Option {
	return func(compiler *Compiler) {
		compiler.defaultCloneImage = cloneImage
	}
}

// WithTrusted configures the compiler with the trusted repo option
func WithTrusted(trusted bool) Option {
	return func(compiler *Compiler) {
		compiler.trustedPipeline = trusted
	}
}

// WithNetrcOnlyTrusted configures the compiler with the netrcOnlyTrusted repo option
func WithNetrcOnlyTrusted(only bool) Option {
	return func(compiler *Compiler) {
		compiler.netrcOnlyTrusted = only
	}
}

// TODO(bradrydzewski) consider an alternate approach to
// WithProxy where the proxy strings are passed directly
// to the function as named parameters.

// func WithProxy2(http, https, none string) Option {
// 	return WithEnviron(
// 		map[string]string{
// 			"no_proxy":    none,
// 			"NO_PROXY":    none,
// 			"http_proxy":  http,
// 			"HTTP_PROXY":  http,
// 			"HTTPS_PROXY": https,
// 			"https_proxy": https,
// 		},
// 	)
// }

var (
	noProxy    = getenv("no_proxy")
	httpProxy  = getenv("https_proxy")
	httpsProxy = getenv("https_proxy")
)

// getenv returns the named environment variable.
func getenv(name string) (value string) {
	name = strings.ToUpper(name)
	if value := os.Getenv(name); value != "" {
		return value
	}
	name = strings.ToLower(name)
	if value := os.Getenv(name); value != "" {
		return value
	}
	return
}

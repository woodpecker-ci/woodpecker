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
	"path"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
)

// Option configures a compiler option.
type Option func(*Compiler)

func noopOption() Option {
	return func(*Compiler) {}
}

// WithOption configures the compiler with the given option if
// boolean b evaluates to true.
func WithOption(option Option, b bool) Option {
	switch {
	case b:
		return option
	default:
		return func(_ *Compiler) {}
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
		compiler.workspaceBase = base
		compiler.workspacePath = path
	}
}

// WithWorkspaceFromURL configures the compiler with the workspace
// base and path based on the repository url.
func WithWorkspaceFromURL(base, u string) Option {
	srcPath := "src"
	parsed, err := url.Parse(u)
	if err == nil {
		srcPath = path.Join(srcPath, parsed.Hostname(), parsed.Path)
	}
	return WithWorkspace(base, srcPath)
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

// WithNetworks configures the compiler with additional networks
// to be connected to pipeline containers.
func WithNetworks(networks ...string) Option {
	return func(compiler *Compiler) {
		compiler.networks = networks
	}
}

func WithDefaultClonePlugin(cloneImage string) Option {
	return func(compiler *Compiler) {
		compiler.defaultClonePlugin = cloneImage
	}
}

func WithTrustedClonePlugins(images []string) Option {
	return func(compiler *Compiler) {
		compiler.trustedClonePlugins = images
	}
}

// WithTrusted configures the compiler with the trusted repo option.
func WithTrusted(trusted bool) Option {
	return func(compiler *Compiler) {
		compiler.trustedPipeline = trusted
	}
}

// WithNetrcOnlyTrusted configures the compiler with the netrcOnlyTrusted repo option.
func WithNetrcOnlyTrusted(only bool) Option {
	return func(compiler *Compiler) {
		compiler.netrcOnlyTrusted = only
	}
}

type ProxyOptions struct {
	NoProxy    string
	HTTPProxy  string
	HTTPSProxy string
}

// WithProxy configures the compiler with HTTP_PROXY, HTTPS_PROXY,
// and NO_PROXY environment variables added by default to every
// container in the pipeline.
func WithProxy(opt ProxyOptions) Option {
	if opt.HTTPProxy == "" &&
		opt.HTTPSProxy == "" &&
		opt.NoProxy == "" {
		return noopOption()
	}
	return WithEnviron(
		map[string]string{
			"no_proxy":    opt.NoProxy,
			"NO_PROXY":    opt.NoProxy,
			"http_proxy":  opt.HTTPProxy,
			"HTTP_PROXY":  opt.HTTPProxy,
			"HTTPS_PROXY": opt.HTTPSProxy,
			"https_proxy": opt.HTTPSProxy,
		},
	)
}

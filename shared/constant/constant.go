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

package constant

// PrivilegedPlugins can be changed by 'WOODPECKER_ESCALATE' at runtime.
var PrivilegedPlugins = []string{
	"docker.io/woodpeckerci/plugin-docker-buildx",
	"codeberg.org/woodpecker-plugins/docker-buildx",
}

// DefaultConfigOrder represent the priority in witch woodpecker search for a pipeline config by default
// folders are indicated by supplying a trailing slash.
var DefaultConfigOrder = [...]string{
	".woodpecker/",
	".woodpecker.yaml",
	".woodpecker.yml",
}

const (
	// DefaultCloneImage can be changed by 'WOODPECKER_DEFAULT_CLONE_IMAGE' at runtime.
	// renovate: datasource=docker depName=woodpeckerci/plugin-git
	DefaultCloneImage = "docker.io/woodpeckerci/plugin-git:2.5.2"
)

var TrustedCloneImages = []string{
	// we should trust to inject netrc to the clone step image we assign ourselves
	DefaultCloneImage,
	// we should trust the latest versions of our clone plugin(s)
	"docker.io/woodpeckerci/plugin-git:latest",
	"quay.io/woodpeckerci/plugin-git:latest",
	// alternate valid trusted images
	// renovate: datasource=docker depName=quay.io/woodpeckerci/plugin-git
	"quay.io/woodpeckerci/plugin-git:2.5.2",

	// allow the dev image
	"docker.io/woodpeckerci/plugin-git:next",

	// old version witch we know have no problem (e.g. allow-list)
	"docker.io/woodpeckerci/plugin-git:2.5.2",
	"quay.io/woodpeckerci/plugin-git:2.5.2",
	"docker.io/woodpeckerci/plugin-git:2.5.1",
	"quay.io/woodpeckerci/plugin-git:2.5.1",
	"docker.io/woodpeckerci/plugin-git:2.5.0",
	"quay.io/woodpeckerci/plugin-git:2.5.0",
	"docker.io/woodpeckerci/plugin-git:2.4.0",
	"quay.io/woodpeckerci/plugin-git:2.4.0",
	"docker.io/woodpeckerci/plugin-git:2.3.1",
	"quay.io/woodpeckerci/plugin-git:2.3.1",
	"docker.io/woodpeckerci/plugin-git:2.3.0",
	"quay.io/woodpeckerci/plugin-git:2.3.0",
	"docker.io/woodpeckerci/plugin-git:2.2.0",
	"quay.io/woodpeckerci/plugin-git:2.2.0",
	"docker.io/woodpeckerci/plugin-git:2.1.2",
	"quay.io/woodpeckerci/plugin-git:2.1.2",
	"docker.io/woodpeckerci/plugin-git:2.1.0",
	"quay.io/woodpeckerci/plugin-git:2.1.0",
	"docker.io/woodpeckerci/plugin-git:2.0.3",
	"quay.io/woodpeckerci/plugin-git:2.0.3",
	"docker.io/woodpeckerci/plugin-git:2.0.2",
	"quay.io/woodpeckerci/plugin-git:2.0.2",
	"docker.io/woodpeckerci/plugin-git:2.0.1",
	"quay.io/woodpeckerci/plugin-git:2.0.1",
}

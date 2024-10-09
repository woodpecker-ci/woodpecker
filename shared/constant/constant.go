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
	"plugins/docker",
	"plugins/gcr",
	"plugins/ecr",
	"woodpeckerci/plugin-docker-buildx:5",
	"codeberg.org/woodpecker-plugins/docker-buildx:5",
	"woodpeckerci/plugin-docker-buildx:5.0",
	"codeberg.org/woodpecker-plugins/docker-buildx:5.0",
	"woodpeckerci/plugin-docker-buildx:5.0.0",
	"codeberg.org/woodpecker-plugins/docker-buildx:5.0.0",
	"woodpeckerci/plugin-docker-buildx:latest",
	"codeberg.org/woodpecker-plugins/docker-buildx:latest",
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
	DefaultCloneImage = "docker.io/woodpeckerci/plugin-git:2.5.1"
)

var TrustedCloneImages = []string{
	DefaultCloneImage,
	"quay.io/woodpeckerci/plugin-git",
}

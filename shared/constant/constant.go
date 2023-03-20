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

// PrivilegedPlugins can be changed by 'WOODPECKER_ESCALATE' at runtime
var PrivilegedPlugins = []string{
	"docker.io/plugins/docker",
	"docker.io/plugins/gcr",
	"docker.io/plugins/ecr",
	"docker.io/woodpeckerci/plugin-docker-buildx",
	"quay.io/woodpeckerci/plugin-docker-buildx",
	"codeberg.org/woodpecker-plugins/docker-buildx",
}

// DefaultConfigOrder represent the priority in witch woodpecker search for a pipeline config by default
// folders are indicated by supplying a trailing /
var DefaultConfigOrder = [...]string{
	".woodpecker/",
	".woodpecker.yml",
	".woodpecker.yaml",
	".drone.yml",
}

const (
	// DefaultCloneImage can be changed by 'WOODPECKER_DEFAULT_CLONE_IMAGE' at runtime
	DefaultCloneImage = "quay.io/woodpeckerci/plugin-git:2.0.3"
)

var TrustedCloneImages = []string{
	DefaultCloneImage,
}

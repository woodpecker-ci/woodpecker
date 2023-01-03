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

var PrivilegedPlugins = []string{
	"docker.io/plugins/docker:20.14.0",
	"docker.io/plugins/gcr:20.14.0",
	"docker.io/plugins/ecr:20.14.0",
	"docker.io/woodpeckerci/plugin-docker-buildx:2.1.0",
	// "docker.io/woodpeckerci/plugin-docker",
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
	DefaultCloneImage = "docker.io/woodpeckerci/plugin-git:2.0.3"
)

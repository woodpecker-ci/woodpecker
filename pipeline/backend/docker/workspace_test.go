// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestWorkspaceBindIndex(t *testing.T) {
	binds := []string{
		"/etc/ssl/certs:/etc/ssl/certs:ro",
		"/home/me/repo:/woodpecker/src",
		"wp_123_default:/woodpecker",
		"cache:/cache",
	}
	assert.Equal(t, 2, workspaceBindIndex(binds, "/woodpecker"))
	assert.Equal(t, -1, workspaceBindIndex(binds, "/go"))
	assert.Equal(t, -1, workspaceBindIndex(nil, "/woodpecker"))
	assert.Equal(t, 0, workspaceBindIndex([]string{"wp_123_default:/woodpecker:ro"}, "/woodpecker"))
}

func TestOverlayVolumeNames(t *testing.T) {
	overlay, upper, work := overlayVolumeNames(&backend_types.Step{UUID: "01abc"})
	assert.Equal(t, "wp_01abc_overlay", overlay)
	assert.Equal(t, "wp_01abc_upper", upper)
	assert.Equal(t, "wp_01abc_work", work)
}

func TestOverlayDriverOpts(t *testing.T) {
	assert.Equal(t, map[string]string{
		"type":   "overlay",
		"device": "overlay",
		"o":      "lowerdir=/var/lib/docker/volumes/ws/_data,upperdir=/var/lib/docker/volumes/up/_data,workdir=/var/lib/docker/volumes/work/_data",
	}, overlayDriverOpts("/var/lib/docker/volumes/ws/_data", "/var/lib/docker/volumes/up/_data", "/var/lib/docker/volumes/work/_data"))
}

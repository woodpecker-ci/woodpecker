// Copyright 2023 Woodpecker Authors
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

package compiler

import (
	"path"
	"strings"

	yaml_types "go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/yaml/types"
)

// Cacher defines a compiler transform that can be used
// to implement default caching for a repository.
type Cacher interface {
	Restore(repo, branch string, mounts []string) *yaml_types.Container
	Rebuild(repo, branch string, mounts []string) *yaml_types.Container
}

type volumeCacher struct {
	base string
}

func (c *volumeCacher) Restore(repo, branch string, mounts []string) *yaml_types.Container {
	return &yaml_types.Container{
		Name:  "rebuild_cache",
		Image: "plugins/volume-cache:1.0.0",
		Settings: map[string]any{
			"mount":       mounts,
			"path":        "/cache",
			"restore":     true,
			"file":        strings.ReplaceAll(branch, "/", "_") + ".tar",
			"fallback_to": "main.tar",
		},
		Volumes: yaml_types.Volumes{
			Volumes: []*yaml_types.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

func (c *volumeCacher) Rebuild(repo, branch string, mounts []string) *yaml_types.Container {
	return &yaml_types.Container{
		Name:  "rebuild_cache",
		Image: "plugins/volume-cache:1.0.0",
		Settings: map[string]any{
			"mount":   mounts,
			"path":    "/cache",
			"rebuild": true,
			"flush":   true,
			"file":    strings.ReplaceAll(branch, "/", "_") + ".tar",
		},
		Volumes: yaml_types.Volumes{
			Volumes: []*yaml_types.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

type s3Cacher struct {
	bucket string
	access string
	secret string
	region string
}

func (c *s3Cacher) Restore(_, _ string, mounts []string) *yaml_types.Container {
	return &yaml_types.Container{
		Name:  "rebuild_cache",
		Image: "plugins/s3-cache:latest",
		Settings: map[string]any{
			"mount":      mounts,
			"access_key": c.access,
			"secret_key": c.secret,
			"bucket":     c.bucket,
			"region":     c.region,
			"rebuild":    true,
		},
	}
}

func (c *s3Cacher) Rebuild(_, _ string, mounts []string) *yaml_types.Container {
	return &yaml_types.Container{
		Name:  "rebuild_cache",
		Image: "plugins/s3-cache:latest",
		Settings: map[string]any{
			"mount":      mounts,
			"access_key": c.access,
			"secret_key": c.secret,
			"bucket":     c.bucket,
			"region":     c.region,
			"rebuild":    true,
			"flush":      true,
		},
	}
}

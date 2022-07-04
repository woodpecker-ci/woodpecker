package compiler

import (
	"path"
	"strings"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
)

// Cacher defines a compiler transform that can be used
// to implement default caching for a repository.
type Cacher interface {
	Restore(repo, branch string, mounts []string) *yaml.Container
	Rebuild(repo, branch string, mounts []string) *yaml.Container
}

type volumeCacher struct {
	base string
}

func (c *volumeCacher) Restore(repo, branch string, mounts []string) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/volume-cache:1.0.0",
		Settings: map[string]interface{}{
			"mount":       mounts,
			"path":        "/cache",
			"restore":     true,
			"file":        strings.ReplaceAll(branch, "/", "_") + ".tar",
			"fallback_to": "master.tar",
		},
		Volumes: types.Volumes{
			Volumes: []*types.Volume{
				{
					Source:      path.Join(c.base, repo),
					Destination: "/cache",
					// TODO add access mode
				},
			},
		},
	}
}

func (c *volumeCacher) Rebuild(repo, branch string, mounts []string) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/volume-cache:1.0.0",
		Settings: map[string]interface{}{
			"mount":   mounts,
			"path":    "/cache",
			"rebuild": true,
			"flush":   true,
			"file":    strings.ReplaceAll(branch, "/", "_") + ".tar",
		},
		Volumes: types.Volumes{
			Volumes: []*types.Volume{
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

func (c *s3Cacher) Restore(repo, branch string, mounts []string) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/s3-cache:latest",
		Settings: map[string]interface{}{
			"mount":      mounts,
			"access_key": c.access,
			"secret_key": c.secret,
			"bucket":     c.bucket,
			"region":     c.region,
			"rebuild":    true,
		},
	}
}

func (c *s3Cacher) Rebuild(repo, branch string, mounts []string) *yaml.Container {
	return &yaml.Container{
		Name:  "rebuild_cache",
		Image: "plugins/s3-cache:latest",
		Settings: map[string]interface{}{
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

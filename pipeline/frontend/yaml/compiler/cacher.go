package compiler

import (
	"path"
	"strings"

	yaml_types "github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml/types"
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
		Settings: map[string]interface{}{
			"mount":       mounts,
			"path":        "/cache",
			"restore":     true,
			"file":        strings.Replace(branch, "/", "_", -1) + ".tar",
			"fallback_to": "master.tar",
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
		Settings: map[string]interface{}{
			"mount":   mounts,
			"path":    "/cache",
			"rebuild": true,
			"flush":   true,
			"file":    strings.Replace(branch, "/", "_", -1) + ".tar",
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

func (c *s3Cacher) Rebuild(_, _ string, mounts []string) *yaml_types.Container {
	return &yaml_types.Container{
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

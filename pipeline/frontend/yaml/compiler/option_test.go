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
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/frontend/metadata"
	"go.woodpecker-ci.org/woodpecker/v2/shared/constant"
)

func TestWithWorkspace(t *testing.T) {
	compiler := New(
		WithWorkspace(
			"/pipeline",
			"src/github.com/octocat/hello-world",
		),
	)
	assert.Equal(t, "/pipeline", compiler.workspaceBase)
	assert.Equal(t, "src/github.com/octocat/hello-world", compiler.workspacePath)
}

func TestWithEscalated(t *testing.T) {
	compiler := New(
		WithEscalated(
			"docker",
			"docker-dev",
		),
	)
	assert.Equal(t, "docker", compiler.escalated[0])
	assert.Equal(t, "docker-dev", compiler.escalated[1])
}

func TestWithVolumes(t *testing.T) {
	compiler := New(
		WithVolumes(
			"/tmp:/tmp",
			"/foo:/foo",
		),
	)
	assert.Equal(t, "/tmp:/tmp", compiler.volumes[0])
	assert.Equal(t, "/foo:/foo", compiler.volumes[1])
}

func TestWithNetworks(t *testing.T) {
	compiler := New(
		WithNetworks(
			"overlay_1",
			"overlay_bar",
		),
	)
	assert.Equal(t, "overlay_1", compiler.networks[0])
	assert.Equal(t, "overlay_bar", compiler.networks[1])
}

func TestWithResourceLimit(t *testing.T) {
	compiler := New(
		WithResourceLimit(
			1,
			2,
			3,
			4,
			5,
			"0,2-5",
		),
	)
	assert.EqualValues(t, 1, compiler.reslimit.MemSwapLimit)
	assert.EqualValues(t, 2, compiler.reslimit.MemLimit)
	assert.EqualValues(t, 3, compiler.reslimit.ShmSize)
	assert.EqualValues(t, 4, compiler.reslimit.CPUQuota)
	assert.EqualValues(t, 5, compiler.reslimit.CPUShares)
	assert.Equal(t, "0,2-5", compiler.reslimit.CPUSet)
}

func TestWithPrefix(t *testing.T) {
	assert.Equal(t, "someprefix_", New(WithPrefix("someprefix_")).prefix)
}

func TestWithMetadata(t *testing.T) {
	metadata := metadata.Metadata{
		Repo: metadata.Repo{
			Owner:    "octacat",
			Name:     "hello-world",
			Private:  true,
			ForgeURL: "https://github.com/octocat/hello-world",
			CloneURL: "https://github.com/octocat/hello-world.git",
		},
	}
	compiler := New(
		WithMetadata(metadata),
	)

	assert.Equal(t, metadata, compiler.metadata)
	assert.Equal(t, metadata.Repo.Name, compiler.env["CI_REPO_NAME"])
	assert.Equal(t, metadata.Repo.ForgeURL, compiler.env["CI_REPO_URL"])
	assert.Equal(t, metadata.Repo.CloneURL, compiler.env["CI_REPO_CLONE_URL"])
}

func TestWithLocal(t *testing.T) {
	assert.True(t, New(WithLocal(true)).local)
	assert.False(t, New(WithLocal(false)).local)
}

func TestWithNetrc(t *testing.T) {
	compiler := New(
		WithNetrc(
			"octocat",
			"password",
			"github.com",
		),
	)
	assert.Equal(t, "octocat", compiler.cloneEnv["CI_NETRC_USERNAME"])
	assert.Equal(t, "password", compiler.cloneEnv["CI_NETRC_PASSWORD"])
	assert.Equal(t, "github.com", compiler.cloneEnv["CI_NETRC_MACHINE"])
}

func TestWithProxy(t *testing.T) {
	// alter the default values
	noProxy := "example.com"
	httpProxy := "bar.com"
	httpsProxy := "baz.com"

	testdata := map[string]string{
		"no_proxy":    noProxy,
		"NO_PROXY":    noProxy,
		"http_proxy":  httpProxy,
		"HTTP_PROXY":  httpProxy,
		"https_proxy": httpsProxy,
		"HTTPS_PROXY": httpsProxy,
	}
	compiler := New(
		WithProxy(ProxyOptions{
			NoProxy:    noProxy,
			HTTPProxy:  httpProxy,
			HTTPSProxy: httpsProxy,
		}),
	)
	for key, value := range testdata {
		assert.Equal(t, value, compiler.env[key])
	}
}

func TestWithEnviron(t *testing.T) {
	compiler := New(
		WithEnviron(
			map[string]string{
				"RACK_ENV": "development",
				"SHOW":     "true",
			},
		),
	)
	assert.Equal(t, "development", compiler.env["RACK_ENV"])
	assert.Equal(t, "true", compiler.env["SHOW"])
}

func TestDefaultClonePlugin(t *testing.T) {
	compiler := New(
		WithDefaultClonePlugin("not-an-image"),
	)
	assert.Equal(t, "not-an-image", compiler.defaultClonePlugin)
}

func TestWithTrustedClonePlugins(t *testing.T) {
	compiler := New(WithTrustedClonePlugins([]string{"not-an-image"}))
	assert.ElementsMatch(t, []string{"not-an-image"}, compiler.trustedClonePlugins)

	compiler = New()
	assert.ElementsMatch(t, constant.TrustedClonePlugins, compiler.trustedClonePlugins)
}

// Copyright 2026 Woodpecker Authors
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

package registry

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func writeDockerConfig(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func auth(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func TestFilesystem(t *testing.T) {
	all := &model.ListOptionsWithAll{All: true}

	t.Run("no docker config configured", func(t *testing.T) {
		regs, err := NewFilesystem("").GlobalRegistryList(all)
		require.NoError(t, err)
		assert.Empty(t, regs)
	})

	t.Run("plain and encoded credentials are read only registries", func(t *testing.T) {
		fs := NewFilesystem(writeDockerConfig(t, `{"auths":{`+
			`"docker.io":{"username":"user","password":"plain"},`+
			`"ghcr.io":{"auth":"`+auth("bot:pass:with:colons")+`"}}}`))

		regs, err := fs.GlobalRegistryList(all)
		require.NoError(t, err)
		assert.ElementsMatch(t, []*model.Registry{
			{Address: "docker.io", Username: "user", Password: "plain", ReadOnly: true},
			// only the first colon separates user and password
			{Address: "ghcr.io", Username: "bot", Password: "pass:with:colons", ReadOnly: true},
		}, regs)

		reg, err := fs.GlobalRegistryFind("ghcr.io")
		require.NoError(t, err)
		assert.Equal(t, "bot", reg.Username)

		_, err = fs.GlobalRegistryFind("quay.io")
		assert.ErrorIs(t, err, types.ErrRecordNotExist)
	})

	t.Run("pagination", func(t *testing.T) {
		fs := NewFilesystem(writeDockerConfig(t, `{"auths":{"a.io":{"username":"a"},"b.io":{"username":"b"},"c.io":{"username":"c"}}}`))

		regs, err := fs.GlobalRegistryList(&model.ListOptionsWithAll{ListOptions: &model.ListOptions{Page: 2, PerPage: 2}})
		require.NoError(t, err)
		assert.Len(t, regs, 1)
	})

	t.Run("unusable credential helpers are skipped", func(t *testing.T) {
		fs := NewFilesystem(writeDockerConfig(t, `{"auths":{"docker.io":{"username":"user","password":"pw"}},`+
			`"credHelpers":{"helper.example.com":"woodpecker-test-missing-helper"}}`))

		regs, err := fs.GlobalRegistryList(all)
		require.NoError(t, err)
		assert.Equal(t, []*model.Registry{{Address: "docker.io", Username: "user", Password: "pw", ReadOnly: true}}, regs)
	})

	for name, content := range map[string]string{
		"invalid json":            `{"auths":`,
		"auth is not base64":      `{"auths":{"docker.io":{"auth":"not base64!"}}}`,
		"auth without password":   `{"auths":{"docker.io":{"auth":"` + auth("user") + `"}}}`,
		"auth without a username": `{"auths":{"docker.io":{"auth":"` + auth(":password") + `"}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			fs := NewFilesystem(writeDockerConfig(t, content))
			_, err := fs.GlobalRegistryList(all)
			assert.Error(t, err)
			_, err = fs.GlobalRegistryFind("docker.io")
			assert.Error(t, err)
		})
	}

	t.Run("missing file", func(t *testing.T) {
		_, err := NewFilesystem(filepath.Join(t.TempDir(), "missing.json")).GlobalRegistryList(all)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})
}

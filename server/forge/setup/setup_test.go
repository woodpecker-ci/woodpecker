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

package setup

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestForgeUnknownType(t *testing.T) {
	t.Parallel()
	_, err := Forge(&model.Forge{Type: model.ForgeType("nope")})
	assert.Error(t, err)
}

func TestForgeBitbucket(t *testing.T) {
	t.Parallel()
	// bitbucket needs no URL, so it constructs successfully
	f, err := Forge(&model.Forge{
		ID:                1,
		Type:              model.ForgeTypeBitbucket,
		OAuthClientID:     "id",
		OAuthClientSecret: "secret",
	})
	require.NoError(t, err)
	assert.NotNil(t, f)
}

func TestForgeGitHub(t *testing.T) {
	t.Parallel()
	f, err := Forge(&model.Forge{
		ID:   1,
		Type: model.ForgeTypeGithub,
		URL:  "https://github.com",
	})
	require.NoError(t, err)
	assert.NotNil(t, f)
}

func TestForgeGiteaRequiresURL(t *testing.T) {
	t.Parallel()
	_, err := Forge(&model.Forge{Type: model.ForgeTypeGitea, URL: ""})
	assert.Error(t, err)
}

func TestForgeForgejoRequiresURL(t *testing.T) {
	t.Parallel()
	_, err := Forge(&model.Forge{Type: model.ForgeTypeForgejo, URL: ""})
	assert.Error(t, err)
}

func TestForgeBitbucketDatacenterMissingOptions(t *testing.T) {
	t.Parallel()

	t.Run("missing git-username", func(t *testing.T) {
		t.Parallel()
		_, err := Forge(&model.Forge{
			Type:              model.ForgeTypeBitbucketDatacenter,
			AdditionalOptions: map[string]any{},
		})
		assert.Error(t, err)
	})

	t.Run("missing git-password", func(t *testing.T) {
		t.Parallel()
		_, err := Forge(&model.Forge{
			Type:              model.ForgeTypeBitbucketDatacenter,
			AdditionalOptions: map[string]any{"git-username": "u"},
		})
		assert.Error(t, err)
	})

	t.Run("missing oauth-enable-project-admin-scope", func(t *testing.T) {
		t.Parallel()
		_, err := Forge(&model.Forge{
			Type: model.ForgeTypeBitbucketDatacenter,
			AdditionalOptions: map[string]any{
				"git-username": "u",
				"git-password": "p",
			},
		})
		assert.Error(t, err)
	})
}

func TestForgeAddonMissingExecutable(t *testing.T) {
	t.Parallel()
	_, err := Forge(&model.Forge{
		Type:              model.ForgeTypeAddon,
		AdditionalOptions: map[string]any{},
	})
	assert.Error(t, err)
}

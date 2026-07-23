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
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func generateAppPrivateKey(t *testing.T) string {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
}

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

func TestForgeGitHubApp(t *testing.T) {
	t.Parallel()
	pemKey := generateAppPrivateKey(t)

	t.Run("app-id as string", func(t *testing.T) {
		t.Parallel()
		f, err := Forge(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			AdditionalOptions: map[string]any{
				"app-id":          "12345",
				"app-private-key": pemKey,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, f)
		assert.Equal(t, "github", f.Name())
	})

	t.Run("app-id as JSON number", func(t *testing.T) {
		t.Parallel()
		// the API rejects non-string app ids at save time, so setup treats a
		// numeric value as unset - together with the private key this fails
		// the pairing check instead of being silently coerced
		_, err := Forge(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			AdditionalOptions: map[string]any{
				"app-id":          float64(12345),
				"app-private-key": pemKey,
			},
		})
		require.Error(t, err)
	})

	t.Run("app-id without private key", func(t *testing.T) {
		t.Parallel()
		_, err := Forge(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			AdditionalOptions: map[string]any{
				"app-id": "12345",
			},
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "app id")
		assert.Contains(t, err.Error(), "private key")
	})

	t.Run("app disabled when neither is set", func(t *testing.T) {
		t.Parallel()
		f, err := Forge(&model.Forge{
			ID:   1,
			Type: model.ForgeTypeGithub,
			URL:  "https://github.com",
			AdditionalOptions: map[string]any{
				"merge-ref":   true,
				"public-only": true,
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, f)
	})
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

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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestForge_RedactSecrets(t *testing.T) {
	t.Run("GithubWithPrivateKey", func(t *testing.T) {
		forge := &Forge{
			Type: ForgeTypeGithub,
			AdditionalOptions: map[string]any{
				"app-id":          int64(123),
				"app-private-key": "-----BEGIN RSA PRIVATE KEY-----\nsecret\n-----END RSA PRIVATE KEY-----",
			},
		}

		forge.RedactSecrets()

		assert.Equal(t, map[string]any{
			"app-id":              int64(123),
			"app-private-key-set": true,
		}, forge.AdditionalOptions)
	})

	t.Run("GithubWithEmptyPrivateKey", func(t *testing.T) {
		forge := &Forge{
			Type: ForgeTypeGithub,
			AdditionalOptions: map[string]any{
				"app-private-key": "",
			},
		}

		forge.RedactSecrets()

		assert.Equal(t, map[string]any{}, forge.AdditionalOptions)
	})

	t.Run("GithubWithoutPrivateKey", func(t *testing.T) {
		forge := &Forge{
			Type: ForgeTypeGithub,
			AdditionalOptions: map[string]any{
				"app-id": int64(123),
			},
		}

		forge.RedactSecrets()

		assert.Equal(t, map[string]any{
			"app-id": int64(123),
		}, forge.AdditionalOptions)
	})

	t.Run("NonGithubForgeUntouched", func(t *testing.T) {
		forge := &Forge{
			Type: ForgeTypeBitbucketDatacenter,
			AdditionalOptions: map[string]any{
				"git-username":    "ci-user",
				"git-password":    "hunter2",
				"app-private-key": "should stay, not a github forge",
			},
		}

		forge.RedactSecrets()

		assert.Equal(t, map[string]any{
			"git-username":    "ci-user",
			"git-password":    "hunter2",
			"app-private-key": "should stay, not a github forge",
		}, forge.AdditionalOptions)
	})

	t.Run("NilAdditionalOptions", func(t *testing.T) {
		forge := &Forge{
			Type: ForgeTypeGithub,
		}

		assert.NotPanics(t, forge.RedactSecrets)
		assert.Nil(t, forge.AdditionalOptions)
	})

	t.Run("NilForge", func(t *testing.T) {
		var forge *Forge

		assert.NotPanics(t, forge.RedactSecrets)
	})
}

func TestForge_PublicCopy(t *testing.T) {
	forge := &Forge{
		ID:                1,
		Type:              ForgeTypeGithub,
		URL:               "https://github.com",
		OAuthClientID:     "client-id",
		OAuthClientSecret: "client-secret",
		SkipVerify:        true,
		OAuthHost:         "https://oauth.example.com",
		AdditionalOptions: map[string]any{
			"app-private-key": "secret",
		},
	}

	public := forge.PublicCopy()

	assert.Equal(t, &Forge{
		ID:   1,
		Type: ForgeTypeGithub,
		URL:  "https://github.com",
	}, public)

	// the original forge must not be modified
	assert.Equal(t, "client-secret", forge.OAuthClientSecret)
	assert.Equal(t, map[string]any{"app-private-key": "secret"}, forge.AdditionalOptions)
}

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

package services

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// TestSetupForgeServiceKeepsOrgs ensures the allowed orgs of the forge
// configured through environment variables survive a server start. Only the
// options that have an environment setting are written back, the orgs are
// edited in the admin UI and must not be reset.
func TestSetupForgeServiceKeepsOrgs(t *testing.T) {
	_store := store_mocks.NewMockStore(t)
	_store.On("ForgeGet", int64(1)).Return(&model.Forge{
		ID:   1,
		URL:  "https://github.com",
		Orgs: []string{"github-org"},
	}, nil)

	var updated *model.Forge
	_store.On("ForgeUpdate", mock.Anything).Run(func(args mock.Arguments) {
		updated, _ = args.Get(0).(*model.Forge)
	}).Return(nil)

	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "addon-forge"},
			&cli.BoolFlag{Name: "github", Value: true},
			&cli.BoolFlag{Name: "github-merge-ref"},
			&cli.BoolFlag{Name: "github-public-only"},
			&cli.StringFlag{Name: "forge-oauth-client", Value: "client-id"},
			&cli.StringFlag{Name: "forge-oauth-secret", Value: "client-secret"},
			&cli.StringFlag{Name: "forge-url", Value: "https://github.com"},
			&cli.StringFlag{Name: "forge-oauth-host"},
			&cli.BoolFlag{Name: "forge-skip-verify"},
		},
	}

	assert.NoError(t, setupForgeService(cmd, _store))

	assert.NotNil(t, updated)
	assert.Equal(t, []string{"github-org"}, updated.Orgs)
	// the options coming from the environment are written back
	assert.Equal(t, model.ForgeTypeGithub, updated.Type)
	assert.Equal(t, "client-id", updated.OAuthClientID)
}

func forgeCommand(flags ...cli.Flag) *cli.Command {
	defaults := []cli.Flag{
		&cli.StringFlag{Name: "addon-forge"},
		&cli.BoolFlag{Name: "github"},
		&cli.BoolFlag{Name: "github-merge-ref"},
		&cli.BoolFlag{Name: "github-public-only"},
		&cli.BoolFlag{Name: "gitlab"},
		&cli.BoolFlag{Name: "gitea"},
		&cli.BoolFlag{Name: "forgejo"},
		&cli.BoolFlag{Name: "bitbucket"},
		&cli.BoolFlag{Name: "bitbucket-dc"},
		&cli.StringFlag{Name: "bitbucket-dc-git-username"},
		&cli.StringFlag{Name: "bitbucket-dc-git-password"},
		&cli.BoolFlag{Name: "bitbucket-dc-oauth-enable-oauth2-scope-project-admin"},
		&cli.StringFlag{Name: "forge-oauth-client"},
		&cli.StringFlag{Name: "forge-oauth-secret"},
		&cli.StringFlag{Name: "forge-url"},
		&cli.StringFlag{Name: "forge-oauth-host"},
		&cli.BoolFlag{Name: "forge-skip-verify"},
	}
	// later flags with the same name replace the defaults
	byName := make(map[string]int)
	for i, f := range defaults {
		byName[f.Names()[0]] = i
	}
	for _, f := range flags {
		if i, ok := byName[f.Names()[0]]; ok {
			defaults[i] = f
			continue
		}
		defaults = append(defaults, f)
	}
	return &cli.Command{Flags: defaults}
}

func TestSetupForgeService(t *testing.T) {
	for _, tc := range []struct {
		name    string
		flags   []cli.Flag
		want    model.ForgeType
		url     string
		options map[string]any
	}{
		{
			name:    "github defaults to github.com",
			flags:   []cli.Flag{&cli.BoolFlag{Name: "github", Value: true}, &cli.BoolFlag{Name: "github-merge-ref", Value: true}},
			want:    model.ForgeTypeGithub,
			url:     "https://github.com",
			options: map[string]any{"merge-ref": true, "public-only": false},
		},
		{
			name:    "gitlab defaults to gitlab.com",
			flags:   []cli.Flag{&cli.BoolFlag{Name: "gitlab", Value: true}},
			want:    model.ForgeTypeGitlab,
			url:     "https://gitlab.com",
			options: make(map[string]any),
		},
		{
			name:    "gitea keeps a configured url",
			flags:   []cli.Flag{&cli.BoolFlag{Name: "gitea", Value: true}, &cli.StringFlag{Name: "forge-url", Value: "https://gitea.example.com"}},
			want:    model.ForgeTypeGitea,
			url:     "https://gitea.example.com",
			options: make(map[string]any),
		},
		{
			name:    "forgejo defaults to next.forgejo.org",
			flags:   []cli.Flag{&cli.BoolFlag{Name: "forgejo", Value: true}},
			want:    model.ForgeTypeForgejo,
			url:     "https://next.forgejo.org",
			options: make(map[string]any),
		},
		{
			name:    "bitbucket",
			flags:   []cli.Flag{&cli.BoolFlag{Name: "bitbucket", Value: true}},
			want:    model.ForgeTypeBitbucket,
			options: make(map[string]any),
		},
		{
			name: "bitbucket data center with machine account",
			flags: []cli.Flag{
				&cli.BoolFlag{Name: "bitbucket-dc", Value: true},
				&cli.StringFlag{Name: "forge-url", Value: "https://bitbucket.example.com"},
				&cli.StringFlag{Name: "bitbucket-dc-git-username", Value: "machine"},
				&cli.StringFlag{Name: "bitbucket-dc-git-password", Value: "secret"},
				&cli.BoolFlag{Name: "bitbucket-dc-oauth-enable-oauth2-scope-project-admin", Value: true},
			},
			want: model.ForgeTypeBitbucketDatacenter,
			url:  "https://bitbucket.example.com",
			options: map[string]any{
				"git-username":                     "machine",
				"git-password":                     "secret",
				"oauth-enable-project-admin-scope": true,
			},
		},
		{
			name: "addon wins over other forge flags",
			flags: []cli.Flag{
				&cli.StringFlag{Name: "addon-forge", Value: "/usr/bin/forge-addon"},
				&cli.BoolFlag{Name: "github", Value: true},
			},
			want:    model.ForgeTypeAddon,
			options: map[string]any{"executable": "/usr/bin/forge-addon"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// the first start creates the forge
			_store := store_mocks.NewMockStore(t)
			_store.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)
			var created *model.Forge
			_store.On("ForgeCreate", mock.Anything).Run(func(args mock.Arguments) {
				created, _ = args.Get(0).(*model.Forge)
			}).Return(nil)

			flags := append([]cli.Flag{&cli.StringFlag{Name: "forge-oauth-client", Value: " client-id "}}, tc.flags...)
			require.NoError(t, setupForgeService(forgeCommand(flags...), _store))

			require.NotNil(t, created)
			assert.Equal(t, tc.want, created.Type)
			assert.Equal(t, tc.url, created.URL)
			assert.Equal(t, tc.options, created.AdditionalOptions)
			// surrounding whitespace of copied credentials is dropped
			assert.Equal(t, "client-id", created.OAuthClientID)
		})
	}

	t.Run("no forge configured", func(t *testing.T) {
		_store := store_mocks.NewMockStore(t)
		_store.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)
		assert.EqualError(t, setupForgeService(forgeCommand(), _store), "forge not configured")
	})

	t.Run("store errors", func(t *testing.T) {
		_store := store_mocks.NewMockStore(t)
		_store.On("ForgeGet", int64(1)).Return(nil, assert.AnError)
		assert.ErrorIs(t, setupForgeService(forgeCommand(&cli.BoolFlag{Name: "gitea", Value: true}), _store), assert.AnError)

		_store = store_mocks.NewMockStore(t)
		_store.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)
		_store.On("ForgeCreate", mock.Anything).Return(assert.AnError)
		assert.ErrorIs(t, setupForgeService(forgeCommand(&cli.BoolFlag{Name: "gitea", Value: true}), _store), assert.AnError)

		_store = store_mocks.NewMockStore(t)
		_store.On("ForgeGet", int64(1)).Return(&model.Forge{ID: 1}, nil)
		_store.On("ForgeUpdate", mock.Anything).Return(assert.AnError)
		assert.ErrorIs(t, setupForgeService(forgeCommand(&cli.BoolFlag{Name: "gitea", Value: true}), _store), assert.AnError)
	})
}

func TestSetupSignatureKeys(t *testing.T) {
	t.Run("first start generates and stores a key", func(t *testing.T) {
		_store := store_mocks.NewMockStore(t)
		_store.On("ServerConfigGet", "signature-private-key").Return("", types.ErrRecordNotExist)
		var stored string
		_store.On("ServerConfigSet", "signature-private-key", mock.Anything).Run(func(args mock.Arguments) {
			stored = args.String(1)
		}).Return(nil)

		priv, pub, err := setupSignatureKeys(_store)
		require.NoError(t, err)
		assert.Equal(t, hex.EncodeToString(priv), stored)
		assert.Equal(t, priv.Public(), pub)

		// the next start loads the same key
		_store = store_mocks.NewMockStore(t)
		_store.On("ServerConfigGet", "signature-private-key").Return(stored, nil)
		loaded, loadedPub, err := setupSignatureKeys(_store)
		require.NoError(t, err)
		assert.Equal(t, priv, loaded)
		assert.Equal(t, pub, loadedPub)
	})

	t.Run("errors", func(t *testing.T) {
		_store := store_mocks.NewMockStore(t)
		_store.On("ServerConfigGet", "signature-private-key").Return("", assert.AnError)
		_, _, err := setupSignatureKeys(_store)
		assert.ErrorIs(t, err, assert.AnError)

		_store = store_mocks.NewMockStore(t)
		_store.On("ServerConfigGet", "signature-private-key").Return("not hex", nil)
		_, _, err = setupSignatureKeys(_store)
		assert.ErrorContains(t, err, "failed to decode private key")

		_store = store_mocks.NewMockStore(t)
		_store.On("ServerConfigGet", "signature-private-key").Return("", types.ErrRecordNotExist)
		_store.On("ServerConfigSet", "signature-private-key", mock.Anything).Return(assert.AnError)
		_, _, err = setupSignatureKeys(_store)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestSetupConfigService(t *testing.T) {
	command := func(flags ...cli.Flag) *cli.Command {
		return &cli.Command{Flags: append([]cli.Flag{
			&cli.DurationFlag{Name: "forge-timeout"},
			&cli.StringSliceFlag{Name: "default-pipeline-configs"},
			&cli.StringSliceFlag{Name: "default-pipeline-config-extensions"},
			&cli.BoolFlag{Name: "config-extension-netrc"},
		}, flags...)}
	}

	_, err := setupConfigService(command(&cli.UintFlag{Name: "forge-retry"}, &cli.StringFlag{Name: "config-extension-endpoint"}, &cli.BoolFlag{Name: "config-extension-exclusive"}), nil)
	assert.EqualError(t, err, "WOODPECKER_FORGE_RETRY can not be 0")

	server := extensionServer(t)
	client := testClient(t)
	forgeConfigs := []*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: []byte("steps: [forge]")}}
	extensionConfigs := []*forge_types.FileMeta{{Name: "extension.yaml", Data: []byte("steps: []")}}

	newForge := func(t *testing.T, calls int) *forge_mocks.MockForge {
		f := forge_mocks.NewMockForge(t)
		if calls > 0 {
			f.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return([]byte("steps: [forge]"), nil).Times(calls)
		}
		return f
	}

	for _, tc := range []struct {
		name       string
		endpoint   string
		exclusive  bool
		forgeCalls int
		want       []*forge_types.FileMeta
	}{
		{name: "forge only", forgeCalls: 1, want: forgeConfigs},
		{name: "extension after forge", endpoint: server.URL + "/config", forgeCalls: 1, want: extensionConfigs},
		{name: "exclusive extension", endpoint: server.URL + "/config", exclusive: true, want: extensionConfigs},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service, err := setupConfigService(command(
				&cli.UintFlag{Name: "forge-retry", Value: 1},
				&cli.StringFlag{Name: "config-extension-endpoint", Value: tc.endpoint},
				&cli.BoolFlag{Name: "config-extension-exclusive", Value: tc.exclusive},
			), client)
			require.NoError(t, err)

			configs, err := service.Fetch(t.Context(), newForge(t, tc.forgeCalls), &model.User{}, &model.Repo{Config: ".woodpecker.yaml"}, &model.Pipeline{}, nil, false)
			require.NoError(t, err)
			assert.Equal(t, tc.want, configs)
		})
	}
}

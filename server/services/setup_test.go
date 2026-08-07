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
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
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

// forgeSetupFlags declares the flags setupForgeService reads, mirroring the
// definitions in cmd/server/flags.go.
func forgeSetupFlags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{Name: "forge-oauth-client"},
		&cli.StringFlag{Name: "forge-oauth-secret"},
		&cli.StringFlag{Name: "forge-url"},
		&cli.BoolFlag{Name: "forge-skip-verify"},
		&cli.StringFlag{Name: "forge-oauth-host"},
		&cli.StringFlag{Name: "addon-forge"},
		&cli.BoolFlag{Name: "github"},
		&cli.BoolFlag{Name: "github-merge-ref"},
		&cli.BoolFlag{Name: "github-public-only"},
		&cli.StringFlag{Name: "github-app-id"},
		&cli.StringFlag{Name: "github-app-private-key"},
		&cli.BoolFlag{Name: "gitlab"},
		&cli.BoolFlag{Name: "gitea"},
		&cli.BoolFlag{Name: "forgejo"},
		&cli.BoolFlag{Name: "bitbucket"},
		&cli.BoolFlag{Name: "bitbucket-dc"},
		&cli.StringFlag{Name: "bitbucket-dc-git-username"},
		&cli.StringFlag{Name: "bitbucket-dc-git-password"},
		&cli.BoolFlag{Name: "bitbucket-dc-oauth-enable-oauth2-scope-project-admin"},
	}
}

// runForgeSetup runs setupForgeService behind a cli.Command so flag values are
// parsed exactly like in production, and returns the error of setupForgeService.
func runForgeSetup(t *testing.T, args []string, mockStore *store_mocks.MockStore, setupForge SetupForge) error {
	t.Helper()
	var setupErr error
	cmd := &cli.Command{
		Flags: forgeSetupFlags(),
		Action: func(_ context.Context, c *cli.Command) error {
			setupErr = setupForgeService(c, mockStore, setupForge)
			return nil
		},
	}
	require.NoError(t, cmd.Run(t.Context(), append([]string{"woodpecker"}, args...)))
	return setupErr
}

func captureForge(target **model.Forge) func(args mock.Arguments) {
	return func(args mock.Arguments) {
		if f, ok := args.Get(0).(*model.Forge); ok {
			*target = f
		}
	}
}

func TestSetupForgeServiceGithubAppCreate(t *testing.T) {
	t.Parallel()
	pemKey := generateAppPrivateKey(t)

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)
	var created *model.Forge
	mockStore.On("ForgeCreate", mock.AnythingOfType("*model.Forge")).Run(captureForge(&created)).Return(nil)

	var validated *model.Forge
	setupForge := func(f *model.Forge) (forge.Forge, error) {
		validated = f
		return nil, nil
	}

	err := runForgeSetup(t, []string{
		"--github",
		"--forge-oauth-client", " oauth-client ",
		"--forge-oauth-secret", " oauth-secret ",
		"--github-merge-ref",
		"--github-public-only",
		"--github-app-id", "12345",
		"--github-app-private-key", pemKey,
	}, mockStore, setupForge)
	require.NoError(t, err)

	require.NotNil(t, created)
	assert.Equal(t, model.ForgeTypeGithub, created.Type)
	assert.Equal(t, "https://github.com", created.URL, "github forge URL should default to github.com")
	assert.Equal(t, "oauth-client", created.OAuthClientID, "oauth client id should be trimmed")
	assert.Equal(t, "oauth-secret", created.OAuthClientSecret, "oauth client secret should be trimmed")
	assert.Equal(t, "12345", created.AdditionalOptions["app-id"])
	assert.Equal(t, pemKey, created.AdditionalOptions["app-private-key"])
	assert.Equal(t, true, created.AdditionalOptions["merge-ref"])
	assert.Equal(t, true, created.AdditionalOptions["public-only"])

	// the fully populated forge must have been validated before it was persisted
	require.NotNil(t, validated)
	assert.Same(t, created, validated)
}

func TestSetupForgeServiceGithubAppUpdate(t *testing.T) {
	t.Parallel()
	pemKey := generateAppPrivateKey(t)

	// an existing install enables GitHub App credentials on restart
	existing := &model.Forge{
		ID:   1,
		Type: model.ForgeTypeGithub,
		URL:  "https://github.com",
		AdditionalOptions: map[string]any{
			"merge-ref":   true,
			"public-only": false,
		},
	}
	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("ForgeGet", int64(1)).Return(existing, nil)
	var updated *model.Forge
	mockStore.On("ForgeUpdate", mock.AnythingOfType("*model.Forge")).Run(captureForge(&updated)).Return(nil)

	setupForge := func(_ *model.Forge) (forge.Forge, error) {
		return nil, nil
	}

	err := runForgeSetup(t, []string{
		"--github",
		"--github-app-id", "54321",
		"--github-app-private-key", pemKey,
	}, mockStore, setupForge)
	require.NoError(t, err)

	require.NotNil(t, updated)
	assert.EqualValues(t, 1, updated.ID)
	assert.Equal(t, "54321", updated.AdditionalOptions["app-id"])
	assert.Equal(t, pemKey, updated.AdditionalOptions["app-private-key"])
	mockStore.AssertNotCalled(t, "ForgeCreate", mock.Anything)
}

func TestSetupForgeServiceGithubAppInvalidCredentials(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)

	pairErr := errors.New("github app id and app private key must be provided together")
	setupForge := func(f *model.Forge) (forge.Forge, error) {
		assert.Equal(t, "12345", f.AdditionalOptions["app-id"])
		assert.Empty(t, f.AdditionalOptions["app-private-key"])
		return nil, pairErr
	}

	err := runForgeSetup(t, []string{"--github", "--github-app-id", "12345"}, mockStore, setupForge)
	require.ErrorIs(t, err, pairErr)
	assert.ErrorContains(t, err, "invalid forge configuration")

	// fail fast: an invalid configuration must never be persisted
	mockStore.AssertNotCalled(t, "ForgeCreate", mock.Anything)
	mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
}

func TestSetupForgeServiceGithubNilSetupForge(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)
	var created *model.Forge
	mockStore.On("ForgeCreate", mock.AnythingOfType("*model.Forge")).Run(captureForge(&created)).Return(nil)

	// app-id without a private key would fail validation, but with a nil
	// factory no validation runs and the forge is persisted as-is
	err := runForgeSetup(t, []string{"--github", "--github-app-id", "12345"}, mockStore, nil)
	require.NoError(t, err)

	require.NotNil(t, created)
	assert.Equal(t, model.ForgeTypeGithub, created.Type)
	assert.Equal(t, "12345", created.AdditionalOptions["app-id"])
}

func TestSetupForgeServiceNotConfigured(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)
	mockStore.On("ForgeGet", int64(1)).Return(nil, types.ErrRecordNotExist)

	validated := false
	setupForge := func(_ *model.Forge) (forge.Forge, error) {
		validated = true
		return nil, nil
	}

	err := runForgeSetup(t, nil, mockStore, setupForge)
	assert.ErrorContains(t, err, "forge not configured")
	assert.False(t, validated, "validation should not run when no forge is configured")
	mockStore.AssertNotCalled(t, "ForgeCreate", mock.Anything)
	mockStore.AssertNotCalled(t, "ForgeUpdate", mock.Anything)
}

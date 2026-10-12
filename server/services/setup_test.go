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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
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

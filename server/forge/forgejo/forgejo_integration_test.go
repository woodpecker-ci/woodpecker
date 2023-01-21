// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0

package forgejo

import (
	"context"
	"os"
	"testing"

	"github.com/woodpecker-ci/woodpecker/server/model"

	"github.com/stretchr/testify/assert"
	"github.com/woodpecker-ci/woodpecker/server/forge"
)

func newTestForgejo(t *testing.T) forge.Forge {
	if os.Getenv("FORGEJO_URL") == "" {
		t.Skip("FORGEJO_URL not set in the environment")
	}

	f, err := New(Opts{
		URL:        os.Getenv("FORGEJO_URL"),
		SkipVerify: true,
		Debug:      true,
		PerPage:    5,
	})
	assert.NoError(t, err)
	return f
}

func Test_forgejo_integration_File(t *testing.T) {
	f := newTestForgejo(t)
	cfg, err := f.File(context.Background(),
		&model.User{Token: os.Getenv("FORGEJO_TOKEN_ADMIN")},
		&model.Repo{Owner: "root", Name: "forgejo-test"},
		&model.Pipeline{Commit: os.Getenv("FORGEJO_COMMIT")},
		"README")
	assert.NoError(t, err)
	assert.Contains(t, string(cfg), "SOMETHING")
}

func Test_forgejo_integration_FileShaNotFound(t *testing.T) {
	f := newTestForgejo(t)
	_, err := f.File(context.Background(),
		&model.User{Token: os.Getenv("FORGEJO_TOKEN_ADMIN")},
		&model.Repo{Owner: "root", Name: "forgejo-test"},
		&model.Pipeline{Commit: "13245"},
		"README")
	if assert.Error(t, err) {
		assert.ErrorContains(t, err, "sha not found")
	}
}

func Test_forgejo_integration_Dir(t *testing.T) {
	f := newTestForgejo(t)
	cfg, err := f.Dir(context.Background(),
		&model.User{Token: os.Getenv("FORGEJO_TOKEN_ADMIN")},
		&model.Repo{Owner: "root", Name: "forgejo-test"},
		&model.Pipeline{Commit: os.Getenv("FORGEJO_COMMIT")},
		".woodpecker/")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cfg))
}

func Test_forgejo_integration_DirShaNotFound(t *testing.T) {
	f := newTestForgejo(t)
	_, err := f.Dir(context.Background(),
		&model.User{Token: os.Getenv("FORGEJO_TOKEN_ADMIN")},
		&model.Repo{Owner: "root", Name: "forgejo-test"},
		&model.Pipeline{Commit: "13245"},
		".woodpecker/")
	if assert.Error(t, err) {
		assert.ErrorContains(t, err, "sha not found")
	}
}

func Test_forgejo_integration_Perm(t *testing.T) {
	f := newTestForgejo(t)
	perm, err := f.Perm(context.Background(),
		&model.User{Token: os.Getenv("FORGEJO_TOKEN_ADMIN")},
		&model.Repo{Owner: "root", Name: "forgejo-test"})
	assert.NoError(t, err)
	assert.True(t, perm.Pull)
	assert.True(t, perm.Push)
	assert.True(t, perm.Admin)
}

func Test_forgejo_integration_OrgMembership(t *testing.T) {
	f := newTestForgejo(t)
	perm, err := f.OrgMembership(context.Background(),
		&model.User{
			Token: os.Getenv("FORGEJO_TOKEN_ADMIN"),
			Login: "root",
		},
		"myorg")
	assert.NoError(t, err)
	assert.Equal(t, &model.OrgPerm{
		Member: true,
		Admin:  true,
	}, perm)

	perm, err = f.OrgMembership(context.Background(),
		&model.User{
			Token: os.Getenv("FORGEJO_TOKEN_USER"),
			Login: "normaluser",
		},
		"myorg")
	assert.NoError(t, err)
	assert.Equal(t, &model.OrgPerm{
		Member: true,
		Admin:  true,
	}, perm)
}

func Test_forgejo_integration_Teams(t *testing.T) {
	f := newTestForgejo(t)
	teams, err := f.Teams(context.Background(),
		&model.User{
			Token: os.Getenv("FORGEJO_TOKEN_USER"),
			Login: "normaluser",
		})
	assert.NoError(t, err)
	if assert.Equal(t, 1, len(teams)) {
		assert.Equal(t, "myorg", teams[0].Login)
	}
}

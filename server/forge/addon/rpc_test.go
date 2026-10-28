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

package addon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/go-plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	forge_mocks "go.woodpecker-ci.org/woodpecker/v3/server/forge/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// newRPCForge connects the addon RPC client to an RPC server backed by impl
// in process, so every call takes the same JSON and net/rpc path as with a
// real addon binary.
func newRPCForge(t *testing.T, impl forge.Forge) forge.Forge {
	t.Helper()

	client, _ := plugin.TestPluginRPCConn(t, map[string]plugin.Plugin{
		pluginKey: &Plugin{Impl: impl},
	}, nil)
	t.Cleanup(func() { _ = client.Close() })

	raw, err := client.Dispense(pluginKey)
	require.NoError(t, err)
	f, ok := raw.(forge.Forge)
	require.True(t, ok)
	return f
}

// user and repo carry the fields model.User and model.Repo hide from JSON,
// the addon wrappers must still transport them.
func testUser() *model.User {
	return &model.User{
		ID:            1,
		ForgeID:       2,
		ForgeRemoteID: "3",
		Login:         "octocat",
		Email:         "octocat@example.com",
		AccessToken:   "access",
		RefreshToken:  "refresh",
		Expiry:        42,
		Hash:          "user-hash",
	}
}

func testRepo() *model.Repo {
	return &model.Repo{
		ID:            4,
		UserID:        1,
		ForgeRemoteID: "5",
		Owner:         "octocat",
		Name:          "hello",
		FullName:      "octocat/hello",
		WebhookSecret: "repo-hash",
		Perm:          &model.Perm{Pull: true, Push: true},
	}
}

func TestRPCRoundTrip(t *testing.T) {
	ctx := t.Context()
	user := testUser()
	repo := testRepo()
	pipeline := &model.Pipeline{Number: 6, Event: model.EventPush, Commit: "abc", Ref: "refs/heads/main"}
	listOpts := &model.ListOptions{Page: 2, PerPage: 10}

	t.Run("name and url", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Name").Return("addon")
		impl.On("URL").Return("https://forge.example.com")

		f := newRPCForge(t, impl)
		assert.Equal(t, "addon", f.Name())
		assert.Equal(t, "https://forge.example.com", f.URL())
	})

	t.Run("login", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Login", mock.Anything, &types.OAuthRequest{Code: "code", State: "state"}).Return(user, "/redirect", nil)

		u, redirect, err := newRPCForge(t, impl).Login(ctx, &types.OAuthRequest{Code: "code", State: "state"})
		require.NoError(t, err)
		assert.Equal(t, user, u)
		assert.Equal(t, "/redirect", redirect)
	})

	t.Run("teams", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		teams := []*model.Team{{Login: "team", Avatar: "https://avatar"}}
		impl.On("Teams", mock.Anything, user, listOpts).Return(teams, nil)

		got, err := newRPCForge(t, impl).Teams(ctx, user, listOpts)
		require.NoError(t, err)
		assert.Equal(t, teams, got)
	})

	t.Run("repo", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Repo", mock.Anything, user, model.ForgeRemoteID("5"), "octocat", "hello").Return(repo, nil)

		got, err := newRPCForge(t, impl).Repo(ctx, user, "5", "octocat", "hello")
		require.NoError(t, err)
		assert.Equal(t, repo, got)
	})

	t.Run("repos", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Repos", mock.Anything, user, listOpts).Return([]*model.Repo{repo}, nil)

		got, err := newRPCForge(t, impl).Repos(ctx, user, listOpts)
		require.NoError(t, err)
		assert.Equal(t, []*model.Repo{repo}, got)
	})

	t.Run("file and dir", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("File", mock.Anything, user, repo, pipeline, ".woodpecker.yaml").Return([]byte("steps: []"), nil)
		meta := []*types.FileMeta{{Name: ".woodpecker/a.yaml", Data: []byte("a")}}
		impl.On("Dir", mock.Anything, user, repo, pipeline, ".woodpecker").Return(meta, nil)

		f := newRPCForge(t, impl)
		data, err := f.File(ctx, user, repo, pipeline, ".woodpecker.yaml")
		require.NoError(t, err)
		assert.Equal(t, []byte("steps: []"), data)

		dir, err := f.Dir(ctx, user, repo, pipeline, ".woodpecker")
		require.NoError(t, err)
		assert.Equal(t, meta, dir)
	})

	t.Run("status", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		workflow := &model.Workflow{ID: 7, Name: "build", State: model.StatusSuccess}
		impl.On("Status", mock.Anything, user, repo, pipeline, workflow).Return(nil)

		assert.NoError(t, newRPCForge(t, impl).Status(ctx, user, repo, pipeline, workflow))
	})

	t.Run("netrc", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		netrc := &model.Netrc{Machine: "forge.example.com", Login: "octocat", Password: "access", Type: model.ForgeTypeAddon}
		impl.On("Netrc", user, repo).Return(netrc, nil)

		got, err := newRPCForge(t, impl).Netrc(user, repo)
		require.NoError(t, err)
		assert.Equal(t, netrc, got)
	})

	t.Run("activate and deactivate", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Activate", mock.Anything, user, repo, "https://ci.example.com/hook").Return(nil)
		impl.On("Deactivate", mock.Anything, user, repo, "https://ci.example.com/hook").Return(nil)

		f := newRPCForge(t, impl)
		assert.NoError(t, f.Activate(ctx, user, repo, "https://ci.example.com/hook"))
		assert.NoError(t, f.Deactivate(ctx, user, repo, "https://ci.example.com/hook"))
	})

	t.Run("branches, branch head and pull requests", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Branches", mock.Anything, user, repo, listOpts).Return([]string{"main", "dev"}, nil)
		commit := &model.Commit{SHA: "abc", ForgeURL: "https://forge.example.com/commit/abc"}
		impl.On("BranchHead", mock.Anything, user, repo, "main").Return(commit, nil)
		prs := []*model.PullRequest{{Index: "1", Title: "Fix"}}
		impl.On("PullRequests", mock.Anything, user, repo, listOpts).Return(prs, nil)

		f := newRPCForge(t, impl)
		branches, err := f.Branches(ctx, user, repo, listOpts)
		require.NoError(t, err)
		assert.Equal(t, []string{"main", "dev"}, branches)

		head, err := f.BranchHead(ctx, user, repo, "main")
		require.NoError(t, err)
		assert.Equal(t, commit, head)

		gotPRs, err := f.PullRequests(ctx, user, repo, listOpts)
		require.NoError(t, err)
		assert.Equal(t, prs, gotPRs)
	})

	t.Run("org membership and org", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("OrgMembership", mock.Anything, user, "org").Return(&model.OrgPerm{Member: true, Admin: true}, nil)
		impl.On("Org", mock.Anything, user, "org").Return(&model.Org{Name: "org", IsUser: false}, nil)

		f := newRPCForge(t, impl)
		perm, err := f.OrgMembership(ctx, user, "org")
		require.NoError(t, err)
		assert.Equal(t, &model.OrgPerm{Member: true, Admin: true}, perm)

		org, err := f.Org(ctx, user, "org")
		require.NoError(t, err)
		assert.Equal(t, "org", org.Name)
		assert.False(t, org.IsUser)
	})

	t.Run("hook forwards method, url, header, body and secret", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Hook", mock.Anything, mock.Anything, repo.WebhookSecret).Return(func(_ context.Context, r *http.Request, _ string) (*model.Repo, *model.Pipeline, error) {
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "/api/hook?access_token=tok", r.URL.String())
			assert.Equal(t, "push", r.Header.Get("X-Event"))
			assert.JSONEq(t, `{"ref":"refs/heads/main"}`, string(body))
			return repo, pipeline, nil
		})

		req := httptest.NewRequest(http.MethodPost, "/api/hook?access_token=tok", strings.NewReader(`{"ref":"refs/heads/main"}`))
		req.Header.Set("X-Event", "push")

		gotRepo, gotPipeline, err := newRPCForge(t, impl).Hook(ctx, req, repo.WebhookSecret)
		require.NoError(t, err)
		assert.Equal(t, repo, gotRepo)
		assert.Equal(t, pipeline, gotPipeline)
	})

	t.Run("hook without pipeline", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Hook", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, nil)

		req := httptest.NewRequest(http.MethodPost, "/api/hook", strings.NewReader("{}"))
		gotRepo, gotPipeline, err := newRPCForge(t, impl).Hook(ctx, req, "")
		require.NoError(t, err)
		assert.Nil(t, gotRepo)
		assert.Nil(t, gotPipeline)
	})

	t.Run("hook keeps ignored event", func(t *testing.T) {
		ignored := &types.ErrIgnoreEvent{Event: "ping", Reason: "no pipeline event"}
		impl := forge_mocks.NewMockForge(t)
		impl.On("Hook", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, ignored)

		req := httptest.NewRequest(http.MethodPost, "/api/hook", strings.NewReader("{}"))
		gotRepo, gotPipeline, err := newRPCForge(t, impl).Hook(ctx, req, "")
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
		assert.ErrorIs(t, err, &ErrAddon{})
		assert.Equal(t, ignored.Error(), err.Error())
		assert.Nil(t, gotRepo)
		assert.Nil(t, gotPipeline)
	})

	for name, hookErr := range map[string]error{
		"hook keeps invalid signature":         types.ErrInvalidWebhookSignature,
		"hook keeps wrapped invalid signature": fmt.Errorf("verify hook: %w", types.ErrInvalidWebhookSignature),
	} {
		t.Run(name, func(t *testing.T) {
			impl := forge_mocks.NewMockForge(t)
			impl.On("Hook", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, hookErr)

			req := httptest.NewRequest(http.MethodPost, "/api/hook", strings.NewReader("{}"))
			gotRepo, gotPipeline, err := newRPCForge(t, impl).Hook(ctx, req, "")
			assert.ErrorIs(t, err, types.ErrInvalidWebhookSignature)
			assert.ErrorIs(t, err, &ErrAddon{})
			assert.Equal(t, hookErr.Error(), err.Error())
			assert.Nil(t, gotRepo)
			assert.Nil(t, gotPipeline)
		})
	}
}

func TestDecodeErrorKinds(t *testing.T) {
	// an addon built with a newer woodpecker may send kinds this server does not know
	err := decodeErrorKinds("future_kind", "something new")
	assert.EqualError(t, err, "something new")
	assert.ErrorIs(t, err, &ErrAddon{})
	assert.ErrorContains(t, errors.Unwrap(err), "future_kind")
	assert.NotErrorIs(t, err, types.ErrInvalidWebhookSignature)
	assert.NotErrorIs(t, err, &types.ErrIgnoreEvent{})
}

func TestEncodeErrorKinds(t *testing.T) {
	found, kind, msg := encodeErrorKinds(fmt.Errorf("verify hook: %w", types.ErrInvalidWebhookSignature))
	assert.True(t, found)
	assert.Equal(t, errorKindInvalidSignature, kind)
	assert.Equal(t, "verify hook: invalid webhook signature", msg)

	found, kind, msg = encodeErrorKinds(&types.ErrIgnoreEvent{Event: "ping"})
	assert.True(t, found)
	assert.Equal(t, errorKindIgnoreEvent, kind)
	assert.Equal(t, "explicit ignored event 'ping'", msg)

	found, kind, msg = encodeErrorKinds(errors.New("forge unavailable"))
	assert.False(t, found)
	assert.Empty(t, kind)
	assert.Empty(t, msg)
}

func TestErrorKindsOfForgeErrors(t *testing.T) {
	for _, forgeErr := range []error{
		fmt.Errorf("teams: %w", types.ErrNotImplemented),
		errors.Join(errors.New("404"), types.ErrRepoNotFound),
		errors.Join(errors.New("404"), &types.ErrConfigNotFound{Configs: []string{".woodpecker.yaml"}}),
	} {
		t.Run(forgeErr.Error(), func(t *testing.T) {
			base, err := errorResponse(forgeErr)
			require.NoError(t, err)

			assertAddonError(t, forgeErr, base.asError())
		})
	}
}

// assertAddonError checks that got is the error of the addon for want: same
// message and the same errors of errorKindMap match.
func assertAddonError(t *testing.T, want, got error) {
	t.Helper()
	assert.EqualError(t, got, want.Error())
	assert.ErrorIs(t, got, &ErrAddon{})
	for kind, target := range errorKindMap {
		assert.Equal(t, errors.Is(want, target), errors.Is(got, target), kind)
	}
}

func TestErrorResponse(t *testing.T) {
	base, err := errorResponse(nil)
	assert.NoError(t, err)
	assert.Equal(t, baseErrorResponse{}, base)
	assert.NoError(t, base.asError())

	// errors without kind stay errors of the RPC call
	other := errors.New("forge unavailable")
	_, err = errorResponse(other)
	assert.Same(t, other, err)
}

func TestRPCHookAddonError(t *testing.T) {
	ctx := t.Context()

	t.Run("error of the addon", func(t *testing.T) {
		impl := forge_mocks.NewMockForge(t)
		impl.On("Hook", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("parse hook: %w", errors.New("bad payload")))

		req := httptest.NewRequest(http.MethodPost, "/api/hook", strings.NewReader("{}"))
		gotRepo, gotPipeline, err := newRPCForge(t, impl).Hook(ctx, req, "")
		assert.ErrorIs(t, err, &ErrAddon{})
		assert.EqualError(t, err, "parse hook: bad payload")
		assert.Nil(t, gotRepo)
		assert.Nil(t, gotPipeline)
	})

	t.Run("error of the addon system", func(t *testing.T) {
		client, _ := plugin.TestPluginRPCConn(t, map[string]plugin.Plugin{
			pluginKey: &Plugin{Impl: forge_mocks.NewMockForge(t)},
		}, nil)
		raw, err := client.Dispense(pluginKey)
		require.NoError(t, err)
		f, ok := raw.(forge.Forge)
		require.True(t, ok)
		require.NoError(t, client.Close())

		req := httptest.NewRequest(http.MethodPost, "/api/hook", strings.NewReader("{}"))
		_, _, err = f.Hook(ctx, req, "")
		require.Error(t, err)
		assert.NotErrorIs(t, err, &ErrAddon{})
	})
}

func TestRPCErrors(t *testing.T) {
	ctx := t.Context()
	user := testUser()
	repo := testRepo()
	errForge := errors.New("forge unavailable")

	impl := forge_mocks.NewMockForge(t)
	impl.On("Login", mock.Anything, mock.Anything).Return(nil, "", errForge)
	impl.On("Teams", mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Repo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Repos", mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Status", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errForge)
	impl.On("Netrc", mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Activate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errForge)
	impl.On("Deactivate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errForge)
	impl.On("Branches", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("BranchHead", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("PullRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("OrgMembership", mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Org", mock.Anything, mock.Anything, mock.Anything).Return(nil, errForge)
	impl.On("Hook", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, errForge)

	f := newRPCForge(t, impl)

	// the error message is the only thing that crosses the RPC boundary
	assertForgeErr := func(t *testing.T, err error) {
		t.Helper()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "forge unavailable")
	}

	_, _, err := f.Login(ctx, &types.OAuthRequest{})
	assertForgeErr(t, err)
	_, err = f.Teams(ctx, user, &model.ListOptions{})
	assertForgeErr(t, err)
	_, err = f.Repo(ctx, user, "5", "octocat", "hello")
	assertForgeErr(t, err)
	_, err = f.Repos(ctx, user, &model.ListOptions{})
	assertForgeErr(t, err)
	_, err = f.File(ctx, user, repo, &model.Pipeline{}, "f")
	assertForgeErr(t, err)
	_, err = f.Dir(ctx, user, repo, &model.Pipeline{}, "d")
	assertForgeErr(t, err)
	assertForgeErr(t, f.Status(ctx, user, repo, &model.Pipeline{}, &model.Workflow{}))
	_, err = f.Netrc(user, repo)
	assertForgeErr(t, err)
	assertForgeErr(t, f.Activate(ctx, user, repo, "link"))
	assertForgeErr(t, f.Deactivate(ctx, user, repo, "link"))
	_, err = f.Branches(ctx, user, repo, &model.ListOptions{})
	assertForgeErr(t, err)
	_, err = f.BranchHead(ctx, user, repo, "main")
	assertForgeErr(t, err)
	_, err = f.PullRequests(ctx, user, repo, &model.ListOptions{})
	assertForgeErr(t, err)
	_, err = f.OrgMembership(ctx, user, "org")
	assertForgeErr(t, err)
	_, err = f.Org(ctx, user, "org")
	assertForgeErr(t, err)
	_, _, err = f.Hook(ctx, httptest.NewRequest(http.MethodPost, "/api/hook", strings.NewReader("{}")), "")
	assertForgeErr(t, err)
}

func TestRPCErrorKinds(t *testing.T) {
	ctx := t.Context()
	user := testUser()

	for _, tc := range []struct {
		name string
		err  error
		mock func(impl *forge_mocks.MockForge, err error)
		call func(f forge.Forge) error
	}{
		{
			name: "teams not implemented",
			err:  fmt.Errorf("teams: %w", types.ErrNotImplemented),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("Teams", mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.Teams(ctx, user, &model.ListOptions{})
				return err
			},
		},
		{
			name: "repo not found",
			err:  errors.Join(errors.New("404 Not Found"), types.ErrRepoNotFound),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("Repo", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.Repo(ctx, user, "5", "octocat", "hello")
				return err
			},
		},
		{
			name: "file config not found",
			err:  errors.Join(errors.New("404 Not Found"), &types.ErrConfigNotFound{Configs: []string{".woodpecker"}}),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("File", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.File(ctx, user, testRepo(), &model.Pipeline{}, ".woodpecker")
				return err
			},
		},
		{
			name: "dir not implemented",
			err:  fmt.Errorf("dir: %w", types.ErrNotImplemented),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.Dir(ctx, user, testRepo(), &model.Pipeline{}, ".woodpecker")
				return err
			},
		},
		{
			name: "dir config not found",
			err:  errors.Join(errors.New("404 Not Found"), &types.ErrConfigNotFound{Configs: []string{".woodpecker"}}),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("Dir", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.Dir(ctx, user, testRepo(), &model.Pipeline{}, ".woodpecker")
				return err
			},
		},
		{
			name: "deactivate repo not found",
			err:  errors.Join(errors.New("404 Not Found"), types.ErrRepoNotFound),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("Deactivate", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(err)
			},
			call: func(f forge.Forge) error {
				return f.Deactivate(ctx, user, testRepo(), "link")
			},
		},
		{
			name: "branches not implemented",
			err:  fmt.Errorf("branches: %w", types.ErrNotImplemented),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("Branches", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.Branches(ctx, user, testRepo(), &model.ListOptions{})
				return err
			},
		},
		{
			name: "pull requests not implemented",
			err:  fmt.Errorf("pull requests: %w", types.ErrNotImplemented),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("PullRequests", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.PullRequests(ctx, user, testRepo(), &model.ListOptions{})
				return err
			},
		},
		{
			name: "org membership not implemented",
			err:  fmt.Errorf("org membership: %w", types.ErrNotImplemented),
			mock: func(impl *forge_mocks.MockForge, err error) {
				impl.On("OrgMembership", mock.Anything, mock.Anything, mock.Anything).Return(nil, err)
			},
			call: func(f forge.Forge) error {
				_, err := f.OrgMembership(ctx, user, "org")
				return err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			impl := forge_mocks.NewMockForge(t)
			tc.mock(impl, tc.err)

			assertAddonError(t, tc.err, tc.call(newRPCForge(t, impl)))
		})
	}
}

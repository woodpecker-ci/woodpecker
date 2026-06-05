// Copyright 2022 Woodpecker Authors
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

package common_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func Test_Netrc(t *testing.T) {
	host, err := common.ExtractHostFromCloneURL("https://git.example.com/foo/bar.git")
	assert.NoError(t, err)
	assert.Equal(t, "git.example.com", host)
}

func TestExtractHostFromCloneURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		cloneURL string
		want     string
		wantErr  bool
	}{
		{name: "host without port", cloneURL: "https://git.example.com/foo/bar.git", want: "git.example.com"},
		{name: "host with port", cloneURL: "https://git.example.com:8443/foo/bar.git", want: "git.example.com"},
		{name: "invalid url", cloneURL: "://not a url", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			host, err := common.ExtractHostFromCloneURL(tt.cloneURL)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, host)
		})
	}
}

func TestNormalizeEventReason(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"labels_cleared": "label_cleared",
		"labels_updated": "label_updated",
		"labels_added":   "label_added",
		"opened":         "opened", // unknown reason passes through
		"":               "",
	}

	for in, want := range tests {
		assert.Equalf(t, want, common.NormalizeEventReason(in), "input %q", in)
	}
}

func TestGetPipelineStatusDescription(t *testing.T) {
	t.Parallel()

	tests := map[model.StatusValue]string{
		model.StatusPending:        "Pipeline is pending",
		model.StatusRunning:        "Pipeline is running",
		model.StatusSuccess:        "Pipeline was successful",
		model.StatusFailure:        "Pipeline failed",
		model.StatusError:          "Pipeline failed",
		model.StatusKilled:         "Pipeline was canceled",
		model.StatusBlocked:        "Pipeline is pending approval",
		model.StatusDeclined:       "Pipeline was rejected",
		model.StatusValue("bogus"): "unknown status",
	}

	for status, want := range tests {
		assert.Equalf(t, want, common.GetPipelineStatusDescription(status), "status %q", status)
	}
}

func TestGetPipelineStatusURL(t *testing.T) {
	// touches the server.Config global, so must not run in parallel
	orig := server.Config.Server.Host
	defer func() { server.Config.Server.Host = orig }()
	server.Config.Server.Host = "https://ci.example.com"

	repo := &model.Repo{ID: 5}
	pipeline := &model.Pipeline{Number: 42}

	t.Run("without workflow", func(t *testing.T) {
		url := common.GetPipelineStatusURL(repo, pipeline, nil)
		assert.Equal(t, "https://ci.example.com/repos/5/pipeline/42", url)
	})

	t.Run("with workflow", func(t *testing.T) {
		wf := &model.Workflow{PID: 3}
		url := common.GetPipelineStatusURL(repo, pipeline, wf)
		assert.Equal(t, "https://ci.example.com/repos/5/pipeline/42/3", url)
	})
}

func TestUserToken(t *testing.T) {
	t.Parallel()

	t.Run("returns user token directly when user present", func(t *testing.T) {
		t.Parallel()
		u := &model.User{AccessToken: "tok-123"}
		assert.Equal(t, "tok-123", common.UserToken(context.Background(), nil, u))
	})

	t.Run("falls back to repo owner token when user nil", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("GetUser", int64(7)).Return(&model.User{AccessToken: "owner-tok"}, nil)
		ctx := store.InjectToContext(context.Background(), s)

		got := common.UserToken(ctx, &model.Repo{UserID: 7}, nil)
		assert.Equal(t, "owner-tok", got)
	})

	t.Run("returns empty when repo user lookup fails", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("GetUser", int64(7)).Return(nil, errors.New("not found"))
		ctx := store.InjectToContext(context.Background(), s)

		assert.Empty(t, common.UserToken(ctx, &model.Repo{UserID: 7}, nil))
	})
}

func TestRepoUser(t *testing.T) {
	t.Parallel()

	t.Run("no store in context", func(t *testing.T) {
		t.Parallel()
		_, err := common.RepoUser(context.Background(), &model.Repo{UserID: 1})
		assert.Error(t, err)
	})

	t.Run("nil repo", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		ctx := store.InjectToContext(context.Background(), s)
		_, err := common.RepoUser(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("found", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		want := &model.User{ID: 9}
		s.On("GetUser", int64(9)).Return(want, nil)
		ctx := store.InjectToContext(context.Background(), s)

		got, err := common.RepoUser(ctx, &model.Repo{UserID: 9})
		assert.NoError(t, err)
		assert.Equal(t, want, got)
	})
}

func TestRepoUserForgeID(t *testing.T) {
	t.Parallel()

	t.Run("no store in context", func(t *testing.T) {
		t.Parallel()
		_, err := common.RepoUserForgeID(context.Background(), 1, model.ForgeRemoteID("x"))
		assert.Error(t, err)
	})

	t.Run("resolves repo then user", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("GetRepoForgeID", int64(1), model.ForgeRemoteID("rid")).Return(&model.Repo{UserID: 4}, nil)
		s.On("GetUser", int64(4)).Return(&model.User{ID: 4}, nil)
		ctx := store.InjectToContext(context.Background(), s)

		got, err := common.RepoUserForgeID(ctx, 1, model.ForgeRemoteID("rid"))
		assert.NoError(t, err)
		assert.EqualValues(t, 4, got.ID)
	})

	t.Run("repo lookup error", func(t *testing.T) {
		t.Parallel()
		s := store_mocks.NewMockStore(t)
		s.On("GetRepoForgeID", int64(1), model.ForgeRemoteID("rid")).Return(nil, errors.New("nope"))
		ctx := store.InjectToContext(context.Background(), s)

		_, err := common.RepoUserForgeID(ctx, 1, model.ForgeRemoteID("rid"))
		assert.Error(t, err)
	})
}

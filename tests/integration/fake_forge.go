// Copyright 2024 Woodpecker Authors
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

//go:build test
// +build test

package integration

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var (
	forgeLock                = sync.Mutex{}
	currentForge forge.Forge = nil
)

func WithForge(t *testing.T, _forge forge.Forge, fn func()) {
	forgeLock.Lock()
	currentForge = _forge
	defer forgeLock.Unlock()
	fn()
	currentForge = nil
}

type fakeForge struct{}

func (fakeForge) Name() string {
	return currentForge.Name()
}

func (fakeForge) URL() string {
	return currentForge.URL()
}

func (fakeForge) Login(ctx context.Context, r *types.OAuthRequest) (*model.User, string, error) {
	return currentForge.Login(ctx, r)
}

func (fakeForge) Auth(ctx context.Context, token, secret string) (string, error) {
	return currentForge.Auth(ctx, token, secret)
}

func (fakeForge) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	return currentForge.Teams(ctx, u)
}

func (fakeForge) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	return currentForge.Repo(ctx, u, remoteID, owner, name)
}

func (fakeForge) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	return currentForge.Repos(ctx, u)
}

func (fakeForge) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error) {
	return currentForge.File(ctx, u, r, b, f)
}

func (fakeForge) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*types.FileMeta, error) {
	return currentForge.Dir(ctx, u, r, b, f)
}

func (fakeForge) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Workflow) error {
	return currentForge.Status(ctx, u, r, b, p)
}

func (fakeForge) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	return currentForge.Netrc(u, r)
}

func (fakeForge) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	return currentForge.Activate(ctx, u, r, link)
}

func (fakeForge) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	return currentForge.Deactivate(ctx, u, r, link)
}

func (fakeForge) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	return currentForge.Branches(ctx, u, r, p)
}

func (fakeForge) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	return currentForge.BranchHead(ctx, u, r, branch)
}

func (fakeForge) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	return currentForge.PullRequests(ctx, u, r, p)
}

func (fakeForge) Hook(ctx context.Context, r *http.Request) (repo *model.Repo, pipeline *model.Pipeline, err error) {
	return currentForge.Hook(ctx, r)
}

func (fakeForge) OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	return currentForge.OrgMembership(ctx, u, org)
}

func (fakeForge) Org(ctx context.Context, u *model.User, org string) (*model.Org, error) {
	return currentForge.Org(ctx, u, org)
}

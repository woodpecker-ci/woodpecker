// Copyright 2018 Drone.IO Inc.
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

package gerrit

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

// IMPORTANT Gerrit support is not yet implemented. This file is a placeholder
// for future implementation.

// Opts defines configuration options.
type Opts struct {
	URL         string // Gerrit server url.
	Username    string // Optional machine account username.
	Password    string // Optional machine account password.
	PrivateMode bool   // Gerrit is running in private mode.
	SkipVerify  bool   // Skip ssl verification.
}

type client struct {
	URL         string
	Machine     string
	Username    string
	Password    string
	PrivateMode bool
	SkipVerify  bool
}

// New returns a Remote implementation that integrates with Getter, an open
// source Git hosting service and code review system.
func New(opts Opts) (remote.Remote, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err == nil {
		u.Host = host
	}
	return &client{
		URL:         opts.URL,
		Machine:     u.Host,
		Username:    opts.Username,
		Password:    opts.Password,
		PrivateMode: opts.PrivateMode,
		SkipVerify:  opts.SkipVerify,
	}, nil
}

// Login authenticates an account with Gerrit using oauth authentication. The
// Gerrit account details are returned when the user is successfully authenticated.
func (c *client) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	return nil, nil
}

// Auth is not supported by the Gerrit driver.
func (c *client) Auth(ctx context.Context, token, secret string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

// Teams is not supported by the Gerrit driver.
func (c *client) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	var empty []*model.Team
	return empty, nil
}

// Repo is not supported by the Gerrit driver.
func (c *client) Repo(ctx context.Context, u *model.User, owner, name string) (*model.Repo, error) {
	return nil, nil
}

// Repos is not supported by the Gerrit driver.
func (c *client) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	return nil, nil
}

// Perm is not supported by the Gerrit driver.
func (c *client) Perm(ctx context.Context, u *model.User, owner, repo string) (*model.Perm, error) {
	return nil, nil
}

// File is not supported by the Gerrit driver.
func (c *client) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	return nil, nil
}

func (c *client) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]*remote.FileMeta, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Status is not supported by the Gogs driver.
func (c *client) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, link string, proc *model.Proc) error {
	return nil
}

// Netrc is not supported by the Gerrit driver.
func (c *client) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	return nil, nil
}

// Activate is not supported by the Gerrit driver.
func (c *client) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	return nil
}

// Deactivate is not supported by the Gogs driver.
func (c *client) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	return nil
}

// Hook is not supported by the Gerrit driver.
func (c *client) Hook(r *http.Request) (*model.Repo, *model.Build, error) {
	return nil, nil, nil
}

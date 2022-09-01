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

package gogs

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gogits/go-gogs-client"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/common"
)

// Opts defines configuration options.
type Opts struct {
	URL         string // Gogs server url.
	Username    string // Optional machine account username.
	Password    string // Optional machine account password.
	PrivateMode bool   // Gogs is running in private mode.
	SkipVerify  bool   // Skip ssl verification.
}

type client struct {
	URL         string
	Username    string
	Password    string
	PrivateMode bool
	SkipVerify  bool
}

// New returns a Remote implementation that integrates with Gogs, an open
// source Git service written in Go. See https://gogs.io/
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
		Username:    opts.Username,
		Password:    opts.Password,
		PrivateMode: opts.PrivateMode,
		SkipVerify:  opts.SkipVerify,
	}, nil
}

// Name returns the string name of this driver
func (c *client) Name() string {
	return "gogs"
}

// Login authenticates an account with Gogs using basic authentication. The
// Gogs account details are returned when the user is successfully authenticated.
func (c *client) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	var (
		username = req.FormValue("username")
		password = req.FormValue("password")
	)

	// if the username or password is empty we re-direct to the login screen.
	if len(username) == 0 || len(password) == 0 {
		http.Redirect(res, req, "/login/form", http.StatusSeeOther)
		return nil, nil
	}

	client := c.newClient()

	// try to fetch woodpecker token if it exists
	var accessToken string
	tokens, err := client.ListAccessTokens(username, password)
	if err == nil {
		for _, token := range tokens {
			if token.Name == "woodpecker" {
				accessToken = token.Sha1
				break
			}
		}
	}

	// if woodpecker token not found, create it
	if accessToken == "" {
		token, terr := client.CreateAccessToken(
			username,
			password,
			gogs.CreateAccessTokenOption{Name: "woodpecker"},
		)
		if terr != nil {
			return nil, terr
		}
		accessToken = token.Sha1
	}

	client = c.newClientToken(accessToken)
	account, err := client.GetUserInfo(username)
	if err != nil {
		return nil, err
	}

	return &model.User{
		Token:  accessToken,
		Login:  account.UserName,
		Email:  account.Email,
		Avatar: expandAvatar(c.URL, account.AvatarUrl),
	}, nil
}

// Auth is not supported by the Gogs driver.
func (c *client) Auth(ctx context.Context, token, secret string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

// Teams is not supported by the Gogs driver.
func (c *client) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	client := c.newClientToken(u.Token)
	orgs, err := client.ListMyOrgs()
	if err != nil {
		return nil, err
	}

	var teams []*model.Team
	for _, org := range orgs {
		teams = append(teams, toTeam(org, c.URL))
	}
	return teams, nil
}

// Repo returns the named Gogs repository.
func (c *client) Repo(ctx context.Context, u *model.User, id, owner, name string) (*model.Repo, error) {
	client := c.newClientToken(u.Token)
	repo, err := client.GetRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return toRepo(repo, c.PrivateMode), nil
}

// Repos returns a list of all repositories for the Gogs account, including
// organization repositories.
func (c *client) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	var repos []*model.Repo

	client := c.newClientToken(u.Token)
	all, err := client.ListMyRepos()
	if err != nil {
		return repos, err
	}

	for _, repo := range all {
		repos = append(repos, toRepo(repo, c.PrivateMode))
	}
	return repos, err
}

// Perm returns the user permissions for the named Gogs repository.
func (c *client) Perm(ctx context.Context, u *model.User, r *model.Repo) (*model.Perm, error) {
	client := c.newClientToken(u.Token)
	repo, err := client.GetRepo(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}
	return toPerm(repo.Permissions), nil
}

// File fetches the file from the Gogs repository and returns its contents.
func (c *client) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	client := c.newClientToken(u.Token)
	ref := b.Commit

	// TODO gogs does not yet return a sha with the pull request
	// so unfortunately we need to use the pull request branch.
	if b.Event == model.EventPull {
		ref = b.Branch
	}
	if ref == "" {
		// Remove refs/tags or refs/heads, Gogs needs a short ref
		ref = strings.TrimPrefix(
			strings.TrimPrefix(
				b.Ref,
				"refs/heads/",
			),
			"refs/tags/",
		)
	}
	cfg, err := client.GetFile(r.Owner, r.Name, ref, f)
	return cfg, err
}

func (c *client) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]*remote.FileMeta, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Status is not supported by the Gogs driver.
func (c *client) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, proc *model.Proc) error {
	return nil
}

// Netrc returns a netrc file capable of authenticating Gogs requests and
// cloning Gogs repositories. The netrc will use the global machine account
// when configured.
func (c *client) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	host, err := common.ExtractHostFromCloneURL(r.Clone)
	if err != nil {
		return nil, err
	}

	if c.Password != "" {
		return &model.Netrc{
			Login:    c.Username,
			Password: c.Password,
			Machine:  host,
		}, nil
	}
	return &model.Netrc{
		Login:    u.Token,
		Password: "x-oauth-basic",
		Machine:  host,
	}, nil
}

// Activate activates the repository by registering post-commit hooks with
// the Gogs repository.
func (c *client) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	config := map[string]string{
		"url":          link,
		"secret":       r.Hash,
		"content_type": "json",
	}
	hook := gogs.CreateHookOption{
		Type:   "gogs",
		Config: config,
		Events: []string{"push", "create", "pull_request"},
		Active: true,
	}

	client := c.newClientToken(u.Token)
	_, err := client.CreateRepoHook(r.Owner, r.Name, hook)
	return err
}

// Deactivate is not supported by the Gogs driver.
func (c *client) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	return nil
}

// Branches returns the names of all branches for the named repository.
func (c *client) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	token := ""
	if u != nil {
		token = u.Token
	}
	client := c.newClientToken(token)
	gogsBranches, err := client.ListRepoBranches(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range gogsBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// BranchHead returns sha of commit on top of the specified branch
func (c *client) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error) {
	token := ""
	if u != nil {
		token = u.Token
	}
	b, err := c.newClientToken(token).GetRepoBranch(r.Owner, r.Name, branch)
	if err != nil {
		return "", err
	}
	return b.Commit.ID, nil
}

// Hook parses the incoming Gogs hook and returns the Repository and Build
// details. If the hook is unsupported nil values are returned.
func (c *client) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Build, error) {
	return parseHook(r)
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *client) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client := c.newClientToken(u.Token)

	orgs, err := client.ListMyOrgs()
	if err != nil {
		return nil, err
	}

	for _, org := range orgs {
		if org.UserName == owner {
			// TODO: API does not support checking if user is admin/owner of org
			return &model.OrgPerm{Member: true}, nil
		}
	}
	return &model.OrgPerm{}, nil
}

// helper function to return the Gogs client
func (c *client) newClient() *gogs.Client {
	return c.newClientToken("")
}

// helper function to return the Gogs client
func (c *client) newClientToken(token string) *gogs.Client {
	client := gogs.NewClient(c.URL, token)
	if c.SkipVerify {
		httpClient := &http.Client{}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client.SetHTTPClient(httpClient)
	}
	return client
}

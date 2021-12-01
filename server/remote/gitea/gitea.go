// Copyright 2018 Drone.IO Inc.
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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
//
// This file has been modified by Informatyka Boguslawski sp. z o.o. sp.k.

package gitea

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"path/filepath"

	"code.gitea.io/sdk/gitea"
	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

const (
	authorizeTokenURL = "%s/login/oauth/authorize"
	accessTokenURL    = "%s/login/oauth/access_token"
	perPage           = 50
)

type Gitea struct {
	URL          string
	Context      string
	Machine      string
	ClientID     string
	ClientSecret string
	Username     string
	Password     string
	PrivateMode  bool
	SkipVerify   bool
}

// Opts defines configuration options.
type Opts struct {
	URL         string // Gitea server url.
	Context     string // Context to display in status check
	Client      string // OAuth2 Client ID
	Secret      string // OAuth2 Client Secret
	Username    string // Optional machine account username.
	Password    string // Optional machine account password.
	PrivateMode bool   // Gitea is running in private mode.
	SkipVerify  bool   // Skip ssl verification.
}

// New returns a Remote implementation that integrates with Gitea,
// an open source Git service written in Go. See https://gitea.io/
func New(opts Opts) (remote.Remote, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err == nil {
		u.Host = host
	}
	return &Gitea{
		URL:          opts.URL,
		Context:      opts.Context,
		Machine:      u.Host,
		ClientID:     opts.Client,
		ClientSecret: opts.Secret,
		Username:     opts.Username,
		Password:     opts.Password,
		PrivateMode:  opts.PrivateMode,
		SkipVerify:   opts.SkipVerify,
	}, nil
}

// Login authenticates an account with Gitea using basic authentication. The
// Gitea account details are returned when the user is successfully authenticated.
func (c *Gitea) Login(ctx context.Context, w http.ResponseWriter, req *http.Request) (*model.User, error) {
	config := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf(authorizeTokenURL, c.URL),
			TokenURL: fmt.Sprintf(accessTokenURL, c.URL),
		},
		RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.Host),
	}

	// get the OAuth errors
	if err := req.FormValue("error"); err != "" {
		return nil, &remote.AuthError{
			Err:         err,
			Description: req.FormValue("error_description"),
			URI:         req.FormValue("error_uri"),
		}
	}

	// get the OAuth code
	code := req.FormValue("code")
	if len(code) == 0 {
		http.Redirect(w, req, config.AuthCodeURL("woodpecker"), http.StatusSeeOther)
		return nil, nil
	}

	token, err := config.Exchange(ctx, code)
	if err != nil {
		return nil, err
	}

	client, err := c.newClientToken(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	account, _, err := client.GetMyUserInfo()
	if err != nil {
		return nil, err
	}

	return &model.User{
		Token:  token.AccessToken,
		Secret: token.RefreshToken,
		Expiry: token.Expiry.UTC().Unix(),
		Login:  account.UserName,
		Email:  account.Email,
		Avatar: expandAvatar(c.URL, account.AvatarURL),
	}, nil
}

// Auth uses the Gitea oauth2 access token and refresh token to authenticate
// a session and return the Gitea account login.
func (c *Gitea) Auth(ctx context.Context, token, secret string) (string, error) {
	client, err := c.newClientToken(ctx, token)
	if err != nil {
		return "", err
	}
	user, _, err := client.GetMyUserInfo()
	if err != nil {
		return "", err
	}
	return user.UserName, nil
}

// Refresh refreshes the Gitea oauth2 access token. If the token is
// refreshed the user is updated and a true value is returned.
func (c *Gitea) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config := &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf(authorizeTokenURL, c.URL),
			TokenURL: fmt.Sprintf(accessTokenURL, c.URL),
		},
	}
	source := config.TokenSource(ctx, &oauth2.Token{RefreshToken: user.Secret})

	token, err := source.Token()
	if err != nil || len(token.AccessToken) == 0 {
		return false, err
	}

	user.Token = token.AccessToken
	user.Secret = token.RefreshToken
	user.Expiry = token.Expiry.UTC().Unix()
	return true, nil
}

// Teams is supported by the Gitea driver.
func (c *Gitea) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	teams := make([]*model.Team, 0, perPage)

	page := 1
	for {
		orgs, _, err := client.ListMyOrgs(
			gitea.ListOrgsOptions{
				ListOptions: gitea.ListOptions{
					Page:     page,
					PageSize: perPage,
				},
			},
		)
		if err != nil {
			return nil, err
		}

		for _, org := range orgs {
			teams = append(teams, toTeam(org, c.URL))
		}

		if len(orgs) < perPage {
			break
		}
		page++
	}

	return teams, nil
}

// TeamPerm is not supported by the Gitea driver.
func (c *Gitea) TeamPerm(u *model.User, org string) (*model.Perm, error) {
	return nil, nil
}

// Repo returns the named Gitea repository.
func (c *Gitea) Repo(ctx context.Context, u *model.User, owner, name string) (*model.Repo, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	repo, _, err := client.GetRepo(owner, name)
	if err != nil {
		return nil, err
	}
	if c.PrivateMode {
		repo.Private = true
	}
	return toRepo(repo, c.PrivateMode), nil
}

// Repos returns a list of all repositories for the Gitea account, including
// organization repositories.
func (c *Gitea) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	repos := make([]*model.Repo, 0, perPage)

	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	// Gitea SDK forces us to read repo list paginated.
	var page = 1
	for {
		all, _, err := client.ListMyRepos(
			gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{
					Page:     page,
					PageSize: perPage,
				},
			},
		)
		if err != nil {
			return nil, err
		}

		for _, repo := range all {
			repos = append(repos, toRepo(repo, c.PrivateMode))
		}

		if len(all) < perPage {
			break
		}
		// Last page was not empty so more repos may be available - continue loop.
		page++
	}

	return repos, nil
}

// Perm returns the user permissions for the named Gitea repository.
func (c *Gitea) Perm(ctx context.Context, u *model.User, owner, name string) (*model.Perm, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	repo, _, err := client.GetRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return toPerm(repo.Permissions), nil
}

// File fetches the file from the Gitea repository and returns its contents.
func (c *Gitea) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	cfg, _, err := client.GetFile(r.Owner, r.Name, b.Commit, f)
	return cfg, err
}

func (c *Gitea) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]*remote.FileMeta, error) {
	var configs []*remote.FileMeta

	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	// List files in repository. Path from root
	tree, _, err := client.GetTrees(r.Owner, r.Name, b.Commit, true)
	if err != nil {
		return nil, err
	}

	f = path.Clean(f) // We clean path and remove trailing slash
	f += "/" + "*"    // construct pattern for match i.e. file in subdir
	for _, e := range tree.Entries {
		// Filter path matching pattern and type file (blob)
		if m, _ := filepath.Match(f, e.Path); m && e.Type == "blob" {
			data, err := c.File(ctx, u, r, b, e.Path)
			if err != nil {
				return nil, fmt.Errorf("multi-pipeline cannot get %s: %s", e.Path, err)
			}

			configs = append(configs, &remote.FileMeta{
				Name: e.Path,
				Data: data,
			})
		}
	}

	return configs, nil
}

// Status is supported by the Gitea driver.
func (c *Gitea) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, link string, proc *model.Proc) error {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}

	context := c.Context

	switch b.Event {
	case model.EventPull:
		context += "/pr"
	default:
		if len(b.Event) > 0 {
			context += "/" + b.Event
		}
	}

	status := getStatus(b.Status)
	desc := getDesc(b.Status)

	if proc != nil {
		context += "/" + proc.Name
		status = getStatus(proc.State)
		desc = getDesc(proc.State)
	}

	_, _, err = client.CreateStatus(
		r.Owner,
		r.Name,
		b.Commit,
		gitea.CreateStatusOption{
			State:       status,
			TargetURL:   link,
			Description: desc,
			Context:     context,
		},
	)

	return err
}

// Netrc returns a netrc file capable of authenticating Gitea requests and
// cloning Gitea repositories. The netrc will use the global machine account
// when configured.
func (c *Gitea) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	if c.Password != "" {
		return &model.Netrc{
			Login:    c.Username,
			Password: c.Password,
			Machine:  c.Machine,
		}, nil
	}
	return &model.Netrc{
		Login:    u.Login,
		Password: u.Token,
		Machine:  c.Machine,
	}, nil
}

// Activate activates the repository by registering post-commit hooks with
// the Gitea repository.
func (c *Gitea) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	config := map[string]string{
		"url":          link,
		"secret":       r.Hash,
		"content_type": "json",
	}
	hook := gitea.CreateHookOption{
		Type:   gitea.HookTypeGitea,
		Config: config,
		Events: []string{"push", "create", "pull_request"},
		Active: true,
	}

	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}
	_, _, err = client.CreateRepoHook(r.Owner, r.Name, hook)
	return err
}

// Deactivate deactives the repository be removing repository push hooks from
// the Gitea repository.
func (c *Gitea) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}

	hooks, _, err := client.ListRepoHooks(r.Owner, r.Name, gitea.ListHooksOptions{})
	if err != nil {
		return err
	}

	hook := matchingHooks(hooks, link)
	if hook != nil {
		_, err := client.DeleteRepoHook(r.Owner, r.Name, hook.ID)
		return err
	}

	return nil
}

// Branches returns the names of all branches for the named repository.
func (c *Gitea) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	giteaBranches, _, err := client.ListRepoBranches(r.Owner, r.Name, gitea.ListRepoBranchesOptions{})
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range giteaBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// Hook parses the incoming Gitea hook and returns the Repository and Build
// details. If the hook is unsupported nil values are returned.
func (c *Gitea) Hook(r *http.Request) (*model.Repo, *model.Build, error) {
	return parseHook(r)
}

// helper function to return the Gitea client with Token
func (c *Gitea) newClientToken(ctx context.Context, token string) (*gitea.Client, error) {
	httpClient := &http.Client{}
	if c.SkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	return gitea.NewClient(c.URL, gitea.SetToken(token), gitea.SetHTTPClient(httpClient), gitea.SetContext(ctx))
}

const (
	DescPending  = "the build is pending"
	DescRunning  = "the build is running"
	DescSuccess  = "the build was successful"
	DescFailure  = "the build failed"
	DescCanceled = "the build canceled"
	DescBlocked  = "the build is pending approval"
	DescDeclined = "the build was rejected"
)

// getStatus is a helper function that converts a Woodpecker
// status to a Gitea status.
func getStatus(status model.StatusValue) gitea.StatusState {
	switch status {
	case model.StatusPending, model.StatusBlocked:
		return gitea.StatusPending
	case model.StatusRunning:
		return gitea.StatusPending
	case model.StatusSuccess:
		return gitea.StatusSuccess
	case model.StatusFailure, model.StatusError:
		return gitea.StatusFailure
	case model.StatusKilled:
		return gitea.StatusFailure
	case model.StatusDeclined:
		return gitea.StatusWarning
	default:
		return gitea.StatusFailure
	}
}

// getDesc is a helper function that generates a description
// message for the build based on the status.
func getDesc(status model.StatusValue) string {
	switch status {
	case model.StatusPending:
		return DescPending
	case model.StatusRunning:
		return DescRunning
	case model.StatusSuccess:
		return DescSuccess
	case model.StatusFailure, model.StatusError:
		return DescFailure
	case model.StatusKilled:
		return DescCanceled
	case model.StatusBlocked:
		return DescBlocked
	case model.StatusDeclined:
		return DescDeclined
	default:
		return DescFailure
	}
}

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

package bitbucket

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/bitbucket/internal"
	"github.com/woodpecker-ci/woodpecker/server/remote/common"
)

// Bitbucket cloud endpoints.
const (
	DefaultAPI = "https://api.bitbucket.org"
	DefaultURL = "https://bitbucket.org"
)

// Opts are remote options for bitbucket
type Opts struct {
	Client string
	Secret string
}

type config struct {
	API    string
	URL    string
	Client string
	Secret string
}

// New returns a new remote Configuration for integrating with the Bitbucket
// repository hosting service at https://bitbucket.org
func New(opts *Opts) (remote.Remote, error) {
	return &config{
		API:    DefaultAPI,
		URL:    DefaultURL,
		Client: opts.Client,
		Secret: opts.Secret,
	}, nil
	// TODO: add checks
}

// Name returns the string name of this driver
func (c *config) Name() string {
	return "bitbucket"
}

// Login authenticates an account with Bitbucket using the oauth2 protocol. The
// Bitbucket account details are returned when the user is successfully authenticated.
func (c *config) Login(ctx context.Context, w http.ResponseWriter, req *http.Request) (*model.User, error) {
	config := c.newConfig(server.Config.Server.Host)

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

	client := internal.NewClient(ctx, c.API, config.Client(ctx, token))
	curr, err := client.FindCurrent()
	if err != nil {
		return nil, err
	}
	return convertUser(curr, token), nil
}

// Auth uses the Bitbucket oauth2 access token and refresh token to authenticate
// a session and return the Bitbucket account login.
func (c *config) Auth(ctx context.Context, token, secret string) (string, error) {
	client := c.newClientToken(ctx, token, secret)
	user, err := client.FindCurrent()
	if err != nil {
		return "", err
	}
	return user.Login, nil
}

// Refresh refreshes the Bitbucket oauth2 access token. If the token is
// refreshed the user is updated and a true value is returned.
func (c *config) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config := c.newConfig("")
	source := config.TokenSource(
		ctx, &oauth2.Token{RefreshToken: user.Secret})

	token, err := source.Token()
	if err != nil || len(token.AccessToken) == 0 {
		return false, err
	}

	user.Token = token.AccessToken
	user.Secret = token.RefreshToken
	user.Expiry = token.Expiry.UTC().Unix()
	return true, nil
}

// Teams returns a list of all team membership for the Bitbucket account.
func (c *config) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	opts := &internal.ListTeamOpts{
		PageLen: 100,
		Role:    "member",
	}
	resp, err := c.newClient(ctx, u).ListTeams(opts)
	if err != nil {
		return nil, err
	}
	return convertTeamList(resp.Values), nil
}

// Repo returns the named Bitbucket repository.
func (c *config) Repo(ctx context.Context, u *model.User, id, owner, name string) (*model.Repo, error) {
	repo, err := c.newClient(ctx, u).FindRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo), nil
}

// Repos returns a list of all repositories for Bitbucket account, including
// organization repositories.
func (c *config) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client := c.newClient(ctx, u)

	var all []*model.Repo

	accounts := []string{u.Login}
	resp, err := client.ListTeams(&internal.ListTeamOpts{
		PageLen: 100,
		Role:    "member",
	})
	if err != nil {
		return all, err
	}
	for _, team := range resp.Values {
		accounts = append(accounts, team.Login)
	}

	for _, account := range accounts {
		repos, err := client.ListReposAll(account)
		if err != nil {
			return all, err
		}
		for _, repo := range repos {
			all = append(all, convertRepo(repo))
		}
	}
	return all, nil
}

// Perm returns the user permissions for the named repository. Because Bitbucket
// does not have an endpoint to access user permissions, we attempt to fetch
// the repository hook list, which is restricted to administrators to calculate
// administrative access to a repository.
func (c *config) Perm(ctx context.Context, u *model.User, r *model.Repo) (*model.Perm, error) {
	client := c.newClient(ctx, u)

	perms := new(model.Perm)
	repo, err := client.FindRepo(r.Owner, r.Name)
	if err != nil {
		return perms, err
	}

	perm, err := client.GetPermission(repo.FullName)
	if err != nil {
		return perms, err
	}

	switch perm.Permission {
	case "admin":
		perms.Admin = true
		fallthrough
	case "write":
		perms.Push = true
		fallthrough
	default:
		perms.Pull = true
	}

	return perms, nil
}

// File fetches the file from the Bitbucket repository and returns its contents.
func (c *config) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	config, err := c.newClient(ctx, u).FindSource(r.Owner, r.Name, b.Commit, f)
	if err != nil {
		return nil, err
	}
	return []byte(*config), err
}

func (c *config) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]*remote.FileMeta, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Status creates a build status for the Bitbucket commit.
func (c *config) Status(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, proc *model.Proc) error {
	status := internal.BuildStatus{
		State: convertStatus(build.Status),
		Desc:  common.GetBuildStatusDescription(build.Status),
		Key:   "Woodpecker",
		URL:   common.GetBuildStatusLink(repo, build, nil),
	}
	return c.newClient(ctx, user).CreateStatus(repo.Owner, repo.Name, build.Commit, &status)
}

// Activate activates the repository by registering repository push hooks with
// the Bitbucket repository. Prior to registering hook, previously created hooks
// are deleted.
func (c *config) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	rawurl, err := url.Parse(link)
	if err != nil {
		return err
	}
	_ = c.Deactivate(ctx, u, r, link)

	return c.newClient(ctx, u).CreateHook(r.Owner, r.Name, &internal.Hook{
		Active: true,
		Desc:   rawurl.Host,
		Events: []string{"repo:push"},
		URL:    link,
	})
}

// Deactivate deactives the repository be removing repository push hooks from
// the Bitbucket repository.
func (c *config) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := c.newClient(ctx, u)

	hooks, err := client.ListHooks(r.Owner, r.Name, &internal.ListOpts{})
	if err != nil {
		return err
	}
	hook := matchingHooks(hooks.Values, link)
	if hook != nil {
		return client.DeleteHook(r.Owner, r.Name, hook.UUID)
	}
	return nil
}

// Netrc returns a netrc file capable of authenticating Bitbucket requests and
// cloning Bitbucket repositories.
func (c *config) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	return &model.Netrc{
		Machine:  "bitbucket.org",
		Login:    "x-token-auth",
		Password: u.Token,
	}, nil
}

// Branches returns the names of all branches for the named repository.
func (c *config) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	bitbucketBranches, err := c.newClient(ctx, u).ListBranches(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range bitbucketBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// Hook parses the incoming Bitbucket hook and returns the Repository and
// Build details. If the hook is unsupported nil values are returned.
func (c *config) Hook(ctx context.Context, req *http.Request) (*model.Repo, *model.Build, error) {
	return parseHook(req)
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *config) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	perm, err := c.newClient(ctx, u).GetUserWorkspaceMembership(owner, u.Login)
	if err != nil {
		return nil, err
	}
	return &model.OrgPerm{Member: perm != "", Admin: perm == "owner"}, nil
}

// helper function to return the bitbucket oauth2 client
func (c *config) newClient(ctx context.Context, u *model.User) *internal.Client {
	if u == nil {
		return c.newClientToken(ctx, "", "")
	}
	return c.newClientToken(ctx, u.Token, u.Secret)
}

// helper function to return the bitbucket oauth2 client
func (c *config) newClientToken(ctx context.Context, token, secret string) *internal.Client {
	return internal.NewClientToken(
		ctx,
		c.API,
		c.Client,
		c.Secret,
		&oauth2.Token{
			AccessToken:  token,
			RefreshToken: secret,
		},
	)
}

// helper function to return the bitbucket oauth2 config
func (c *config) newConfig(redirect string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.Client,
		ClientSecret: c.Secret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/site/oauth2/authorize", c.URL),
			TokenURL: fmt.Sprintf("%s/site/oauth2/access_token", c.URL),
		},
		RedirectURL: fmt.Sprintf("%s/authorize", redirect),
	}
}

// helper function to return matching hooks.
func matchingHooks(hooks []*internal.Hook, rawurl string) *internal.Hook {
	link, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		hookurl, err := url.Parse(hook.URL)
		if err == nil && hookurl.Host == link.Host {
			return hook
		}
	}
	return nil
}

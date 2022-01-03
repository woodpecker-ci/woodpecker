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

package github

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v39/github"
	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/common"
)

const (
	defaultURL = "https://github.com"      // Default GitHub URL
	defaultAPI = "https://api.github.com/" // Default GitHub API URL
)

// Opts defines configuration options.
type Opts struct {
	URL         string   // GitHub server url.
	Context     string   // Context to display in status check
	Client      string   // GitHub oauth client id.
	Secret      string   // GitHub oauth client secret.
	Scopes      []string // GitHub oauth scopes
	Username    string   // Optional machine account username.
	Password    string   // Optional machine account password.
	PrivateMode bool     // GitHub is running in private mode.
	SkipVerify  bool     // Skip ssl verification.
	MergeRef    bool     // Clone pull requests using the merge ref.
}

// New returns a Remote implementation that integrates with a GitHub Cloud or
// GitHub Enterprise version control hosting provider.
func New(opts Opts) (remote.Remote, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err == nil {
		u.Host = host
	}
	r := &client{
		API:         defaultAPI,
		URL:         defaultURL,
		Context:     opts.Context,
		Client:      opts.Client,
		Secret:      opts.Secret,
		Scopes:      opts.Scopes,
		PrivateMode: opts.PrivateMode,
		SkipVerify:  opts.SkipVerify,
		MergeRef:    opts.MergeRef,
		Machine:     u.Host,
		Username:    opts.Username,
		Password:    opts.Password,
	}
	if opts.URL != defaultURL {
		r.URL = strings.TrimSuffix(opts.URL, "/")
		r.API = r.URL + "/api/v3/"
	}

	return r, nil
}

type client struct {
	URL         string
	Context     string
	API         string
	Client      string
	Secret      string
	Scopes      []string
	Machine     string
	Username    string
	Password    string
	PrivateMode bool
	SkipVerify  bool
	MergeRef    bool
}

// Login authenticates the session and returns the remote user details.
func (c *client) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	config := c.newConfig(req)

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
		// TODO(bradrydzewski) we really should be using a random value here and
		// storing in a cookie for verification in the next stage of the workflow.

		http.Redirect(res, req, config.AuthCodeURL("woodpecker"), http.StatusSeeOther)
		return nil, nil
	}

	token, err := config.Exchange(c.newContext(ctx), code)
	if err != nil {
		return nil, err
	}

	client := c.newClientToken(ctx, token.AccessToken)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, err
	}

	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return nil, err
	}
	email := matchingEmail(emails, c.API)
	if email == nil {
		return nil, fmt.Errorf("No verified Email address for GitHub account")
	}

	return &model.User{
		Login:  *user.Login,
		Email:  *email.Email,
		Token:  token.AccessToken,
		Avatar: *user.AvatarURL,
	}, nil
}

// Auth returns the GitHub user login for the given access token.
func (c *client) Auth(ctx context.Context, token, secret string) (string, error) {
	client := c.newClientToken(ctx, token)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return "", err
	}
	return *user.Login, nil
}

// Teams returns a list of all team membership for the GitHub account.
func (c *client) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.ListOptions)
	opts.Page = 1

	var teams []*model.Team
	for opts.Page > 0 {
		list, resp, err := client.Organizations.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}
		teams = append(teams, convertTeamList(list)...)
		opts.Page = resp.NextPage
	}
	return teams, nil
}

// Repo returns the named GitHub repository.
func (c *client) Repo(ctx context.Context, u *model.User, owner, name string) (*model.Repo, error) {
	client := c.newClientToken(ctx, u.Token)
	repo, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo, c.PrivateMode), nil
}

// Repos returns a list of all repositories for GitHub account, including
// organization repositories.
func (c *client) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.RepositoryListOptions)
	opts.PerPage = 100
	opts.Page = 1

	var repos []*model.Repo
	for opts.Page > 0 {
		list, resp, err := client.Repositories.List(ctx, "", opts)
		if err != nil {
			return nil, err
		}
		repos = append(repos, convertRepoList(list, c.PrivateMode)...)
		opts.Page = resp.NextPage
	}
	return repos, nil
}

// Perm returns the user permissions for the named GitHub repository.
func (c *client) Perm(ctx context.Context, u *model.User, r *model.Repo) (*model.Perm, error) {
	client := c.newClientToken(ctx, u.Token)
	repo, _, err := client.Repositories.Get(ctx, r.Owner, r.Name)
	if err != nil {
		return nil, err
	}
	return convertPerm(repo), nil
}

// File fetches the file from the GitHub repository and returns its contents.
func (c *client) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.RepositoryContentGetOptions)
	opts.Ref = b.Commit
	content, _, _, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, f, opts)
	if err != nil {
		return nil, err
	}
	if content == nil {
		return nil, fmt.Errorf("%s is a folder not a file use Dir(..)", f)
	}
	data, err := content.GetContent()
	return []byte(data), err
}

func (c *client) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]*remote.FileMeta, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.RepositoryContentGetOptions)
	opts.Ref = b.Commit
	_, data, _, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, f, opts)
	if err != nil {
		return nil, err
	}

	fc := make(chan *remote.FileMeta)
	errc := make(chan error)

	for _, file := range data {
		go func(path string) {
			content, err := c.File(ctx, u, r, b, path)
			if err != nil {
				errc <- err
			} else {
				fc <- &remote.FileMeta{
					Name: path,
					Data: content,
				}
			}
		}(f + "/" + *file.Name)
	}

	var files []*remote.FileMeta

	for i := 0; i < len(data); i++ {
		select {
		case err := <-errc:
			return nil, err
		case fileMeta := <-fc:
			files = append(files, fileMeta)
		}
	}

	close(fc)
	close(errc)

	return files, nil
}

// Netrc returns a netrc file capable of authenticating GitHub requests and
// cloning GitHub repositories. The netrc will use the global machine account
// when configured.
func (c *client) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	if c.Password != "" {
		return &model.Netrc{
			Login:    c.Username,
			Password: c.Password,
			Machine:  c.Machine,
		}, nil
	}
	return &model.Netrc{
		Login:    u.Token,
		Password: "x-oauth-basic",
		Machine:  c.Machine,
	}, nil
}

// Deactivate deactives the repository be removing registered push hooks from
// the GitHub repository.
func (c *client) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := c.newClientToken(ctx, u.Token)
	hooks, _, err := client.Repositories.ListHooks(ctx, r.Owner, r.Name, nil)
	if err != nil {
		return err
	}
	match := matchingHooks(hooks, link)
	if match == nil {
		return nil
	}
	_, err = client.Repositories.DeleteHook(ctx, r.Owner, r.Name, *match.ID)
	return err
}

// helper function to return the GitHub oauth2 context using an HTTPClient that
// disables TLS verification if disabled in the remote settings.
func (c *client) newContext(ctx context.Context) context.Context {
	if !c.SkipVerify {
		return ctx
	}
	return context.WithValue(ctx, oauth2.HTTPClient, &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})
}

// helper function to return the GitHub oauth2 config
func (c *client) newConfig(req *http.Request) *oauth2.Config {
	var redirect string

	intendedURL := req.URL.Query()["url"]
	if len(intendedURL) > 0 {
		redirect = fmt.Sprintf("%s/authorize?url=%s", server.Config.Server.OAuthHost, intendedURL[0])
	} else {
		redirect = fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost)
	}

	return &oauth2.Config{
		ClientID:     c.Client,
		ClientSecret: c.Secret,
		Scopes:       c.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", c.URL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", c.URL),
		},
		RedirectURL: redirect,
	}
}

// helper function to return the GitHub oauth2 client
func (c *client) newClientToken(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	if c.SkipVerify {
		tc.Transport.(*oauth2.Transport).Base = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	client := github.NewClient(tc)
	client.BaseURL, _ = url.Parse(c.API)
	return client
}

// helper function to return matching user email.
func matchingEmail(emails []*github.UserEmail, rawURL string) *github.UserEmail {
	for _, email := range emails {
		if email.Email == nil || email.Primary == nil || email.Verified == nil {
			continue
		}
		if *email.Primary && *email.Verified {
			return email
		}
	}
	// github enterprise does not support verified email addresses so instead
	// we'll return the first email address in the list.
	if len(emails) != 0 && rawURL != defaultAPI {
		return emails[0]
	}
	return nil
}

// helper function to return matching hook.
func matchingHooks(hooks []*github.Hook, rawurl string) *github.Hook {
	link, err := url.Parse(rawurl)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		if hook.ID == nil {
			continue
		}
		v, ok := hook.Config["url"]
		if !ok {
			continue
		}
		s, ok := v.(string)
		if !ok {
			continue
		}
		hookURL, err := url.Parse(s)
		if err == nil && hookURL.Host == link.Host {
			return hook
		}
	}
	return nil
}

var reDeploy = regexp.MustCompile(`.+/deployments/(\d+)`)

// Status sends the commit status to the remote system.
// An example would be the GitHub pull request status.
func (c *client) Status(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, proc *model.Proc) error {
	client := c.newClientToken(ctx, user.Token)

	if build.Event == model.EventDeploy {
		matches := reDeploy.FindStringSubmatch(build.Link)
		if len(matches) != 2 {
			return nil
		}
		id, _ := strconv.Atoi(matches[1])

		_, _, err := client.Repositories.CreateDeploymentStatus(ctx, repo.Owner, repo.Name, int64(id), &github.DeploymentStatusRequest{
			State:       github.String(convertStatus(build.Status)),
			Description: github.String(common.GetBuildStatusDescription(build.Status)),
			LogURL:      github.String(common.GetBuildStatusLink(repo, build, nil)),
		})
		return err
	}

	_, _, err := client.Repositories.CreateStatus(ctx, repo.Owner, repo.Name, build.Commit, &github.RepoStatus{
		Context:     github.String(common.GetBuildStatusContext(repo, build, proc)),
		State:       github.String(convertStatus(proc.State)),
		Description: github.String(common.GetBuildStatusDescription(proc.State)),
		TargetURL:   github.String(common.GetBuildStatusLink(repo, build, proc)),
	})
	return err
}

// Activate activates a repository by creating the post-commit hook and
// adding the SSH deploy key, if applicable.
func (c *client) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	if err := c.Deactivate(ctx, u, r, link); err != nil {
		return err
	}
	client := c.newClientToken(ctx, u.Token)
	hook := &github.Hook{
		Name: github.String("web"),
		Events: []string{
			"push",
			"pull_request",
			"deployment",
		},
		Config: map[string]interface{}{
			"url":          link,
			"content_type": "form",
		},
	}
	_, _, err := client.Repositories.CreateHook(ctx, r.Owner, r.Name, hook)
	return err
}

// Branches returns the names of all branches for the named repository.
func (c *client) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	client := c.newClientToken(ctx, u.Token)

	githubBranches, _, err := client.Repositories.ListBranches(ctx, r.Owner, r.Name, &github.BranchListOptions{})
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range githubBranches {
		branches = append(branches, *branch.Name)
	}
	return branches, nil
}

// Hook parses the post-commit hook from the Request body
// and returns the required data in a standard format.
func (c *client) Hook(r *http.Request) (*model.Repo, *model.Build, error) {
	return parseHook(r, c.MergeRef)
}

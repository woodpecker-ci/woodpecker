// Copyright 2022 Woodpecker Authors
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
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v66/github"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/common"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

const (
	defaultURL = "https://github.com"      // Default GitHub URL
	defaultAPI = "https://api.github.com/" // Default GitHub API URL
)

// Opts defines configuration options.
type Opts struct {
	URL        string // GitHub server url.
	Client     string // GitHub oauth client id.
	Secret     string // GitHub oauth client secret.
	SkipVerify bool   // Skip ssl verification.
	MergeRef   bool   // Clone pull requests using the merge ref.
	OnlyPublic bool   // Only obtain OAuth tokens with access to public repos.
	OAuthHost  string // Public url for oauth if different from url.
}

// New returns a Forge implementation that integrates with a GitHub Cloud or
// GitHub Enterprise version control hosting provider.
func New(opts Opts) (forge.Forge, error) {
	r := &client{
		API:        defaultAPI,
		url:        defaultURL,
		Client:     opts.Client,
		Secret:     opts.Secret,
		oAuthHost:  opts.OAuthHost,
		SkipVerify: opts.SkipVerify,
		MergeRef:   opts.MergeRef,
		OnlyPublic: opts.OnlyPublic,
	}
	if opts.URL != defaultURL {
		r.url = strings.TrimSuffix(opts.URL, "/")
		r.API = r.url + "/api/v3/"
	}

	return r, nil
}

type client struct {
	url        string
	API        string
	Client     string
	Secret     string
	SkipVerify bool
	MergeRef   bool
	OnlyPublic bool
	oAuthHost  string
}

// Name returns the string name of this driver.
func (c *client) Name() string {
	return "github"
}

// URL returns the root url of a configured forge.
func (c *client) URL() string {
	return c.url
}

// Login authenticates the session and returns the forge user details.
func (c *client) Login(ctx context.Context, req *forge_types.OAuthRequest) (*model.User, string, error) {
	config := c.newConfig()
	redirectURL := config.AuthCodeURL(req.State)

	// check the OAuth code
	if len(req.Code) == 0 {
		// TODO(bradrydzewski) we really should be using a random value here and
		// storing in a cookie for verification in the next stage of the workflow.

		return nil, redirectURL, nil
	}

	token, err := config.Exchange(c.newContext(ctx), req.Code)
	if err != nil {
		return nil, redirectURL, err
	}

	client := c.newClientToken(ctx, token.AccessToken)
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, redirectURL, err
	}

	emails, _, err := client.Users.ListEmails(ctx, nil)
	if err != nil {
		return nil, redirectURL, err
	}
	email := matchingEmail(emails, c.API)
	if email == nil {
		return nil, redirectURL, fmt.Errorf("no verified Email address for GitHub account")
	}

	return &model.User{
		Login:         user.GetLogin(),
		Email:         email.GetEmail(),
		Token:         token.AccessToken,
		Avatar:        user.GetAvatarURL(),
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(user.GetID())),
	}, redirectURL, nil
}

// Auth returns the GitHub user login for the given access token.
func (c *client) Auth(ctx context.Context, token, _ string) (string, error) {
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

// Repo returns the GitHub repository.
func (c *client) Repo(ctx context.Context, u *model.User, id model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	client := c.newClientToken(ctx, u.Token)

	if id.IsValid() {
		intID, err := strconv.ParseInt(string(id), 10, 64)
		if err != nil {
			return nil, err
		}
		repo, _, err := client.Repositories.GetByID(ctx, intID)
		if err != nil {
			return nil, err
		}
		return convertRepo(repo), nil
	}

	repo, _, err := client.Repositories.Get(ctx, owner, name)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo), nil
}

// Repos returns a list of all repositories for GitHub account, including
// organization repositories.
func (c *client) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.RepositoryListByAuthenticatedUserOptions)
	opts.PerPage = 100
	opts.Page = 1

	var repos []*model.Repo
	for opts.Page > 0 {
		list, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opts)
		if err != nil {
			return nil, err
		}
		for _, repo := range list {
			if repo.GetArchived() {
				continue
			}
			repos = append(repos, convertRepo(repo))
		}
		opts.Page = resp.NextPage
	}
	return repos, nil
}

// File fetches the file from the GitHub repository and returns its contents.
func (c *client) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.RepositoryContentGetOptions)
	opts.Ref = b.Commit
	content, _, resp, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, f, opts)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, errors.Join(err, &forge_types.ErrConfigNotFound{Configs: []string{f}})
	}
	if err != nil {
		return nil, err
	}
	if content == nil {
		return nil, fmt.Errorf("%s is a folder not a file use Dir(..)", f)
	}
	data, err := content.GetContent()
	return []byte(data), err
}

func (c *client) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*forge_types.FileMeta, error) {
	client := c.newClientToken(ctx, u.Token)

	opts := new(github.RepositoryContentGetOptions)
	opts.Ref = b.Commit
	_, data, resp, err := client.Repositories.GetContents(ctx, r.Owner, r.Name, f, opts)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, errors.Join(err, &forge_types.ErrConfigNotFound{Configs: []string{f}})
	}
	if err != nil {
		return nil, err
	}

	fc := make(chan *forge_types.FileMeta)
	errChan := make(chan error)

	for _, file := range data {
		go func(path string) {
			content, err := c.File(ctx, u, r, b, path)
			if err != nil {
				if errors.Is(err, &forge_types.ErrConfigNotFound{}) {
					err = fmt.Errorf("git tree reported existence of file but we got: %s", err.Error())
				}
				errChan <- err
			} else {
				fc <- &forge_types.FileMeta{
					Name: path,
					Data: content,
				}
			}
		}(f + "/" + *file.Name)
	}

	var files []*forge_types.FileMeta

	for i := 0; i < len(data); i++ {
		select {
		case err := <-errChan:
			return nil, err
		case fileMeta := <-fc:
			files = append(files, fileMeta)
		}
	}

	close(fc)
	close(errChan)

	return files, nil
}

func (c *client) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	token := common.UserToken(ctx, r, u)
	client := c.newClientToken(ctx, token)

	pullRequests, _, err := client.PullRequests.List(ctx, r.Owner, r.Name, &github.PullRequestListOptions{
		ListOptions: github.ListOptions{Page: p.Page, PerPage: p.PerPage},
		State:       "open",
	})
	if err != nil {
		return nil, err
	}

	result := make([]*model.PullRequest, len(pullRequests))
	for i := range pullRequests {
		result[i] = &model.PullRequest{
			Index: model.ForgeRemoteID(strconv.Itoa(pullRequests[i].GetNumber())),
			Title: pullRequests[i].GetTitle(),
		}
	}
	return result, err
}

// Netrc returns a netrc file capable of authenticating GitHub requests and
// cloning GitHub repositories. The netrc will use the global machine account
// when configured.
func (c *client) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	login := ""
	token := ""

	if u != nil {
		login = u.Token
		token = "x-oauth-basic"
	}

	host, err := common.ExtractHostFromCloneURL(r.Clone)
	if err != nil {
		return nil, err
	}

	return &model.Netrc{
		Login:    login,
		Password: token,
		Machine:  host,
	}, nil
}

// Deactivate deactivates the repository be removing registered push hooks from
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

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *client) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client := c.newClientToken(ctx, u.Token)
	org, _, err := client.Organizations.GetOrgMembership(ctx, u.Login, owner)
	if err != nil {
		return nil, err
	}

	return &model.OrgPerm{Member: org.GetState() == "active", Admin: org.GetRole() == "admin"}, nil
}

func (c *client) Org(ctx context.Context, u *model.User, owner string) (*model.Org, error) {
	client := c.newClientToken(ctx, u.Token)

	user, _, err := client.Users.Get(ctx, owner)
	log.Trace().Msgf("GitHub user for owner %s = %v", owner, user)
	if user != nil && err == nil {
		return &model.Org{
			Name:   user.GetLogin(),
			IsUser: true,
		}, nil
	}

	org, _, err := client.Organizations.Get(ctx, owner)
	log.Trace().Msgf("GitHub organization for owner %s = %v", owner, org)
	if err != nil {
		return nil, err
	}

	return &model.Org{
		Name: org.GetLogin(),
	}, nil
}

// newContext returns the GitHub oauth2 context using an HTTPClient that
// disables TLS verification if disabled in the forge settings.
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

// newConfig returns the GitHub oauth2 config.
func (c *client) newConfig() *oauth2.Config {
	scopes := []string{"user:email", "read:org"}
	if c.OnlyPublic {
		scopes = append(scopes, []string{"admin:repo_hook", "repo:status"}...)
	} else {
		scopes = append(scopes, "repo")
	}

	publicOAuthURL := c.oAuthHost
	if publicOAuthURL == "" {
		publicOAuthURL = c.url
	}

	return &oauth2.Config{
		ClientID:     c.Client,
		ClientSecret: c.Secret,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/login/oauth/authorize", publicOAuthURL),
			TokenURL: fmt.Sprintf("%s/login/oauth/access_token", c.url),
		},
		RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
	}
}

// newClientToken returns the GitHub oauth2 client.
func (c *client) newClientToken(ctx context.Context, token string) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)
	if c.SkipVerify {
		tp, _ := tc.Transport.(*oauth2.Transport)
		tp.Base = &http.Transport{
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

// matchingEmail returns matching user email.
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

// matchingHooks returns matching hook.
func matchingHooks(hooks []*github.Hook, rawURL string) *github.Hook {
	link, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		if hook.ID == nil {
			continue
		}
		hookURL, err := url.Parse(hook.Config.GetURL())
		if err == nil && hookURL.Host == link.Host {
			return hook
		}
	}
	return nil
}

var reDeploy = regexp.MustCompile(`.+/deployments/(\d+)`)

// Status sends the commit status to the forge.
// An example would be the GitHub pull request status.
func (c *client) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) error {
	client := c.newClientToken(ctx, user.Token)

	if pipeline.Event == model.EventDeploy {
		// Get id from url. If not found, skip.
		matches := reDeploy.FindStringSubmatch(pipeline.ForgeURL)
		//nolint:mnd
		if len(matches) != 2 {
			return nil
		}
		id, _ := strconv.Atoi(matches[1])

		_, _, err := client.Repositories.CreateDeploymentStatus(ctx, repo.Owner, repo.Name, int64(id), &github.DeploymentStatusRequest{
			State:       github.String(convertStatus(pipeline.Status)),
			Description: github.String(common.GetPipelineStatusDescription(pipeline.Status)),
			LogURL:      github.String(common.GetPipelineStatusURL(repo, pipeline, nil)),
		})
		return err
	}

	_, _, err := client.Repositories.CreateStatus(ctx, repo.Owner, repo.Name, pipeline.Commit, &github.RepoStatus{
		Context:     github.String(common.GetPipelineStatusContext(repo, pipeline, workflow)),
		State:       github.String(convertStatus(workflow.State)),
		Description: github.String(common.GetPipelineStatusDescription(workflow.State)),
		TargetURL:   github.String(common.GetPipelineStatusURL(repo, pipeline, workflow)),
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
		Config: &github.HookConfig{
			URL:         &link,
			ContentType: github.String("form"),
		},
	}
	_, _, err := client.Repositories.CreateHook(ctx, r.Owner, r.Name, hook)
	return err
}

// Branches returns the names of all branches for the named repository.
func (c *client) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	token := common.UserToken(ctx, r, u)
	client := c.newClientToken(ctx, token)

	githubBranches, _, err := client.Repositories.ListBranches(ctx, r.Owner, r.Name, &github.BranchListOptions{
		ListOptions: github.ListOptions{Page: p.Page, PerPage: p.PerPage},
	})
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range githubBranches {
		branches = append(branches, *branch.Name)
	}
	return branches, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch.
func (c *client) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	token := common.UserToken(ctx, r, u)
	b, _, err := c.newClientToken(ctx, token).Repositories.GetBranch(ctx, r.Owner, r.Name, branch, 1)
	if err != nil {
		return nil, err
	}
	return &model.Commit{
		SHA:      b.GetCommit().GetSHA(),
		ForgeURL: b.GetCommit().GetHTMLURL(),
	}, nil
}

// Hook parses the post-commit hook from the Request body
// and returns the required data in a standard format.
func (c *client) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	pull, repo, pipeline, err := parseHook(r, c.MergeRef)
	if err != nil {
		return nil, nil, err
	}

	if pipeline != nil && pipeline.Event == model.EventRelease && pipeline.Commit == "" {
		tagName := strings.Split(pipeline.Ref, "/")[2]
		sha, err := c.getTagCommitSHA(ctx, repo, tagName)
		if err != nil {
			return nil, nil, err
		}
		pipeline.Commit = sha
	}

	if pull != nil && len(pipeline.ChangedFiles) == 0 {
		pipeline, err = c.loadChangedFilesFromPullRequest(ctx, pull, repo, pipeline)
		if err != nil {
			return nil, nil, err
		}
	}

	return repo, pipeline, nil
}

func (c *client) loadChangedFilesFromPullRequest(ctx context.Context, pull *github.PullRequest, tmpRepo *model.Repo, pipeline *model.Pipeline) (*model.Pipeline, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return pipeline, nil
	}

	repo, err := _store.GetRepoNameFallback(tmpRepo.ForgeRemoteID, tmpRepo.FullName)
	if err != nil {
		return nil, err
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, err
	}

	pipeline.ChangedFiles, err = utils.Paginate(func(page int) ([]string, error) {
		opts := &github.ListOptions{Page: page}
		fileList := make([]string, 0, 16)
		for opts.Page > 0 {
			files, resp, err := c.newClientToken(ctx, user.Token).PullRequests.ListFiles(ctx, repo.Owner, repo.Name, pull.GetNumber(), opts)
			if err != nil {
				return nil, err
			}

			for _, file := range files {
				fileList = append(fileList, file.GetFilename(), file.GetPreviousFilename())
			}

			opts.Page = resp.NextPage
		}
		return utils.DeduplicateStrings(fileList), nil
	})

	return pipeline, err
}

func (c *client) getTagCommitSHA(ctx context.Context, repo *model.Repo, tagName string) (string, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return "", nil
	}

	repo, err := _store.GetRepoNameFallback(repo.ForgeRemoteID, repo.FullName)
	if err != nil {
		return "", err
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return "", err
	}

	gh := c.newClientToken(ctx, user.Token)

	page := 1
	var tag *github.RepositoryTag
	for {
		tags, _, err := gh.Repositories.ListTags(ctx, repo.Owner, repo.Name, &github.ListOptions{Page: page})
		if err != nil {
			return "", err
		}

		for _, t := range tags {
			if t.GetName() == tagName {
				tag = t
				break
			}
		}
		if tag != nil {
			break
		}
	}
	if tag == nil {
		return "", fmt.Errorf("could not find tag %s", tagName)
	}
	return tag.GetCommit().GetSHA(), nil
}

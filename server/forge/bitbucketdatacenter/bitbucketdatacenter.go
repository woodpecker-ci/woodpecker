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

package bitbucketdatacenter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	bb "github.com/neticdk/go-bitbucket/bitbucket"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucketdatacenter/internal"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/common"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

const listLimit = 250

// Opts defines configuration options.
type Opts struct {
	URL          string // Bitbucket server url for API access.
	Username     string // Git machine account username.
	Password     string // Git machine account password.
	ClientID     string // OAuth 2.0 client id
	ClientSecret string // OAuth 2.0 client secret
	OAuthHost    string // OAuth 2.0 host
}

type client struct {
	url          string
	urlAPI       string
	clientID     string
	clientSecret string
	oauthHost    string
	username     string
	password     string
}

// New returns a Forge implementation that integrates with Bitbucket DataCenter/Server,
// the on-premise edition of Bitbucket Cloud, formerly known as Stash.
func New(opts Opts) (forge.Forge, error) {
	config := &client{
		url:          opts.URL,
		urlAPI:       fmt.Sprintf("%s/rest", opts.URL),
		clientID:     opts.ClientID,
		clientSecret: opts.ClientSecret,
		oauthHost:    opts.OAuthHost,
		username:     opts.Username,
		password:     opts.Password,
	}

	switch {
	case opts.Username == "":
		return nil, fmt.Errorf("must have a git machine account username")
	case opts.Password == "":
		return nil, fmt.Errorf("must have a git machine account password")
	case opts.ClientID == "":
		return nil, fmt.Errorf("must have an oauth 2.0 client id")
	case opts.ClientSecret == "":
		return nil, fmt.Errorf("must have an oauth 2.0 client secret")
	}

	return config, nil
}

// Name returns the string name of this driver.
func (c *client) Name() string {
	return "bitbucket_dc"
}

// URL returns the root url of a configured forge.
func (c *client) URL() string {
	return c.url
}

func (c *client) Login(ctx context.Context, req *forge_types.OAuthRequest) (*model.User, string, error) {
	config := c.newOAuth2Config()

	// TODO: Use pkce flow (https://oauth.net/2/pkce/) ...
	redirectURL := config.AuthCodeURL(req.State)

	if len(req.Code) == 0 {
		return nil, redirectURL, nil
	}

	token, err := config.Exchange(ctx, req.Code)
	if err != nil {
		return nil, redirectURL, err
	}

	client := internal.NewClientWithToken(ctx, config.TokenSource(ctx, &oauth2.Token{
		AccessToken: token.AccessToken,
	}), c.url)
	userSlug, err := client.FindCurrentUser(ctx)
	if err != nil {
		return nil, "", err
	}

	bc, err := c.newClient(ctx, &model.User{Token: token.AccessToken})
	if err != nil {
		return nil, "", fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	user, _, err := bc.Users.GetUser(ctx, userSlug)
	if err != nil {
		return nil, "", fmt.Errorf("unable to query for user: %w", err)
	}

	u := convertUser(user, c.url)
	updateUserCredentials(u, token)
	return u, "", nil
}

func (c *client) Auth(ctx context.Context, accessToken, _ string) (string, error) {
	config := c.newOAuth2Config()
	token := &oauth2.Token{
		AccessToken: accessToken,
	}
	client := internal.NewClientWithToken(ctx, config.TokenSource(ctx, token), c.url)
	return client.FindCurrentUser(ctx)
}

func (c *client) Refresh(ctx context.Context, u *model.User) (bool, error) {
	config := c.newOAuth2Config()
	t := &oauth2.Token{
		RefreshToken: u.Secret,
	}
	ts := config.TokenSource(ctx, t)

	tok, err := ts.Token()
	if err != nil {
		return false, fmt.Errorf("unable to refresh OAuth 2.0 token from bitbucket datacenter: %w", err)
	}
	updateUserCredentials(u, tok)
	return true, nil
}

func (c *client) Repo(ctx context.Context, u *model.User, rID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	var repo *bb.Repository
	if rID.IsValid() {
		opts := &bb.RepositorySearchOptions{Permission: bb.PermissionRepoWrite, ListOptions: bb.ListOptions{Limit: listLimit}}
		for {
			repos, resp, err := bc.Projects.SearchRepositories(ctx, opts)
			if err != nil {
				return nil, fmt.Errorf("unable to search repositories: %w", err)
			}
			for _, r := range repos {
				if rID == convertID(r.ID) {
					repo = r
					break
				}
			}
			if resp.LastPage {
				break
			}
			opts.Start = resp.NextPageStart
		}
		if repo == nil {
			return nil, fmt.Errorf("unable to find repository with id: %s", rID)
		}
	} else {
		repo, _, err = bc.Projects.GetRepository(ctx, owner, name)
		if err != nil {
			return nil, fmt.Errorf("unable to get repository: %w", err)
		}
	}

	b, _, err := bc.Projects.GetDefaultBranch(ctx, repo.Project.Key, repo.Slug)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch default branch: %w", err)
	}

	perms := &model.Perm{Pull: true, Push: true}
	_, _, err = bc.Projects.ListWebhooks(ctx, repo.Project.Key, repo.Slug, &bb.ListOptions{})
	if err == nil {
		perms.Admin = true
	}

	return convertRepo(repo, perms, b.DisplayID), nil
}

func (c *client) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	opts := &bb.RepositorySearchOptions{Permission: bb.PermissionRepoWrite, ListOptions: bb.ListOptions{Limit: listLimit}}
	all := make([]*model.Repo, 0)
	for {
		repos, resp, err := bc.Projects.SearchRepositories(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to search repositories: %w", err)
		}
		for _, r := range repos {
			perms := &model.Perm{Pull: true, Push: true, Admin: false}
			all = append(all, convertRepo(r, perms, ""))
		}
		if resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}

	// Add admin permissions to relevant repositories
	opts = &bb.RepositorySearchOptions{Permission: bb.PermissionRepoAdmin, ListOptions: bb.ListOptions{Limit: listLimit}}
	for {
		repos, resp, err := bc.Projects.SearchRepositories(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to search repositories: %w", err)
		}
		for _, r := range repos {
			for i, c := range all {
				if c.ForgeRemoteID == convertID(r.ID) {
					all[i].Perm = &model.Perm{Pull: true, Push: true, Admin: true}
					break
				}
			}
		}
		if resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}

	return all, nil
}

func (c *client) File(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]byte, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	b, resp, err := bc.Projects.GetTextFileContent(ctx, r.Owner, r.Name, f, p.Commit)
	if err != nil {
		if resp.StatusCode == http.StatusNotFound {
			// requested directory might not exist
			return nil, &forge_types.ErrConfigNotFound{
				Configs: []string{f},
			}
		}
		return nil, err
	}
	return b, nil
}

func (c *client) Dir(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, path string) ([]*forge_types.FileMeta, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	opts := &bb.FilesListOptions{At: p.Commit}
	all := make([]*forge_types.FileMeta, 0)
	for {
		list, resp, err := bc.Projects.ListFiles(ctx, r.Owner, r.Name, path, opts)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				// requested directory might not exist
				return nil, &forge_types.ErrConfigNotFound{
					Configs: []string{path},
				}
			}
			return nil, err
		}
		for _, f := range list {
			fullPath := fmt.Sprintf("%s/%s", path, f)
			data, err := c.File(ctx, u, r, p, fullPath)
			if err != nil {
				return nil, err
			}
			all = append(all, &forge_types.FileMeta{Name: fullPath, Data: data})
		}
		if resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}
	return all, nil
}

func (c *client) Status(ctx context.Context, u *model.User, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) error {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return fmt.Errorf("unable to create bitbucket client: %w", err)
	}
	status := &bb.BuildStatus{
		State:       convertStatus(pipeline.Status),
		URL:         common.GetPipelineStatusURL(repo, pipeline, workflow),
		Key:         common.GetPipelineStatusContext(repo, pipeline, workflow),
		Description: common.GetPipelineStatusDescription(pipeline.Status),
	}
	_, err = bc.Projects.CreateBuildStatus(ctx, repo.Owner, repo.Name, pipeline.Commit, status)
	return err
}

func (c *client) Netrc(_ *model.User, r *model.Repo) (*model.Netrc, error) {
	host, err := common.ExtractHostFromCloneURL(r.Clone)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	return &model.Netrc{
		Login:    c.username,
		Password: c.password,
		Machine:  host,
	}, nil
}

// Branches returns the names of all branches for the named repository.
func (c *client) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	opts := &bb.BranchSearchOptions{ListOptions: convertListOptions(p)}
	all := make([]string, 0)
	for {
		branches, resp, err := bc.Projects.SearchBranches(ctx, r.Owner, r.Name, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to list branches: %w", err)
		}
		for _, b := range branches {
			all = append(all, b.DisplayID)
		}
		if !p.All || resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}

	return all, nil
}

func (c *client) BranchHead(ctx context.Context, u *model.User, r *model.Repo, b string) (*model.Commit, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}
	branches, _, err := bc.Projects.SearchBranches(ctx, r.Owner, r.Name, &bb.BranchSearchOptions{Filter: b})
	if err != nil {
		return nil, err
	}
	if len(branches) == 0 {
		return nil, fmt.Errorf("no matching branches returned")
	}
	for _, branch := range branches {
		if branch.DisplayID == b {
			return &model.Commit{
				SHA:      branch.LatestCommit,
				ForgeURL: fmt.Sprintf("%s/commits/%s", r.ForgeURL, branch.LatestCommit),
			}, nil
		}
	}
	return nil, fmt.Errorf("no matching branches found")
}

func (c *client) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	opts := &bb.PullRequestSearchOptions{ListOptions: convertListOptions(p)}
	all := make([]*model.PullRequest, 0)
	for {
		prs, resp, err := bc.Projects.SearchPullRequests(ctx, r.Owner, r.Name, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to list pull-requests: %w", err)
		}
		for _, pr := range prs {
			all = append(all, &model.PullRequest{Index: convertID(pr.ID), Title: pr.Title})
		}
		if !p.All || resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}

	return all, nil
}

func (c *client) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	err = c.Deactivate(ctx, u, r, link)
	if err != nil {
		return fmt.Errorf("unable to deactivate old webhooks: %w", err)
	}

	webhook := &bb.Webhook{
		Name:   "Woodpecker",
		URL:    link,
		Events: []bb.EventKey{bb.EventKeyRepoRefsChanged, bb.EventKeyPullRequestFrom, bb.EventKeyPullRequestMerged, bb.EventKeyPullRequestDeclined, bb.EventKeyPullRequestDeleted},
		Active: true,
		Config: &bb.WebhookConfiguration{
			Secret: r.Hash,
		},
	}
	_, _, err = bc.Projects.CreateWebhook(ctx, r.Owner, r.Name, webhook)
	return err
}

func (c *client) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	lu, err := url.Parse(link)
	if err != nil {
		return err
	}

	opts := &bb.ListOptions{}
	var ids []uint64
	for {
		hooks, resp, err := bc.Projects.ListWebhooks(ctx, r.Owner, r.Name, opts)
		if err != nil {
			return err
		}
		for _, h := range hooks {
			hu, err := url.Parse(h.URL)
			if err == nil && hu.Host == lu.Host {
				ids = append(ids, h.ID)
			}
		}
		if resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}

	for _, id := range ids {
		_, err = bc.Projects.DeleteWebhook(ctx, r.Owner, r.Name, id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	ev, payload, err := bb.ParsePayloadWithoutSignature(r)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to parse payload from webhook invocation: %w", err)
	}

	var repo *model.Repo
	var pipe *model.Pipeline
	switch e := ev.(type) {
	case *bb.RepositoryPushEvent:
		repo = convertRepo(&e.Repository, nil, "")
		pipe = convertRepositoryPushEvent(e, c.url)
	case *bb.PullRequestEvent:
		repo = convertRepo(&e.PullRequest.Source.Repository, nil, "")
		pipe = convertPullRequestEvent(e, c.url)
	default:
		return nil, nil, nil
	}

	user, repo, err := c.getUserAndRepo(ctx, repo)
	if err != nil {
		return nil, nil, err
	}

	err = bb.ValidateSignature(r, payload, []byte(repo.Hash))
	if err != nil {
		return nil, nil, fmt.Errorf("unable to validate signature on incoming webhook payload: %w", err)
	}

	pipe, err = c.updatePipelineFromCommit(ctx, user, repo, pipe)
	if err != nil {
		return nil, nil, err
	}

	if pipe == nil {
		return nil, nil, nil
	}

	return repo, pipe, nil
}

func (c *client) getUserAndRepo(ctx context.Context, r *model.Repo) (*model.User, *model.Repo, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return nil, nil, fmt.Errorf("unable to get store from context")
	}

	repo, err := _store.GetRepoForgeID(r.ForgeRemoteID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get repo: %w", err)
	}
	log.Trace().Any("repo", repo).Msg("got repo")

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get user: %w", err)
	}
	log.Trace().Any("user", user).Msg("got user")

	forge.Refresh(ctx, c, _store, user)

	return user, repo, nil
}

func (c *client) updatePipelineFromCommit(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline) (*model.Pipeline, error) {
	if p == nil {
		return nil, nil
	}

	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	commit, _, err := bc.Projects.GetCommit(ctx, r.Owner, r.Name, p.Commit)
	if err != nil {
		return nil, fmt.Errorf("unable to read commit: %w", err)
	}
	p.Message = commit.Message

	opts := &bb.ListOptions{}
	for {
		changes, resp, err := bc.Projects.ListChanges(ctx, r.Owner, r.Name, p.Commit, opts)
		if err != nil {
			return nil, fmt.Errorf("unable to list commit changes: %w", err)
		}
		for _, ch := range changes {
			p.ChangedFiles = append(p.ChangedFiles, ch.Path.Title)
		}
		if resp.LastPage {
			break
		}
		opts.Start = resp.NextPageStart
	}

	return p, nil
}

// Teams is not supported.
func (*client) Teams(_ context.Context, _ *model.User) ([]*model.Team, error) {
	var teams []*model.Team
	return teams, nil
}

// TeamPerm is not supported.
func (*client) TeamPerm(_ *model.User, _ string) (*model.Perm, error) {
	return nil, nil
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *client) OrgMembership(_ context.Context, _ *model.User, _ string) (*model.OrgPerm, error) {
	// TODO: Not implemented currently
	return nil, nil
}

// Org fetches the organization from the forge by name. If the name is a user an org with type user is returned.
func (c *client) Org(_ context.Context, _ *model.User, owner string) (*model.Org, error) {
	if strings.HasPrefix(owner, "~") {
		return &model.Org{
			Name:   owner,
			IsUser: true,
		}, nil
	}
	return &model.Org{
		Name:   owner,
		IsUser: false,
	}, nil
}

func (c *client) newOAuth2Config() *oauth2.Config {
	publicOAuthURL := c.oauthHost
	if publicOAuthURL == "" {
		publicOAuthURL = c.urlAPI
	}

	return &oauth2.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/oauth2/latest/authorize", publicOAuthURL),
			TokenURL: fmt.Sprintf("%s/oauth2/latest/token", c.urlAPI),
		},
		Scopes:      []string{string(bb.PermissionRepoRead), string(bb.PermissionRepoWrite), string(bb.PermissionRepoAdmin)},
		RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
	}
}

func (c *client) newClient(ctx context.Context, u *model.User) (*bb.Client, error) {
	config := c.newOAuth2Config()
	t := &oauth2.Token{
		AccessToken: u.Token,
	}
	client := config.Client(ctx, t)
	return bb.NewClient(c.urlAPI, client)
}

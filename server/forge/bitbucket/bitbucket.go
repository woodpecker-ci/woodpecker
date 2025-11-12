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

package bitbucket

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"

	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/common"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
	shared_utils "go.woodpecker-ci.org/woodpecker/v3/shared/utils"
)

// Bitbucket cloud endpoints.
const (
	DefaultAPI = "https://api.bitbucket.org"
	DefaultURL = "https://bitbucket.org"
	pageSize   = 100
)

// Opts are forge options for bitbucket.
type Opts struct {
	OAuthClientID     string
	OAuthClientSecret string
}

type config struct {
	api           string
	url           string
	oAuthClientID string
	oAuthSecret   string
}

// New returns a new forge Configuration for integrating with the Bitbucket
// repository hosting service at https://bitbucket.org
func New(opts *Opts) (forge.Forge, error) {
	return &config{
		api:           DefaultAPI,
		url:           DefaultURL,
		oAuthClientID: opts.OAuthClientID,
		oAuthSecret:   opts.OAuthClientSecret,
	}, nil
	// TODO: add checks
}

// Name returns the string name of this driver.
func (c *config) Name() string {
	return "bitbucket"
}

// URL returns the root url of a configured forge.
func (c *config) URL() string {
	return c.url
}

// Login authenticates an account with Bitbucket using the oauth2 protocol. The
// Bitbucket account details are returned when the user is successfully authenticated.
func (c *config) Login(ctx context.Context, req *forge_types.OAuthRequest) (*model.User, string, error) {
	config := c.newOAuth2Config()
	redirectURL := config.AuthCodeURL(req.State)

	// check the OAuth code
	if len(req.Code) == 0 {
		return nil, redirectURL, nil
	}

	token, err := config.Exchange(ctx, req.Code)
	if err != nil {
		return nil, redirectURL, err
	}

	client := internal.NewClient(ctx, c.api, config.Client(ctx, token))
	curr, err := client.FindCurrent()
	if err != nil {
		return nil, redirectURL, err
	}
	return convertUser(curr, token), redirectURL, nil
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
	config := c.newOAuth2Config()
	source := config.TokenSource(
		ctx, &oauth2.Token{RefreshToken: user.RefreshToken})

	token, err := source.Token()
	if err != nil || len(token.AccessToken) == 0 {
		return false, err
	}

	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken
	user.Expiry = token.Expiry.UTC().Unix()
	return true, nil
}

// Teams returns a list of all team membership for the Bitbucket account.
func (c *config) Teams(ctx context.Context, u *model.User, p *model.ListOptions) ([]*model.Team, error) {
	setListOptions(p)

	opts := &internal.ListWorkspacesOpts{
		PageLen: p.PerPage,
		Page:    p.Page,
		Role:    "member",
	}
	resp, err := c.newClient(ctx, u).ListWorkspaces(opts)
	if err != nil {
		return nil, err
	}
	return convertWorkspaceList(resp.Values), nil
}

// Repo returns the named Bitbucket repository.
func (c *config) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	if remoteID.IsValid() {
		name = string(remoteID)
	}
	if owner == "" {
		repos, err := c.Repos(ctx, u, &model.ListOptions{Page: 1})
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if string(repo.ForgeRemoteID) == name {
				owner = repo.Owner
				break
			}
		}
	}
	client := c.newClient(ctx, u)
	repo, err := client.FindRepo(owner, name)
	if err != nil {
		return nil, err
	}
	perm, err := client.GetPermission(repo.FullName)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo, perm), nil
}

// Repos returns a list of all repositories for Bitbucket account, including
// organization repositories.
func (c *config) Repos(ctx context.Context, u *model.User, p *model.ListOptions) ([]*model.Repo, error) {
	setListOptions(p)

	client := c.newClient(ctx, u)

	resp, err := client.ListWorkspaces(&internal.ListWorkspacesOpts{
		Page:    p.Page,
		PageLen: p.PerPage,
		Role:    "member",
	})
	if err != nil {
		return nil, err
	}

	userPermissions, err := client.ListPermissionsAll()
	if err != nil {
		return nil, err
	}

	userPermissionsByRepo := make(map[string]*internal.RepoPerm)
	for _, permission := range userPermissions {
		userPermissionsByRepo[permission.Repo.FullName] = permission
	}

	var all []*model.Repo
	for _, workspace := range resp.Values {
		repos, err := client.ListReposAll(workspace.Slug)
		if err != nil {
			return nil, err
		}
		for _, repo := range repos {
			if perm, ok := userPermissionsByRepo[repo.FullName]; ok {
				all = append(all, convertRepo(repo, perm))
			}
		}
	}
	return all, nil
}

// File fetches the file from the Bitbucket repository and returns its contents.
func (c *config) File(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]byte, error) {
	config, err := c.newClient(ctx, u).FindSource(r.Owner, r.Name, p.Commit, f)
	if err != nil {
		var rspErr internal.Error
		if ok := errors.As(err, &rspErr); ok && rspErr.Status == http.StatusNotFound {
			return nil, &forge_types.ErrConfigNotFound{
				Configs: []string{f},
			}
		}
		return nil, err
	}
	return []byte(*config), nil
}

// Dir fetches a folder from the bitbucket repository.
func (c *config) Dir(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]*forge_types.FileMeta, error) {
	var page *string
	repoPathFiles := []*forge_types.FileMeta{}
	client := c.newClient(ctx, u)
	for {
		filesResp, err := client.GetRepoFiles(r.Owner, r.Name, p.Commit, f, page)
		if err != nil {
			var rspErr internal.Error
			if ok := errors.As(err, &rspErr); ok && rspErr.Status == http.StatusNotFound {
				return nil, &forge_types.ErrConfigNotFound{
					Configs: []string{f},
				}
			}
			return nil, err
		}
		for _, file := range filesResp.Values {
			_, filename := filepath.Split(file.Path)
			repoFile := forge_types.FileMeta{
				Name: filename,
			}
			if file.Type == "commit_file" {
				fileData, err := c.newClient(ctx, u).FindSource(r.Owner, r.Name, p.Commit, file.Path)
				if err != nil {
					return nil, err
				}
				if fileData != nil {
					repoFile.Data = []byte(*fileData)
				}
			}
			repoPathFiles = append(repoPathFiles, &repoFile)
		}

		// Check for more results page
		if filesResp.Next == nil {
			break
		}
		nextPageURL, err := url.Parse(*filesResp.Next)
		if err != nil {
			return nil, err
		}
		params, err := url.ParseQuery(nextPageURL.RawQuery)
		if err != nil {
			return nil, err
		}
		nextPage := params.Get("page")
		if len(nextPage) == 0 {
			break
		}
		page = &nextPage
	}
	return repoPathFiles, nil
}

// Status creates a pipeline status for the Bitbucket commit.
func (c *config) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) error {
	status := internal.PipelineStatus{
		State: convertStatus(workflow.State),
		Desc:  common.GetPipelineStatusDescription(workflow.State),
		Key:   common.GetPipelineStatusContext(repo, pipeline, workflow),
		URL:   common.GetPipelineStatusURL(repo, pipeline, workflow),
	}
	return c.newClient(ctx, user).CreateStatus(repo.Owner, repo.Name, pipeline.Commit, &status)
}

// Activate activates the repository by registering repository push hooks with
// the Bitbucket repository. Prior to registering hook, previously created hooks
// are deleted.
func (c *config) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	rawURL, err := url.Parse(link)
	if err != nil {
		return err
	}
	_ = c.Deactivate(ctx, u, r, link)

	return c.newClient(ctx, u).CreateHook(r.Owner, r.Name, &internal.Hook{
		Active: true,
		Desc:   rawURL.Host,
		Events: []string{"repo:push", "pullrequest:created", "pullrequest:updated", "pullrequest:fulfilled", "pullrequest:rejected"},
		URL:    link,
	})
}

// Deactivate deactivates the repository be removing repository push hooks from
// the Bitbucket repository.
func (c *config) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := c.newClient(ctx, u)

	hooks, err := shared_utils.Paginate(func(page int) ([]*internal.Hook, error) {
		hooks, err := client.ListHooks(r.Owner, r.Name, &internal.ListOpts{
			Page: page,
		})
		if err != nil {
			return nil, err
		}
		return hooks.Values, nil
	}, -1)
	if err != nil {
		return err
	}
	hook := matchingHooks(hooks, link)
	if hook != nil {
		return client.DeleteHook(r.Owner, r.Name, hook.UUID)
	}
	return nil
}

// Netrc returns a netrc file capable of authenticating Bitbucket requests and
// cloning Bitbucket repositories.
func (c *config) Netrc(u *model.User, _ *model.Repo) (*model.Netrc, error) {
	return &model.Netrc{
		Machine:  "bitbucket.org",
		Login:    "x-token-auth",
		Password: u.AccessToken,
		Type:     model.ForgeTypeBitbucket,
	}, nil
}

// Branches returns the names of all branches for the named repository.
func (c *config) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	setListOptions(p)

	opts := internal.ListOpts{Page: p.Page, PageLen: p.PerPage}
	bitbucketBranches, err := c.newClient(ctx, u).ListBranches(r.Owner, r.Name, &opts)
	if err != nil {
		return nil, err
	}
	branches := make([]string, 0)
	for _, branch := range bitbucketBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch.
func (c *config) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	commit, err := c.newClient(ctx, u).GetBranchHead(r.Owner, r.Name, branch)
	if err != nil {
		return nil, err
	}
	return &model.Commit{
		SHA:      commit.Hash,
		ForgeURL: commit.Links.HTML.Href,
	}, nil
}

// PullRequests returns the pull requests of the named repository.
func (c *config) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	setListOptions(p)

	opts := internal.ListOpts{Page: p.Page, PageLen: p.PerPage}
	pullRequests, err := c.newClient(ctx, u).ListPullRequests(r.Owner, r.Name, &opts)
	if err != nil {
		return nil, err
	}
	var result []*model.PullRequest
	for _, pullRequest := range pullRequests {
		result = append(result, &model.PullRequest{
			Index: model.ForgeRemoteID(strconv.Itoa(int(pullRequest.ID))),
			Title: pullRequest.Title,
		})
	}
	return result, nil
}

// Hook parses the incoming Bitbucket hook and returns the Repository and
// Pipeline details. If the hook is unsupported nil values are returned.
func (c *config) Hook(ctx context.Context, req *http.Request) (*model.Repo, *model.Pipeline, error) {
	pr, repo, pl, err := parseHook(req)
	if err != nil {
		return nil, nil, err
	}

	u, err := common.RepoUserForgeID(ctx, repo.ForgeRemoteID)
	if err != nil {
		return nil, nil, err
	}

	switch pl.Event {
	case model.EventPush:
		// List only the latest push changes
		pl.ChangedFiles, err = c.newClient(ctx, u).ListChangedFiles(repo.Owner, repo.Name, pl.Commit)
		if err != nil {
			return nil, nil, err
		}
	case model.EventPull:
		client := c.newClient(ctx, u)

		if pr == nil {
			return nil, nil, fmt.Errorf("can't run hook against empty PR information")
		}

		// List all changes between source & destination branch
		pl.ChangedFiles, err = client.ListChangedFiles(repo.Owner, repo.Name, fmt.Sprintf("%s..%s", pr.PullRequest.Source.Branch.Name, pr.PullRequest.Dest.Branch.Name))
		if err != nil {
			return nil, nil, err
		}
	default:
	}

	repo, err = c.Repo(ctx, u, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		return nil, nil, err
	}

	return repo, pl, nil
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

func (c *config) Org(ctx context.Context, u *model.User, owner string) (*model.Org, error) {
	workspace, err := c.newClient(ctx, u).GetWorkspace(owner)
	if err != nil {
		return nil, err
	}

	return &model.Org{
		Name:   workspace.Slug,
		IsUser: false, // bitbucket uses workspaces (similar to orgs) for teams and single users so we cannot distinguish between them
	}, nil
}

// helper function to return the bitbucket oauth2 client.
func (c *config) newClient(ctx context.Context, u *model.User) *internal.Client {
	if u == nil {
		return c.newClientToken(ctx, "", "")
	}
	return c.newClientToken(ctx, u.AccessToken, u.RefreshToken)
}

// helper function to return the bitbucket oauth2 client.
func (c *config) newClientToken(ctx context.Context, accessToken, refreshToken string) *internal.Client {
	client := internal.NewClientToken(
		ctx,
		c.api,
		accessToken,
		refreshToken,
		&oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	)
	client.Client = httputil.WrapClient(client.Client, "forge-bitbucket")
	return client
}

// helper function to return the bitbucket oauth2 config.
func (c *config) newOAuth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.oAuthClientID,
		ClientSecret: c.oAuthSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("%s/site/oauth2/authorize", c.url),
			TokenURL: fmt.Sprintf("%s/site/oauth2/access_token", c.url),
		},
		RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
	}
}

// helper function to return matching hooks.
func matchingHooks(hooks []*internal.Hook, rawURL string) *internal.Hook {
	link, err := url.Parse(rawURL)
	if err != nil {
		return nil
	}
	for _, hook := range hooks {
		hookURL, err := url.Parse(hook.URL)
		if err == nil && hookURL.Host == link.Host {
			return hook
		}
	}
	return nil
}

func setListOptions(p *model.ListOptions) {
	if p.PerPage > pageSize || p.PerPage == 0 {
		p.PerPage = pageSize
	}
}

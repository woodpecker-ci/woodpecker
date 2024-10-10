// Copyright 2022 Woodpecker Authors
// Copyright 2021 Informatyka Boguslawski sp. z o.o. sp.k., http://www.ib.pl/
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

package gitea

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/sdk/gitea"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/common"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	shared_utils "go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

const (
	authorizeTokenURL = "%s/login/oauth/authorize"
	accessTokenURL    = "%s/login/oauth/access_token"
	defaultPageSize   = 50
	giteaDevVersion   = "v1.21.0"
)

type Gitea struct {
	url          string
	ClientID     string
	ClientSecret string
	OAuthHost    string
	SkipVerify   bool
	pageSize     int
}

// Opts defines configuration options.
type Opts struct {
	URL        string // Gitea server url.
	Client     string // OAuth2 Client ID
	Secret     string // OAuth2 Client Secret
	OAuthHost  string // OAuth2 Host
	SkipVerify bool   // Skip ssl verification.
}

// New returns a Forge implementation that integrates with Gitea,
// an open source Git service written in Go. See https://gitea.io/
func New(opts Opts) (forge.Forge, error) {
	return &Gitea{
		url:          opts.URL,
		ClientID:     opts.Client,
		ClientSecret: opts.Secret,
		OAuthHost:    opts.OAuthHost,
		SkipVerify:   opts.SkipVerify,
	}, nil
}

// Name returns the string name of this driver.
func (c *Gitea) Name() string {
	return "gitea"
}

// URL returns the root url of a configured forge.
func (c *Gitea) URL() string {
	return c.url
}

func (c *Gitea) oauth2Config(ctx context.Context) (*oauth2.Config, context.Context) {
	publicOAuthURL := c.OAuthHost
	if publicOAuthURL == "" {
		publicOAuthURL = c.url
	}
	return &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf(authorizeTokenURL, publicOAuthURL),
				TokenURL: fmt.Sprintf(accessTokenURL, c.url),
			},
			RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
		},

		context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.SkipVerify},
			Proxy:           http.ProxyFromEnvironment,
		}})
}

// Login authenticates an account with Gitea using basic authentication. The
// Gitea account details are returned when the user is successfully authenticated.
func (c *Gitea) Login(ctx context.Context, req *forge_types.OAuthRequest) (*model.User, string, error) {
	config, oauth2Ctx := c.oauth2Config(ctx)
	redirectURL := config.AuthCodeURL(req.State)

	// check the OAuth code
	if len(req.Code) == 0 {
		return nil, redirectURL, nil
	}

	token, err := config.Exchange(oauth2Ctx, req.Code)
	if err != nil {
		return nil, redirectURL, err
	}

	client, err := c.newClientToken(ctx, token.AccessToken)
	if err != nil {
		return nil, redirectURL, err
	}
	account, _, err := client.GetMyUserInfo()
	if err != nil {
		return nil, redirectURL, err
	}

	return &model.User{
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		Expiry:        token.Expiry.UTC().Unix(),
		Login:         account.UserName,
		Email:         account.Email,
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(account.ID)),
		Avatar:        expandAvatar(c.url, account.AvatarURL),
	}, redirectURL, nil
}

// Auth uses the Gitea oauth2 access token and refresh token to authenticate
// a session and return the Gitea account login.
func (c *Gitea) Auth(ctx context.Context, token, _ string) (string, error) {
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
// refreshed, the user is updated and a true value is returned.
func (c *Gitea) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config, oauth2Ctx := c.oauth2Config(ctx)
	config.RedirectURL = ""

	source := config.TokenSource(oauth2Ctx, &oauth2.Token{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		Expiry:       time.Unix(user.Expiry, 0),
	})

	token, err := source.Token()
	if err != nil || len(token.AccessToken) == 0 {
		return false, err
	}

	user.AccessToken = token.AccessToken
	user.RefreshToken = token.RefreshToken
	user.Expiry = token.Expiry.UTC().Unix()
	return true, nil
}

// Teams is supported by the Gitea driver.
func (c *Gitea) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return nil, err
	}

	return shared_utils.Paginate(func(page int) ([]*model.Team, error) {
		orgs, _, err := client.ListMyOrgs(
			gitea.ListOrgsOptions{
				ListOptions: gitea.ListOptions{
					Page:     page,
					PageSize: c.perPage(ctx),
				},
			},
		)
		teams := make([]*model.Team, 0, len(orgs))
		for _, org := range orgs {
			teams = append(teams, toTeam(org, c.url))
		}
		return teams, err
	})
}

// TeamPerm is not supported by the Gitea driver.
func (c *Gitea) TeamPerm(_ *model.User, _ string) (*model.Perm, error) {
	return nil, nil
}

// Repo returns the Gitea repository.
func (c *Gitea) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return nil, err
	}

	if remoteID.IsValid() {
		intID, err := strconv.ParseInt(string(remoteID), 10, 64)
		if err != nil {
			return nil, err
		}
		repo, _, err := client.GetRepoByID(intID)
		if err != nil {
			return nil, err
		}
		return toRepo(repo), nil
	}

	repo, _, err := client.GetRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return toRepo(repo), nil
}

// Repos returns a list of all repositories for the Gitea account, including
// organization repositories.
func (c *Gitea) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return nil, err
	}

	repos, err := shared_utils.Paginate(func(page int) ([]*gitea.Repository, error) {
		repos, _, err := client.ListMyRepos(
			gitea.ListReposOptions{
				ListOptions: gitea.ListOptions{
					Page:     page,
					PageSize: c.perPage(ctx),
				},
			},
		)
		return repos, err
	})

	result := make([]*model.Repo, 0, len(repos))
	for _, repo := range repos {
		if repo.Archived {
			continue
		}
		result = append(result, toRepo(repo))
	}
	return result, err
}

// File fetches the file from the Gitea repository and returns its contents.
func (c *Gitea) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error) {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return nil, err
	}

	cfg, resp, err := client.GetFile(r.Owner, r.Name, b.Commit, f)
	if err != nil && resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, errors.Join(err, &forge_types.ErrConfigNotFound{Configs: []string{f}})
	}
	return cfg, err
}

func (c *Gitea) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*forge_types.FileMeta, error) {
	var configs []*forge_types.FileMeta

	client, err := c.newClientToken(ctx, u.AccessToken)
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
				if errors.Is(err, &forge_types.ErrConfigNotFound{}) {
					return nil, fmt.Errorf("git tree reported existence of file but we got: %s", err.Error())
				}
				return nil, fmt.Errorf("multi-pipeline cannot get %s: %w", e.Path, err)
			}

			configs = append(configs, &forge_types.FileMeta{
				Name: e.Path,
				Data: data,
			})
		}
	}

	return configs, nil
}

// Status is supported by the Gitea driver.
func (c *Gitea) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) error {
	client, err := c.newClientToken(ctx, user.AccessToken)
	if err != nil {
		return err
	}

	_, _, err = client.CreateStatus(
		repo.Owner,
		repo.Name,
		pipeline.Commit,
		gitea.CreateStatusOption{
			State:       getStatus(workflow.State),
			TargetURL:   common.GetPipelineStatusURL(repo, pipeline, workflow),
			Description: common.GetPipelineStatusDescription(workflow.State),
			Context:     common.GetPipelineStatusContext(repo, pipeline, workflow),
		},
	)
	return err
}

// Netrc returns a netrc file capable of authenticating Gitea requests and
// cloning Gitea repositories. The netrc will use the global machine account
// when configured.
func (c *Gitea) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	login := ""
	token := ""

	if u != nil {
		login = u.Login
		token = u.AccessToken
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
		Events: []string{"push", "create", "pull_request", "release"},
		Active: true,
	}

	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return err
	}
	_, response, err := client.CreateRepoHook(r.Owner, r.Name, hook)
	if err != nil {
		if response != nil {
			if response.StatusCode == http.StatusNotFound {
				return fmt.Errorf("could not find repository")
			}
			if response.StatusCode == http.StatusOK {
				return fmt.Errorf("could not find repository, repository was probably renamed")
			}
		}
		return err
	}
	return nil
}

// Deactivate deactivates the repository be removing repository push hooks from
// the Gitea repository.
func (c *Gitea) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return err
	}

	hooks, err := shared_utils.Paginate(func(page int) ([]*gitea.Hook, error) {
		hooks, _, err := client.ListRepoHooks(r.Owner, r.Name, gitea.ListHooksOptions{
			ListOptions: gitea.ListOptions{
				Page:     page,
				PageSize: c.perPage(ctx),
			},
		})
		return hooks, err
	})
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
func (c *Gitea) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	token := common.UserToken(ctx, r, u)
	client, err := c.newClientToken(ctx, token)
	if err != nil {
		return nil, err
	}

	branches, _, err := client.ListRepoBranches(r.Owner, r.Name,
		gitea.ListRepoBranchesOptions{ListOptions: gitea.ListOptions{Page: p.Page, PageSize: p.PerPage}})
	if err != nil {
		return nil, err
	}
	result := make([]string, len(branches))
	for i := range branches {
		result[i] = branches[i].Name
	}
	return result, err
}

// BranchHead returns the sha of the head (latest commit) of the specified branch.
func (c *Gitea) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	token := common.UserToken(ctx, r, u)
	client, err := c.newClientToken(ctx, token)
	if err != nil {
		return nil, err
	}

	b, _, err := client.GetRepoBranch(r.Owner, r.Name, branch)
	if err != nil {
		return nil, err
	}
	return &model.Commit{
		SHA:      b.Commit.ID,
		ForgeURL: b.Commit.URL,
	}, nil
}

func (c *Gitea) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	token := common.UserToken(ctx, r, u)
	client, err := c.newClientToken(ctx, token)
	if err != nil {
		return nil, err
	}

	pullRequests, resp, err := client.ListRepoPullRequests(r.Owner, r.Name, gitea.ListPullRequestsOptions{
		ListOptions: gitea.ListOptions{Page: p.Page, PageSize: p.PerPage},
		State:       gitea.StateOpen,
	})
	if err != nil {
		// Repositories without commits return empty list with status code 404
		if pullRequests != nil && resp != nil && resp.StatusCode == http.StatusNotFound {
			err = nil
		} else {
			return nil, err
		}
	}

	result := make([]*model.PullRequest, len(pullRequests))
	for i := range pullRequests {
		result[i] = &model.PullRequest{
			Index: model.ForgeRemoteID(strconv.Itoa(int(pullRequests[i].Index))),
			Title: pullRequests[i].Title,
		}
	}
	return result, err
}

// Hook parses the incoming Gitea hook and returns the Repository and Pipeline
// details. If the hook is unsupported nil values are returned.
func (c *Gitea) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	repo, pipeline, err := parseHook(r)
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

	if pipeline != nil && (pipeline.Event == model.EventPull || pipeline.Event == model.EventPullClosed) && len(pipeline.ChangedFiles) == 0 {
		index, err := strconv.ParseInt(strings.Split(pipeline.Ref, "/")[2], 10, 64)
		if err != nil {
			return nil, nil, err
		}
		pipeline.ChangedFiles, err = c.getChangedFilesForPR(ctx, repo, index)
		if err != nil {
			log.Error().Err(err).Msgf("could not get changed files for PR %s#%d", repo.FullName, index)
		}
	}

	return repo, pipeline, nil
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *Gitea) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return nil, err
	}

	member, _, err := client.CheckOrgMembership(owner, u.Login)
	if err != nil {
		return nil, err
	}

	if !member {
		return &model.OrgPerm{}, nil
	}

	perm, _, err := client.GetOrgPermissions(owner, u.Login)
	if err != nil {
		return &model.OrgPerm{Member: member}, err
	}

	return &model.OrgPerm{Member: member, Admin: perm.IsAdmin || perm.IsOwner}, nil
}

func (c *Gitea) Org(ctx context.Context, u *model.User, owner string) (*model.Org, error) {
	client, err := c.newClientToken(ctx, u.AccessToken)
	if err != nil {
		return nil, err
	}

	org, _, orgErr := client.GetOrg(owner)
	if orgErr == nil && org != nil {
		return &model.Org{
			Name:    org.UserName,
			Private: gitea.VisibleType(org.Visibility) != gitea.VisibleTypePublic,
		}, nil
	}

	user, _, err := client.GetUserInfo(owner)
	if err != nil {
		if orgErr != nil {
			err = errors.Join(orgErr, err)
		}
		return nil, err
	}
	return &model.Org{
		Name:    user.UserName,
		IsUser:  true,
		Private: user.Visibility != gitea.VisibleTypePublic,
	}, nil
}

// newClientToken returns the Gitea client with Token.
func (c *Gitea) newClientToken(ctx context.Context, token string) (*gitea.Client, error) {
	httpClient := &http.Client{}
	if c.SkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client, err := gitea.NewClient(c.url, gitea.SetToken(token), gitea.SetHTTPClient(httpClient), gitea.SetContext(ctx))
	if err != nil &&
		(errors.Is(err, &gitea.ErrUnknownVersion{}) || strings.Contains(err.Error(), "Malformed version")) {
		// we guess it's a dev gitea version
		log.Error().Err(err).Msgf("could not detect gitea version, assume dev version %s", giteaDevVersion)
		client, err = gitea.NewClient(c.url, gitea.SetGiteaVersion(giteaDevVersion), gitea.SetToken(token), gitea.SetHTTPClient(httpClient), gitea.SetContext(ctx))
	}
	return client, err
}

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
	case model.StatusFailure:
		return gitea.StatusFailure
	case model.StatusKilled:
		return gitea.StatusFailure
	case model.StatusDeclined:
		return gitea.StatusWarning
	case model.StatusError:
		return gitea.StatusError
	default:
		return gitea.StatusFailure
	}
}

func (c *Gitea) getChangedFilesForPR(ctx context.Context, repo *model.Repo, index int64) ([]string, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return []string{}, nil
	}

	repo, err := _store.GetRepoNameFallback(repo.ForgeRemoteID, repo.FullName)
	if err != nil {
		return nil, err
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, err
	}

	client, err := c.newClientToken(ctx, user.AccessToken)
	if err != nil {
		return nil, err
	}

	return shared_utils.Paginate(func(page int) ([]string, error) {
		giteaFiles, _, err := client.ListPullRequestFiles(repo.Owner, repo.Name, index,
			gitea.ListPullRequestFilesOptions{ListOptions: gitea.ListOptions{Page: page}})
		if err != nil {
			return nil, err
		}

		var files []string
		for _, file := range giteaFiles {
			files = append(files, file.Filename)
		}
		return files, nil
	})
}

func (c *Gitea) getTagCommitSHA(ctx context.Context, repo *model.Repo, tagName string) (string, error) {
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

	client, err := c.newClientToken(ctx, user.AccessToken)
	if err != nil {
		return "", err
	}

	tag, _, err := client.GetTag(repo.Owner, repo.Name, tagName)
	if err != nil {
		return "", err
	}

	return tag.Commit.SHA, nil
}

func (c *Gitea) perPage(ctx context.Context) int {
	if c.pageSize == 0 {
		client, err := c.newClientToken(ctx, "")
		if err != nil {
			return defaultPageSize
		}

		api, _, err := client.GetGlobalAPISettings()
		if err != nil {
			return defaultPageSize
		}
		c.pageSize = api.MaxResponseItems
	}
	return c.pageSize
}

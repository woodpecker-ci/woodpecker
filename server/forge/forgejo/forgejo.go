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
//
// This file has been modified by Informatyka Boguslawski sp. z o.o. sp.k.

package forgejo

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	forgejo "code.gitea.io/sdk/gitea"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/common"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

const (
	authorizeTokenURL = "%s/login/oauth/authorize"
	accessTokenURL    = "%s/login/oauth/access_token"
	perPage           = 50
	forgejoDevVersion = "v1.18.0"
)

type Forgejo struct {
	URL          string
	ClientID     string
	ClientSecret string
	SkipVerify   bool
}

// Opts defines configuration options.
type Opts struct {
	URL        string // Forgejo server url.
	Client     string // OAuth2 Client ID
	Secret     string // OAuth2 Client Secret
	SkipVerify bool   // Skip ssl verification.
}

// New returns a Forge implementation that integrates with Forgejo,
// an open source Git service written in Go. See https://forgejo.org/
func New(opts Opts) (forge.Forge, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err == nil {
		u.Host = host
	}
	return &Forgejo{
		URL:          opts.URL,
		ClientID:     opts.Client,
		ClientSecret: opts.Secret,
		SkipVerify:   opts.SkipVerify,
	}, nil
}

// Name returns the string name of this driver
func (c *Forgejo) Name() string {
	return "forgejo"
}

func (c *Forgejo) oauth2Config(ctx context.Context) (*oauth2.Config, context.Context) {
	return &oauth2.Config{
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf(authorizeTokenURL, c.URL),
				TokenURL: fmt.Sprintf(accessTokenURL, c.URL),
			},
			RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
		},

		context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.SkipVerify},
			Proxy:           http.ProxyFromEnvironment,
		}})
}

// Login authenticates an account with Forgejo using basic authentication. The
// Forgejo account details are returned when the user is successfully authenticated.
func (c *Forgejo) Login(ctx context.Context, w http.ResponseWriter, req *http.Request) (*model.User, error) {
	config, oauth2Ctx := c.oauth2Config(ctx)

	// get the OAuth errors
	if err := req.FormValue("error"); err != "" {
		return nil, &forge_types.AuthError{
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

	token, err := config.Exchange(oauth2Ctx, code)
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

// Auth uses the Forgejo oauth2 access token and refresh token to authenticate
// a session and return the Forgejo account login.
func (c *Forgejo) Auth(ctx context.Context, token, secret string) (string, error) {
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

// Refresh refreshes the Forgejo oauth2 access token. If the token is
// refreshed the user is updated and a true value is returned.
func (c *Forgejo) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config, oauth2Ctx := c.oauth2Config(ctx)
	config.RedirectURL = ""

	source := config.TokenSource(oauth2Ctx, &oauth2.Token{
		AccessToken:  user.Token,
		RefreshToken: user.Secret,
		Expiry:       time.Unix(user.Expiry, 0),
	})

	token, err := source.Token()
	if err != nil || len(token.AccessToken) == 0 {
		return false, err
	}

	user.Token = token.AccessToken
	user.Secret = token.RefreshToken
	user.Expiry = token.Expiry.UTC().Unix()
	return true, nil
}

// Teams is supported by the Forgejo driver.
func (c *Forgejo) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	return common.Paginate(func(page int) ([]*model.Team, error) {
		orgs, _, err := client.ListMyOrgs(
			forgejo.ListOrgsOptions{
				ListOptions: forgejo.ListOptions{
					Page:     page,
					PageSize: perPage,
				},
			},
		)
		teams := make([]*model.Team, 0, len(orgs))
		for _, org := range orgs {
			teams = append(teams, toTeam(org, c.URL))
		}
		return teams, err
	})
}

// TeamPerm is not supported by the Forgejo driver.
func (c *Forgejo) TeamPerm(u *model.User, org string) (*model.Perm, error) {
	return nil, nil
}

// Repo returns the Forgejo repository.
func (c *Forgejo) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	client, err := c.newClientToken(ctx, u.Token)
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

// Repos returns a list of all repositories for the Forgejo account, including
// organization repositories.
func (c *Forgejo) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	return common.Paginate(func(page int) ([]*model.Repo, error) {
		repos, _, err := client.ListMyRepos(
			forgejo.ListReposOptions{
				ListOptions: forgejo.ListOptions{
					Page:     page,
					PageSize: perPage,
				},
			},
		)
		result := make([]*model.Repo, 0, len(repos))
		for _, repo := range repos {
			result = append(result, toRepo(repo))
		}
		return result, err
	})
}

// Perm returns the user permissions for the named Forgejo repository.
func (c *Forgejo) Perm(ctx context.Context, u *model.User, r *model.Repo) (*model.Perm, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	repo, _, err := client.GetRepo(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}
	return toPerm(repo.Permissions), nil
}

// File fetches the file from the Forgejo repository and returns its contents.
func (c *Forgejo) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error) {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	cfg, _, err := client.GetFile(r.Owner, r.Name, b.Commit, f)
	return cfg, err
}

func (c *Forgejo) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*forge_types.FileMeta, error) {
	var configs []*forge_types.FileMeta

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

			configs = append(configs, &forge_types.FileMeta{
				Name: e.Path,
				Data: data,
			})
		}
	}

	return configs, nil
}

// Status is supported by the Forgejo driver.
func (c *Forgejo) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, step *model.Step) error {
	client, err := c.newClientToken(ctx, user.Token)
	if err != nil {
		return err
	}

	_, _, err = client.CreateStatus(
		repo.Owner,
		repo.Name,
		pipeline.Commit,
		forgejo.CreateStatusOption{
			State:       getStatus(step.State),
			TargetURL:   common.GetPipelineStatusLink(repo, pipeline, step),
			Description: common.GetPipelineStatusDescription(step.State),
			Context:     common.GetPipelineStatusContext(repo, pipeline, step),
		},
	)
	return err
}

// Netrc returns a netrc file capable of authenticating Forgejo requests and
// cloning Forgejo repositories. The netrc will use the global machine account
// when configured.
func (c *Forgejo) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	login := ""
	token := ""

	if u != nil {
		login = u.Login
		token = u.Token
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
// the Forgejo repository.
func (c *Forgejo) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	config := map[string]string{
		"url":          link,
		"secret":       r.Hash,
		"content_type": "json",
	}
	hook := forgejo.CreateHookOption{
		Type:   forgejo.HookTypeForgejo,
		Config: config,
		Events: []string{"push", "create", "pull_request"},
		Active: true,
	}

	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}
	_, response, err := client.CreateRepoHook(r.Owner, r.Name, hook)
	if err != nil {
		if response != nil {
			if response.StatusCode == 404 {
				return fmt.Errorf("Could not find repository")
			}
			if response.StatusCode == 200 {
				return fmt.Errorf("Could not find repository, repository was probably renamed")
			}
		}
		return err
	}
	return nil
}

// Deactivate deactivates the repository be removing repository push hooks from
// the Forgejo repository.
func (c *Forgejo) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client, err := c.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}

	hooks, _, err := client.ListRepoHooks(r.Owner, r.Name, forgejo.ListHooksOptions{})
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
func (c *Forgejo) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	token := ""
	if u != nil {
		token = u.Token
	}
	client, err := c.newClientToken(ctx, token)
	if err != nil {
		return nil, err
	}

	branches, err := common.Paginate(func(page int) ([]string, error) {
		branches, _, err := client.ListRepoBranches(r.Owner, r.Name,
			forgejo.ListRepoBranchesOptions{ListOptions: forgejo.ListOptions{Page: page}})
		result := make([]string, len(branches))
		for i := range branches {
			result[i] = branches[i].Name
		}
		return result, err
	})
	if err != nil {
		return nil, err
	}

	return branches, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch
func (c *Forgejo) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error) {
	token := ""
	if u != nil {
		token = u.Token
	}

	client, err := c.newClientToken(ctx, token)
	if err != nil {
		return "", err
	}

	b, _, err := client.GetRepoBranch(r.Owner, r.Name, branch)
	if err != nil {
		return "", err
	}
	return b.Commit.ID, nil
}

// Hook parses the incoming Forgejo hook and returns the Repository and Pipeline
// details. If the hook is unsupported nil values are returned.
func (c *Forgejo) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	repo, pipeline, err := parseHook(r)
	if err != nil {
		return nil, nil, err
	}

	if repo == nil || pipeline == nil {
		// ignore  hook
		return nil, nil, nil
	}

	if pipeline.Event == model.EventPull && len(pipeline.ChangedFiles) == 0 {
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
func (c *Forgejo) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client, err := c.newClientToken(ctx, u.Token)
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

// helper function to return the Forgejo client with Token
func (c *Forgejo) newClientToken(ctx context.Context, token string) (*forgejo.Client, error) {
	httpClient := &http.Client{}
	if c.SkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	client, err := forgejo.NewClient(c.URL, forgejo.SetToken(token), forgejo.SetHTTPClient(httpClient), forgejo.SetContext(ctx))
	if err != nil && strings.Contains(err.Error(), "Malformed version") {
		// we guess it's a dev forgejo version
		log.Error().Err(err).Msgf("could not detect forgejo version, assume dev version %s", forgejoDevVersion)
		client, err = forgejo.NewClient(c.URL, forgejo.SetForgejoVersion(forgejoDevVersion), forgejo.SetToken(token), forgejo.SetHTTPClient(httpClient), forgejo.SetContext(ctx))
	}
	return client, err
}

// getStatus is a helper function that converts a Woodpecker
// status to a Forgejo status.
func getStatus(status model.StatusValue) forgejo.StatusState {
	switch status {
	case model.StatusPending, model.StatusBlocked:
		return forgejo.StatusPending
	case model.StatusRunning:
		return forgejo.StatusPending
	case model.StatusSuccess:
		return forgejo.StatusSuccess
	case model.StatusFailure, model.StatusError:
		return forgejo.StatusFailure
	case model.StatusKilled:
		return forgejo.StatusFailure
	case model.StatusDeclined:
		return forgejo.StatusWarning
	default:
		return forgejo.StatusFailure
	}
}

func (c *Forgejo) getChangedFilesForPR(ctx context.Context, repo *model.Repo, index int64) ([]string, error) {
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

	client, err := c.newClientToken(ctx, user.Token)
	if err != nil {
		return nil, err
	}

	if client.CheckServerVersionConstraint("1.18.0") != nil {
		// version too low
		log.Debug().Msg("Forgejo version does not support getting changed files for PRs")
		return []string{}, nil
	}

	return common.Paginate(func(page int) ([]string, error) {
		forgejoFiles, _, err := client.ListPullRequestFiles(repo.Owner, repo.Name, index,
			forgejo.ListPullRequestFilesOptions{ListOptions: forgejo.ListOptions{Page: page}})
		if err != nil {
			return nil, err
		}

		var files []string
		for _, file := range forgejoFiles {
			files = append(files, file.Filename)
		}
		return files, nil
	})
}

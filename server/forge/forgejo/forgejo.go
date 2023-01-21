// Copyright The Forgejo Authors.
// SPDX-License-Identifier: Apache-2.0
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
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/common"
	"github.com/woodpecker-ci/woodpecker/server/forge/forgejo/client"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

const (
	authorizeTokenURL = "%s/login/oauth/authorize"
	accessTokenURL    = "%s/login/oauth/access_token"
)

type Forgejo struct {
	URL          string
	ClientID     string
	ClientSecret string
	SkipVerify   bool
	PerPage      int
	logger       zerolog.Logger
}

type Opts struct {
	URL        string
	Client     string // OAuth2 Client ID
	Secret     string // OAuth2 Client Secret
	SkipVerify bool
	PerPage    int
	Debug      bool
}

func New(opts Opts) (forge.Forge, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err == nil {
		u.Host = host
	}
	logger := zerolog.New(os.Stderr).With().Caller().Logger()
	if opts.Debug {
		logger.Level(zerolog.DebugLevel)
	}
	return &Forgejo{
		URL:          opts.URL,
		ClientID:     opts.Client,
		ClientSecret: opts.Secret,
		SkipVerify:   opts.SkipVerify,
		PerPage:      opts.PerPage,
		logger:       logger,
	}, nil
}

func (f *Forgejo) Name() string {
	return "forgejo"
}

func (f *Forgejo) oauth2Config(ctx context.Context) (*oauth2.Config, context.Context) {
	return &oauth2.Config{
			ClientID:     f.ClientID,
			ClientSecret: f.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf(authorizeTokenURL, f.URL),
				TokenURL: fmt.Sprintf(accessTokenURL, f.URL),
			},
			RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
		},

		context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: f.SkipVerify},
			Proxy:           http.ProxyFromEnvironment,
		}})
}

func (f *Forgejo) Login(ctx context.Context, w http.ResponseWriter, req *http.Request) (*model.User, error) {
	config, oauth2Ctx := f.oauth2Config(ctx)

	if err := req.FormValue("error"); err != "" {
		return nil, &forge_types.AuthError{
			Err:         err,
			Description: req.FormValue("error_description"),
			URI:         req.FormValue("error_uri"),
		}
	}

	code := req.FormValue("code")
	if len(code) == 0 {
		http.Redirect(w, req, config.AuthCodeURL("woodpecker"), http.StatusSeeOther)
		return nil, nil
	}

	token, err := config.Exchange(oauth2Ctx, code)
	if err != nil {
		return nil, err
	}

	c, err := f.newClientToken(ctx, token.AccessToken)
	if err != nil {
		return nil, err
	}
	account, _, err := c.GetMyUserInfo()
	if err != nil {
		return nil, err
	}

	return &model.User{
		Token:  token.AccessToken,
		Secret: token.RefreshToken,
		Expiry: token.Expiry.UTC().Unix(),
		Login:  account.UserName,
		Email:  account.Email,
		Avatar: expandAvatar(f.URL, account.AvatarURL),
	}, nil
}

func (f *Forgejo) Auth(ctx context.Context, token, secret string) (string, error) {
	c, err := f.newClientToken(ctx, token)
	if err != nil {
		return "", err
	}
	user, _, err := c.GetMyUserInfo()
	if err != nil {
		return "", err
	}
	return user.UserName, nil
}

func (f *Forgejo) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config, oauth2Ctx := f.oauth2Config(ctx)
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

func (f *Forgejo) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	return common.Paginate(func(page int) ([]*model.Team, error) {
		orgs, _, err := c.ListMyOrgs(client.ListOptions{Page: page, PageSize: f.PerPage})
		teams := make([]*model.Team, 0, len(orgs))
		for _, org := range orgs {
			teams = append(teams, toTeam(org, f.URL))
		}
		return teams, err
	})
}

func (f *Forgejo) TeamPerm(u *model.User, org string) (*model.Perm, error) {
	return nil, nil
}

func (f *Forgejo) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	if remoteID.IsValid() {
		intID, err := strconv.ParseInt(string(remoteID), 10, 64)
		if err != nil {
			return nil, err
		}
		repo, _, err := c.GetRepoByID(intID)
		if err != nil {
			return nil, err
		}
		return toRepo(repo), nil
	}

	repo, _, err := c.GetRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return toRepo(repo), nil
}

func (f *Forgejo) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	return common.Paginate(func(page int) ([]*model.Repo, error) {
		repos, _, err := c.ListMyRepos(client.ListOptions{Page: page, PageSize: f.PerPage})
		result := make([]*model.Repo, 0, len(repos))
		for _, repo := range repos {
			result = append(result, toRepo(repo))
		}
		return result, err
	})
}

func (f *Forgejo) Perm(ctx context.Context, u *model.User, r *model.Repo) (*model.Perm, error) {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	repo, _, err := c.GetRepo(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}
	return toPerm(repo.Permissions), nil
}

func retryShaNotFound[T any](fun func() (T, error), logger zerolog.Logger, shaExists func() (bool, error)) (T, error) {
	result, retries, reasons, err := Retry(func() (T, bool, string, error) {
		result, err := fun()
		if err != nil {
			//
			// If there is an error 4xx, retry if the sha does not exist (yet).
			//
			forgejoErr, ok := err.(client.ForgejoError)
			if ok && forgejoErr.Status/100 == 4 {
				if ok, err := shaExists(); !ok {
					return result, true, "sha not found", nil
				} else if err != nil {
					return result, false, "shaExists nested error", err
				}
			}
		}
		return result, false, "", err
	}, 15)
	logger.Debug().Msgf("Retry: %d %v %v\n", retries, reasons, err)
	return result, err
}

func (f *Forgejo) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, file string) ([]byte, error) {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	cfg, err := retryShaNotFound(func() ([]byte, error) {
		content, resp, err := c.GetFile(r.Owner, r.Name, b.Commit, file)
		if err != nil {
			return nil, err
		}
		return content, c.StatusCodeToError(resp)
	}, f.logger, func() (bool, error) { return c.ShaExists(r.Owner, r.Name, b.Commit) })
	return cfg, err
}

func (f *Forgejo) paginateGetTrees(c *client.Forgejo, r *model.Repo, b *model.Pipeline) ([]client.GitEntry, error) {
	return common.Paginate(func(page int) ([]client.GitEntry, error) {
		return f.getTrees(c, page, r, b)
	})
}

func (f *Forgejo) getTrees(c *client.Forgejo, page int, r *model.Repo, b *model.Pipeline) ([]client.GitEntry, error) {
	tree, err := c.GetTrees(r.Owner, r.Name, b.Commit,
		client.GetTreeOptions{
			ListOptions: client.ListOptions{Page: page, PageSize: f.PerPage},
			Recursive:   true,
		},
	)
	if err != nil {
		return nil, err
	}
	if tree.Entries == nil {
		return []client.GitEntry{}, nil
	}
	return tree.Entries, nil
}

func (f *Forgejo) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, pipelinePath string) ([]*forge_types.FileMeta, error) {
	var configs []*forge_types.FileMeta

	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	entries, err := retryShaNotFound(func() ([]client.GitEntry, error) {
		return f.paginateGetTrees(c, r, b)
	}, f.logger, func() (bool, error) { return c.ShaExists(r.Owner, r.Name, b.Commit) })
	if err != nil {
		return nil, err
	}
	f.logger.Debug().Msgf("Configs: %v\n", configs)

	pipelinePath = path.Clean(pipelinePath) // We clean path and remove trailing slash
	pipelinePath += "/" + "*"               // construct pattern for match i.e. file in subdir
	for _, e := range entries {
		// Filter path matching pattern and type file (blob)
		if m, _ := filepath.Match(pipelinePath, e.Path); m && e.Type == "blob" {
			data, err := f.File(ctx, u, r, b, e.Path)
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

func (f *Forgejo) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, step *model.Step) error {
	c, err := f.newClientToken(ctx, user.Token)
	if err != nil {
		return err
	}

	_, _, err = c.CreateStatus(
		repo.Owner,
		repo.Name,
		pipeline.Commit,
		client.CreateStatusOption{
			State:       getStatus(step.State),
			TargetURL:   common.GetPipelineStatusLink(repo, pipeline, step),
			Description: common.GetPipelineStatusDescription(step.State),
			Context:     common.GetPipelineStatusContext(repo, pipeline, step),
		},
	)
	return err
}

func (f *Forgejo) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
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

func (f *Forgejo) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	config := map[string]string{
		"url":          link,
		"secret":       r.Hash,
		"content_type": "json",
	}
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}
	hookType := "forgejo"
	_, err = c.GetSemVer()
	if err != nil {
		if err, ok := err.(client.ForgejoError); ok && err.Status == 404 {
			// Gitea backward compatibility
			hookType = "gitea"
		} else {
			return err
		}
	}
	hook := client.CreateHookOption{
		Type:   hookType,
		Config: config,
		Events: []string{"push", "create", "pull_request"},
		Active: true,
	}

	_, response, err := c.CreateRepoHook(r.Owner, r.Name, hook)
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

func (f *Forgejo) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return err
	}

	hooks, _, err := c.ListRepoHooks(r.Owner, r.Name)
	if err != nil {
		return err
	}

	hook := matchingHooks(hooks, link)
	if hook != nil {
		_, err := c.DeleteRepoHook(r.Owner, r.Name, hook.ID)
		return err
	}

	return nil
}

func (f *Forgejo) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	token := ""
	if u != nil {
		token = u.Token
	}
	c, err := f.newClientToken(ctx, token)
	if err != nil {
		return nil, err
	}

	branches, err := common.Paginate(func(page int) ([]string, error) {
		branches, _, err := c.ListRepoBranches(r.Owner, r.Name, client.ListOptions{Page: page, PageSize: f.PerPage})
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

func (f *Forgejo) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error) {
	token := ""
	if u != nil {
		token = u.Token
	}

	c, err := f.newClientToken(ctx, token)
	if err != nil {
		return "", err
	}

	b, _, err := c.GetRepoBranch(r.Owner, r.Name, branch)
	if err != nil {
		return "", err
	}
	return b.Commit.ID, nil
}

func (f *Forgejo) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
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
		pipeline.ChangedFiles, err = f.getChangedFilesForPR(ctx, repo, index)
		if err != nil {
			f.logger.Error().Err(err).Msgf("could not get changed files for PR %s#%d", repo.FullName, index)
		}
	}

	return repo, pipeline, nil
}

func (f *Forgejo) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	c, err := f.newClientToken(ctx, u.Token)
	if err != nil {
		return nil, err
	}

	member, _, err := c.CheckOrgMembership(owner, u.Login)
	if err != nil {
		return nil, err
	}

	if !member {
		return &model.OrgPerm{}, nil
	}

	perm, _, err := c.GetOrgPermissions(owner, u.Login)
	if err != nil {
		return &model.OrgPerm{Member: member}, err
	}

	return &model.OrgPerm{Member: member, Admin: perm.IsAdmin || perm.IsOwner}, nil
}

func (f *Forgejo) newClientToken(ctx context.Context, token string) (*client.Forgejo, error) {
	httpClient := &http.Client{}
	if f.SkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}
	c, err := client.NewClient(ctx, f.logger, f.URL, token, httpClient)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getStatus(status model.StatusValue) client.StatusState {
	switch status {
	case model.StatusPending, model.StatusBlocked:
		return client.StatusPending
	case model.StatusRunning:
		return client.StatusPending
	case model.StatusSuccess:
		return client.StatusSuccess
	case model.StatusFailure, model.StatusError:
		return client.StatusFailure
	case model.StatusKilled:
		return client.StatusFailure
	case model.StatusDeclined:
		return client.StatusWarning
	default:
		return client.StatusFailure
	}
}

func (f *Forgejo) getChangedFilesForPR(ctx context.Context, repo *model.Repo, index int64) ([]string, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		f.logger.Error().Msg("could not get store from context")
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

	c, err := f.newClientToken(ctx, user.Token)
	if err != nil {
		return nil, err
	}

	return common.Paginate(func(page int) ([]string, error) {
		forgejoFiles, _, err := c.ListPullRequestFiles(repo.Owner, repo.Name, index, client.ListOptions{Page: page, PageSize: f.PerPage})
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

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

package sourcehut

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"git.sr.ht/~emersion/gqlclient"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/sourcehut/git"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/sourcehut/meta"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	authorizeTokenURL  = "%s/oauth2/authorize"
	accessTokenURL     = "%s/oauth2/access-token"
	graphqlEndpointURL = "%s/query"
)

const (
	gitWebhookPayload = `
		query {
			version {
				settings {
					sshUser
				}
			}

			webhook {
				uuid
				event

				... on GitEvent {
					__typename

					repository {
						id
						name
						visibility
						owner {
							canonicalName
						}
						access

						HEAD {
							name
						}
					}

					pusher {
						canonicalName
					}

					updates {
						ref {
							name
						}

						new {
							id
							type

							... on Commit {
								__typename
								author {
									name
									email
								}
								message
							}
						}

						diff
					}
				}
			}
		}
	`
)

type SourceHut struct {
	id                int64
	url               string
	metaURL           string
	gitURL            string
	listsURL          string
	oauth2URL         string
	oAuthClientID     string
	oAuthClientSecret string
	skipVerify        bool
}

// Opts defines configuration options.
type Opts struct {
	URL               string // SourceHut info URL (e.g. project hub)
	MetaURL           string // SourceHut meta URL
	GitURL            string // SourceHut git URL
	ListsURL          string // SourceHut lists URL
	OAuth2URL         string // User-facing SourceHut server url for OAuth2.
	OAuthClientID     string // OAuth2 Client ID
	OAuthClientSecret string // OAuth2 Client Secret
	SkipVerify        bool   // Skip ssl verification.
}

// New returns a Forge implementation that integrates with SourceHut.
// See https://sourcehut.org
func New(id int64, opts Opts) (forge.Forge, error) {
	if opts.OAuth2URL == "" {
		opts.OAuth2URL = opts.URL
	}

	return &SourceHut{
		id:                id,
		url:               opts.URL,
		metaURL:           opts.MetaURL,
		gitURL:            opts.GitURL,
		listsURL:          opts.ListsURL,
		oauth2URL:         opts.OAuth2URL,
		oAuthClientID:     opts.OAuthClientID,
		oAuthClientSecret: opts.OAuthClientSecret,
		skipVerify:        opts.SkipVerify,
	}, nil
}

// Name returns the string name of this driver.
func (c *SourceHut) Name() string {
	return "sourcehut"
}

// URL returns the root url of a configured forge.
func (c *SourceHut) URL() string {
	return c.metaURL
}

func (c *SourceHut) oauth2Config(ctx context.Context) (*oauth2.Config, context.Context) {
	return &oauth2.Config{
			ClientID:     c.oAuthClientID,
			ClientSecret: c.oAuthClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf(authorizeTokenURL, c.oauth2URL),
				TokenURL: fmt.Sprintf(accessTokenURL, c.oauth2URL),
			},
			RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
			Scopes: []string{
				"meta.sr.ht/PROFILE:RO",
				"git.sr.ht/ACLS:RO",
				"git.sr.ht/PROFILE:RO",
				"git.sr.ht/REPOSITORIES:RO",
				"git.sr.ht/OBJECTS:RO",
				"lists.sr.ht/PROFILE:RO",
				"lists.sr.ht/EMAILS:RO",
				"lists.sr.ht/LISTS:RO",
				"lists.sr.ht/PATCHES:RO",
			},
		},

		context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: c.skipVerify},
			Proxy:           http.ProxyFromEnvironment,
		}})
}

// Login authenticates an account with SourceHut. The SourceHut account details
// are returned when the user is successfully authenticated.
func (c *SourceHut) Login(ctx context.Context, req *forge_types.OAuthRequest) (*model.User, string, error) {
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

	client := c.newClientToken(ctx, c.metaURL, token.AccessToken)
	account, err := meta.FetchLoginUser(client, ctx)
	if err != nil {
		return nil, redirectURL, err
	}

	return &model.User{
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		Expiry:        token.Expiry.UTC().Unix(),
		Login:         account.Username,
		Email:         account.Email,
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(account.Id)),
	}, redirectURL, nil
}

func (c *SourceHut) newClientToken(ctx context.Context, baseURL, accessToken string) *gqlclient.Client {
	cfg, httpCtx := c.oauth2Config(ctx)
	httpClient := cfg.Client(httpCtx, &oauth2.Token{
		AccessToken: accessToken,
	})
	return gqlclient.New(fmt.Sprintf(graphqlEndpointURL, baseURL), httpClient)
}

func (c *SourceHut) Auth(ctx context.Context, token, secret string) (string, error) {
	client := c.newClientToken(ctx, c.metaURL, token)
	account, err := meta.FetchLoginUser(client, ctx)
	if err != nil {
		return "", err
	}
	return account.Username, nil
}

func (c *SourceHut) Teams(ctx context.Context, u *model.User, p *model.ListOptions) ([]*model.Team, error) {
	return []*model.Team{}, nil
}

func (c *SourceHut) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	client := c.newClientToken(ctx, c.gitURL, u.AccessToken)
	owner, name, _ = strings.Cut(string(remoteID), "/")
	user, ver, err := git.GetRepo(client, ctx, owner[1:], name)
	if err != nil {
		return nil, err
	}
	if user.Repository == nil {
		return nil, nil
	}
	return c.toRepo(user.Repository, ver), nil
}

func (c *SourceHut) Repos(ctx context.Context, u *model.User, p *model.ListOptions) ([]*model.Repo, error) {
	// TODO: Pagination on SourceHut does not work well with Woodpecker's
	// higher-level internals (and doesn't work well at all, to be frank)
	if p.Page != 1 {
		return nil, nil
	}

	client := c.newClientToken(ctx, c.gitURL, u.AccessToken)
	me, ver, err := git.GetRepos(client, ctx, nil)
	if err != nil {
		return nil, err
	}

	var repos []*model.Repo
	for _, repo := range me.Repositories.Results {
		repos = append(repos, c.toRepo(&repo, ver))
	}

	return repos, nil
}

func (c *SourceHut) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, fileName string) ([]byte, error) {
	cfg, httpCtx := c.oauth2Config(ctx)
	httpClient := cfg.Client(httpCtx, &oauth2.Token{
		AccessToken: u.AccessToken,
	})
	client := gqlclient.New(fmt.Sprintf(graphqlEndpointURL, c.gitURL), httpClient)

	owner, name, _ := strings.Cut(string(r.ForgeRemoteID), "/")
	user, err := git.GetFile(client, ctx, owner[1:], name, fileName)
	if err != nil {
		return nil, err
	}
	if user.Repository.Path == nil {
		return nil, fmt.Errorf("path %s not found", fileName)
	}

	var content string
	switch obj := user.Repository.Path.Object.Value.(type) {
	case *git.BinaryBlob:
		content = string(obj.Content)
	case *git.TextBlob:
		content = string(obj.Content)
	default:
		return nil, fmt.Errorf("path %s is not a file", fileName)
	}

	resp, err := httpClient.Get(content)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	return data, err
}

func (c *SourceHut) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, dirName string) ([]*forge_types.FileMeta, error) {
	cfg, httpCtx := c.oauth2Config(ctx)
	httpClient := cfg.Client(httpCtx, &oauth2.Token{
		AccessToken: u.AccessToken,
	})
	client := gqlclient.New(fmt.Sprintf(graphqlEndpointURL, c.gitURL), httpClient)

	owner, name, _ := strings.Cut(string(r.ForgeRemoteID), "/")
	user, err := git.GetDir(client, ctx, owner[1:], name, dirName)
	if err != nil {
		return nil, err
	}

	if user.Repository.Path == nil {
		return nil, fmt.Errorf("path %s not found", dirName)
	}

	var tree *git.Tree
	switch obj := user.Repository.Path.Object.Value.(type) {
	case *git.Tree:
		tree = obj
	default:
		return nil, fmt.Errorf("path %s is not a directory", dirName)
	}

	// TODO: Paginate this query
	var entries []*forge_types.FileMeta
	for _, ent := range tree.Entries.Results {
		var content string
		switch obj := ent.Object.Value.(type) {
		case *git.TextBlob:
			content = string(obj.Content)
		case *git.BinaryBlob:
			content = string(obj.Content)
		default:
			continue
		}

		resp, err := httpClient.Get(content)
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		entries = append(entries, &forge_types.FileMeta{
			Name: ent.Name,
			Data: data,
		})
	}

	return entries, nil
}

func (c *SourceHut) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Workflow) error {
	return nil // TODO
}

func (c *SourceHut) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	// XXX: SourceHut does not support cloning private repos over HTTP
	return &model.Netrc{
		Login:    "",
		Password: "",
		Machine:  "",
		Type:     model.ForgeTypeSourceHut,
	}, nil
}

func (c *SourceHut) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := c.newClientToken(ctx, c.gitURL, u.AccessToken)
	owner, name, _ := strings.Cut(string(r.ForgeRemoteID), "/")
	user, _, err := git.GetRepo(client, ctx, owner[1:], name)
	repo := user.Repository

	_, err = git.RegisterPushWebhook(client, ctx, repo.Id, gitWebhookPayload, link)
	if err != nil {
		return err
	}

	// XXX: Ideally we would store the webhook ID here

	return nil
}

func (c *SourceHut) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := c.newClientToken(ctx, c.gitURL, u.AccessToken)
	owner, name, _ := strings.Cut(string(r.ForgeRemoteID), "/")
	user, _, err := git.GetRepo(client, ctx, owner[1:], name)
	if err != nil {
		return err
	}

	webhooks, err := git.GetPushWebhooks(client, ctx, user.Repository.Id)
	if err != nil {
		return err
	}

	// Note: we would only ever have registered one webhook
	webhook := webhooks.Results[0]
	_, err = git.UnregisterPushWebhook(client, ctx, webhook.Id)
	return err
}

func (c *SourceHut) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	// TODO: Pagination
	if p.Page != 1 {
		return nil, nil
	}

	client := c.newClientToken(ctx, c.gitURL, u.AccessToken)
	owner, name, _ := strings.Cut(string(r.ForgeRemoteID), "/")

	var branches []string
	user, err := git.GetReferences(client, ctx, owner[1:], name, nil)
	if err != nil {
		return nil, nil
	}

	for _, ref := range user.Repository.References.Results {
		if strings.HasPrefix(ref.Name, "refs/heads/") {
			branches = append(branches, ref.Name[len("refs/heads/"):])
		}
	}

	return branches, nil
}

func (c *SourceHut) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	client := c.newClientToken(ctx, c.gitURL, u.AccessToken)
	owner, name, _ := strings.Cut(string(r.ForgeRemoteID), "/")
	user, err := git.GetHead(client, ctx, owner[1:], name, "refs/heads/"+branch)
	if err != nil {
		return nil, err
	}

	target := user.Repository.Reference.Follow
	return &model.Commit{
		SHA:      target.Id,
		ForgeURL: fmt.Sprintf("%s/%s/%s/commit/%s", c.gitURL, owner, name, target.Id),
	}, nil
}

func (c *SourceHut) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	return nil, nil // TODO
}

func (c *SourceHut) Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	type hookData struct {
		Data struct {
			Webhook *git.WebhookPayload `json:"webhook"`
			Version *git.Version        `json:"version"`
		} `json:"data"`
	}

	bytes, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	var hook hookData
	err = json.Unmarshal(bytes, &hook)
	if err != nil {
		return nil, nil, err
	}

	var gitEvent *git.GitEvent
	switch payload := hook.Data.Webhook.Value.(type) {
	case *git.GitEvent:
		gitEvent = payload
	default:
		log.Warn().Msg("Ignoring unknown webhook event")
		return nil, nil, nil
	}

	repo := c.toRepo(gitEvent.Repository, hook.Data.Version)
	pipeline := c.toPushPipeline(gitEvent)
	return repo, pipeline, nil
}

func (c *SourceHut) OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	return &model.OrgPerm{
		Member: true,
		Admin:  true,
	}, nil
}

func (c *SourceHut) Org(ctx context.Context, u *model.User, org string) (*model.Org, error) {
	client := c.newClientToken(ctx, c.metaURL, u.AccessToken)
	account, err := meta.FetchUser(client, ctx, org[1:])
	if err != nil {
		return nil, err
	}
	return &model.Org{
		Name:    account.Username,
		ForgeID: int64(account.Id),
		IsUser:  true,
		Private: false,
	}, nil
}

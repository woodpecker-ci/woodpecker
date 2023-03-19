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

package bitbucketserver

// WARNING! This is an work-in-progress patch and does not yet conform to the coding,
// quality or security standards expected of this project. Please use with caution.

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/mrjones/oauth"

	"github.com/woodpecker-ci/woodpecker/server/forge"
	"github.com/woodpecker-ci/woodpecker/server/forge/bitbucketserver/internal"
	"github.com/woodpecker-ci/woodpecker/server/forge/common"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	requestTokenURL   = "%s/plugins/servlet/oauth/request-token"
	authorizeTokenURL = "%s/plugins/servlet/oauth/authorize"
	accessTokenURL    = "%s/plugins/servlet/oauth/access-token"
)

// Opts defines configuration options.
type Opts struct {
	URL               string // Stash server url.
	Username          string // Git machine account username.
	Password          string // Git machine account password.
	ConsumerKey       string // Oauth1 consumer key.
	ConsumerRSA       string // Oauth1 consumer key file.
	ConsumerRSAString string
	SkipVerify        bool // Skip ssl verification.
}

type Config struct {
	URL        string
	Username   string
	Password   string
	SkipVerify bool
	Consumer   *oauth.Consumer
}

// New returns a Forge implementation that integrates with Bitbucket Server,
// the on-premise edition of Bitbucket Cloud, formerly known as Stash.
func New(opts Opts) (forge.Forge, error) {
	config := &Config{
		URL:        opts.URL,
		Username:   opts.Username,
		Password:   opts.Password,
		SkipVerify: opts.SkipVerify,
	}

	switch {
	case opts.Username == "":
		return nil, fmt.Errorf("Must have a git machine account username")
	case opts.Password == "":
		return nil, fmt.Errorf("Must have a git machine account password")
	case opts.ConsumerKey == "":
		return nil, fmt.Errorf("Must have a oauth1 consumer key")
	}

	if opts.ConsumerRSA == "" && opts.ConsumerRSAString == "" {
		return nil, fmt.Errorf("must have CONSUMER_RSA_KEY set to the path of a oauth1 consumer key file or CONSUMER_RSA_KEY_STRING set to the value of a oauth1 consumer key")
	}

	var keyFileBytes []byte
	if opts.ConsumerRSA != "" {
		var err error
		keyFileBytes, err = os.ReadFile(opts.ConsumerRSA)
		if err != nil {
			return nil, err
		}
	} else {
		keyFileBytes = []byte(opts.ConsumerRSAString)
	}

	block, _ := pem.Decode(keyFileBytes)
	PrivateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	config.Consumer = CreateConsumer(opts.URL, opts.ConsumerKey, PrivateKey)
	return config, nil
}

// Name returns the string name of this driver
func (c *Config) Name() string {
	return "stash"
}

func (c *Config) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	requestToken, u, err := c.Consumer.GetRequestTokenAndUrl("oob")
	if err != nil {
		return nil, err
	}
	code := req.FormValue("oauth_verifier")
	if len(code) == 0 {
		http.Redirect(res, req, u, http.StatusSeeOther)
		return nil, nil
	}
	requestToken.Token = req.FormValue("oauth_token")
	accessToken, err := c.Consumer.AuthorizeToken(requestToken, code)
	if err != nil {
		return nil, err
	}

	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, accessToken.Token)

	user, err := client.FindCurrentUser()
	if err != nil {
		return nil, err
	}

	return convertUser(user, accessToken), nil
}

// Auth is not supported by the Stash driver.
func (*Config) Auth(_ context.Context, _, _ string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

// Teams is not supported by the Stash driver.
func (*Config) Teams(_ context.Context, _ *model.User) ([]*model.Team, error) {
	var teams []*model.Team
	return teams, nil
}

// TeamPerm is not supported by the Stash driver.
func (*Config) TeamPerm(_ *model.User, _ string) (*model.Perm, error) {
	return nil, nil
}

func (c *Config) Repo(ctx context.Context, u *model.User, _ model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	repo, err := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token).FindRepo(owner, name)
	if err != nil {
		return nil, err
	}
	return convertRepo(repo), nil
}

func (c *Config) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	repos, err := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token).FindRepos()
	if err != nil {
		return nil, err
	}
	var all []*model.Repo
	for _, repo := range repos {
		all = append(all, convertRepo(repo))
	}

	return all, nil
}

func (c *Config) Perm(ctx context.Context, u *model.User, repo *model.Repo) (*model.Perm, error) {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.FindRepoPerms(repo.Owner, repo.Name)
}

func (c *Config) File(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]byte, error) {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.FindFileForRepo(r.Owner, r.Name, f, p.Ref)
}

func (c *Config) Dir(_ context.Context, _ *model.User, _ *model.Repo, _ *model.Pipeline, _ string) ([]*forge_types.FileMeta, error) {
	return nil, forge_types.ErrNotImplemented
}

// Status is not supported by the bitbucketserver driver.
func (c *Config) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, _ *model.Step) error {
	status := internal.PipelineStatus{
		State: convertStatus(pipeline.Status),
		Desc:  common.GetPipelineStatusDescription(pipeline.Status),
		Name:  fmt.Sprintf("Woodpecker #%d - %s", pipeline.Number, pipeline.Branch),
		Key:   "Woodpecker",
		URL:   common.GetPipelineStatusLink(repo, pipeline, nil),
	}

	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, user.Token)

	return client.CreateStatus(pipeline.Commit, &status)
}

func (c *Config) Netrc(_ *model.User, r *model.Repo) (*model.Netrc, error) {
	host, err := common.ExtractHostFromCloneURL(r.Clone)
	if err != nil {
		return nil, err
	}

	return &model.Netrc{
		Login:    c.Username,
		Password: c.Password,
		Machine:  host,
	}, nil
}

func (c *Config) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.CreateHook(r.Owner, r.Name, link)
}

// Branches returns the names of all branches for the named repository.
func (c *Config) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	bitbucketBranches, err := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token).ListBranches(r.Owner, r.Name)
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range bitbucketBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch
func (c *Config) BranchHead(_ context.Context, _ *model.User, _ *model.Repo, _ string) (string, error) {
	// TODO(1138): missing implementation
	return "", forge_types.ErrNotImplemented
}

func (c *Config) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)
	return client.DeleteHook(r.Owner, r.Name, link)
}

func (c *Config) Hook(_ context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error) {
	return parseHook(r, c.URL)
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *Config) OrgMembership(_ context.Context, _ *model.User, _ string) (*model.OrgPerm, error) {
	// TODO: Not implemented currently
	return nil, nil
}

func CreateConsumer(URL, ConsumerKey string, PrivateKey *rsa.PrivateKey) *oauth.Consumer {
	consumer := oauth.NewRSAConsumer(
		ConsumerKey,
		PrivateKey,
		oauth.ServiceProvider{
			RequestTokenUrl:   fmt.Sprintf(requestTokenURL, URL),
			AuthorizeTokenUrl: fmt.Sprintf(authorizeTokenURL, URL),
			AccessTokenUrl:    fmt.Sprintf(accessTokenURL, URL),
			HttpMethod:        "POST",
		})
	consumer.HttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			Proxy:           http.ProxyFromEnvironment,
		},
	}
	return consumer
}

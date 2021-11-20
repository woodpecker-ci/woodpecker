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
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/mrjones/oauth"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/bitbucketserver/internal"
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

// New returns a Remote implementation that integrates with Bitbucket Server,
// the on-premise edition of Bitbucket Cloud, formerly known as Stash.
func New(opts Opts) (remote.Remote, error) {
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
		keyFileBytes, err = ioutil.ReadFile(opts.ConsumerRSA)
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

func (c *Config) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	requestToken, u, err := c.Consumer.GetRequestTokenAndUrl("oob")
	if err != nil {
		return nil, err
	}
	var code = req.FormValue("oauth_verifier")
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
func (*Config) Auth(ctx context.Context, token, secret string) (string, error) {
	return "", fmt.Errorf("Not Implemented")
}

// Teams is not supported by the Stash driver.
func (*Config) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	var teams []*model.Team
	return teams, nil
}

// TeamPerm is not supported by the Stash driver.
func (*Config) TeamPerm(u *model.User, org string) (*model.Perm, error) {
	return nil, nil
}

func (c *Config) Repo(ctx context.Context, u *model.User, owner, name string) (*model.Repo, error) {
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

func (c *Config) Perm(ctx context.Context, u *model.User, owner, repo string) (*model.Perm, error) {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.FindRepoPerms(owner, repo)
}

func (c *Config) File(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]byte, error) {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.FindFileForRepo(r.Owner, r.Name, f, b.Ref)
}

func (c *Config) Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, f string) ([]*remote.FileMeta, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Status is not supported by the bitbucketserver driver.
func (c *Config) Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Build, link string, proc *model.Proc) error {
	status := internal.BuildStatus{
		State: convertStatus(b.Status),
		Desc:  convertDesc(b.Status),
		Name:  fmt.Sprintf("Woodpecker #%d - %s", b.Number, b.Branch),
		Key:   "Woodpecker",
		Url:   link,
	}

	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.CreateStatus(b.Commit, &status)
}

func (c *Config) Netrc(user *model.User, r *model.Repo) (*model.Netrc, error) {
	u, err := url.Parse(c.URL)
	if err != nil {
		return nil, err
	}
	//remove the port
	tmp := strings.Split(u.Host, ":")
	var host = tmp[0]

	if err != nil {
		return nil, err
	}
	return &model.Netrc{
		Machine:  host,
		Login:    c.Username,
		Password: c.Password,
	}, nil
}

func (c *Config) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)

	return client.CreateHook(r.Owner, r.Name, link)
}

// Branches returns the names of all branches for the named repository.
func (c *Config) Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error) {
	// TODO: fetch all branches
	return []string{r.Branch}, nil
}

func (c *Config) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	client := internal.NewClientWithToken(ctx, c.URL, c.Consumer, u.Token)
	return client.DeleteHook(r.Owner, r.Name, link)
}

func (c *Config) Hook(r *http.Request) (*model.Repo, *model.Build, error) {
	return parseHook(r, c.URL)
}

func CreateConsumer(URL string, ConsumerKey string, PrivateKey *rsa.PrivateKey) *oauth.Consumer {
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

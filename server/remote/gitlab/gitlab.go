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

package gitlab

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/shared/oauth2"
)

const (
	defaultScope  = "api"
	perPage       = 100
	statusContext = "ci/drone"
)

// Opts defines configuration options.
type Opts struct {
	URL          string // Gitlab server url.
	ClientID     string // Oauth2 client id.
	ClientSecret string // Oauth2 client secret.
	Username     string // Optional machine account username.
	Password     string // Optional machine account password.
	PrivateMode  bool   // Gogs is running in private mode.
	SkipVerify   bool   // Skip ssl verification.
}

// Gitlab implements "Remote" interface
type Gitlab struct {
	URL          string
	ClientID     string
	ClientSecret string
	Machine      string
	Username     string
	Password     string
	PrivateMode  bool
	SkipVerify   bool
	HideArchives bool
	Search       bool
}

// New returns a Remote implementation that integrates with Gitlab, an open
// source Git service. See https://gitlab.com
func New(opts Opts) (remote.Remote, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	host, _, err := net.SplitHostPort(u.Host)
	if err == nil {
		u.Host = host
	}
	return &Gitlab{
		URL:          opts.URL,
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		Machine:      u.Host,
		Username:     opts.Username,
		Password:     opts.Password,
		PrivateMode:  opts.PrivateMode,
		SkipVerify:   opts.SkipVerify,
	}, nil
}

// Login authenticates the session and returns the
// remote user details.
func (g *Gitlab) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	var config = &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		Scope:        defaultScope,
		AuthURL:      fmt.Sprintf("%s/oauth/authorize", g.URL),
		TokenURL:     fmt.Sprintf("%s/oauth/token", g.URL),
		RedirectURL:  fmt.Sprintf("%s/authorize", server.Config.Server.Host),
	}

	// get the OAuth errors
	if err := req.FormValue("error"); err != "" {
		return nil, &remote.AuthError{
			Err:         err,
			Description: req.FormValue("error_description"),
			URI:         req.FormValue("error_uri"),
		}
	}

	// get the OAuth code
	var code = req.FormValue("code")
	if len(code) == 0 {
		http.Redirect(res, req, config.AuthCodeURL("drone"), http.StatusSeeOther)
		return nil, nil
	}

	var trans = &oauth2.Transport{Config: config, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: g.SkipVerify},
		Proxy:           http.ProxyFromEnvironment,
	}}
	var _token, err = trans.Exchange(code)
	if err != nil {
		return nil, fmt.Errorf("Error exchanging token. %s", err)
	}

	client, err := newClient(g.URL, _token.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	login, _, err := client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Login:  login.Username,
		Email:  login.Email,
		Avatar: login.AvatarURL,
		Token:  _token.AccessToken,
		Secret: _token.RefreshToken,
	}
	if !strings.HasPrefix(user.Avatar, "http") {
		user.Avatar = g.URL + "/" + login.AvatarURL
	}

	return user, nil
}

// Auth authenticates the session and returns the remote user login for the given token
func (g *Gitlab) Auth(ctx context.Context, token, _ string) (string, error) {
	client, err := newClient(g.URL, token, g.SkipVerify)
	if err != nil {
		return "", err
	}

	login, _, err := client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return "", err
	}
	return login.Username, nil
}

// Teams fetches a list of team memberships from the remote system.
func (g *Gitlab) Teams(ctx context.Context, user *model.User) ([]*model.Team, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	teams := make([]*model.Team, 0, perPage)

	for i := 1; true; i++ {
		batch, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions:    gitlab.ListOptions{Page: i, PerPage: perPage},
			AllAvailable:   gitlab.Bool(false),
			MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions), // TODO: check what's best here
		}, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for i := range batch {
			teams = append(teams, &model.Team{
				Login:  batch[i].Name,
				Avatar: batch[i].AvatarURL,
			},
			)
		}

		if len(batch) < perPage {
			break
		}
	}

	return teams, nil
}

// getProject fetches the named repository from the remote system.
func (g *Gitlab) getProject(ctx context.Context, client *gitlab.Client, owner, name string) (*gitlab.Project, error) {
	repo, _, err := client.Projects.GetProject(fmt.Sprintf("%s/%s", owner, name), nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Repo fetches the named repository from the remote system.
func (g *Gitlab) Repo(ctx context.Context, user *model.User, owner, name string) (*model.Repo, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, owner, name)
	if err != nil {
		return nil, err
	}

	return g.convertGitlabRepo(_repo)
}

// Repos fetches a list of repos from the remote system.
func (g *Gitlab) Repos(ctx context.Context, user *model.User) ([]*model.Repo, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	repos := make([]*model.Repo, 0, perPage)
	opts := &gitlab.ListProjectsOptions{
		ListOptions:    gitlab.ListOptions{PerPage: perPage},
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions), // TODO: check what's best here
	}
	if g.HideArchives {
		opts.Archived = gitlab.Bool(false)
	}

	for i := 1; true; i++ {
		opts.Page = i
		batch, _, err := client.Projects.ListProjects(opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for i := range batch {
			repo, err := g.convertGitlabRepo(batch[i])
			if err != nil {
				return nil, err
			}
			repos = append(repos, repo)
		}

		if len(batch) < perPage {
			break
		}
	}

	return repos, err
}

// Perm fetches the named repository from the remote system.
func (g *Gitlab) Perm(ctx context.Context, user *model.User, owner, name string) (*model.Perm, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	repo, err := g.getProject(ctx, client, owner, name)
	if err != nil {
		return nil, err
	}

	// repo owner is granted full access
	if repo.Owner != nil && repo.Owner.Username == user.Login {
		return &model.Perm{Push: true, Pull: true, Admin: true}, nil
	}

	// return permission for current user
	return &model.Perm{
		Pull:  isRead(repo),
		Push:  isWrite(repo),
		Admin: isAdmin(repo),
	}, nil
}

// File fetches a file from the remote repository and returns in string format.
func (g *Gitlab) File(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, fileName string) ([]byte, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}
	file, _, err := client.RepositoryFiles.GetRawFile(_repo.ID, fileName, &gitlab.GetRawFileOptions{Ref: &build.Commit}, gitlab.WithContext(ctx))
	return file, err
}

// Dir fetches a folder from the remote repository
func (g *Gitlab) Dir(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, path string) ([]*remote.FileMeta, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	files := make([]*remote.FileMeta, 0, perPage)
	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}
	opts := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{PerPage: perPage},
		Path:        &path,
		Ref:         &build.Commit,
		Recursive:   gitlab.Bool(false),
	}

	for i := 1; true; i++ {
		opts.Page = 1
		batch, _, err := client.Repositories.ListTree(_repo.ID, opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for i := range batch {
			if batch[i].Type != "blob" { // no file
				continue
			}
			data, err := g.File(ctx, user, repo, build, batch[i].Path)
			if err != nil {
				return nil, err
			}
			files = append(files, &remote.FileMeta{
				Name: batch[i].Path,
				Data: data,
			})
		}

		if len(batch) < perPage {
			break
		}
	}

	return files, nil
}

// Status sends the commit status back to gitlab.
func (g *Gitlab) Status(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, link string, proc *model.Proc) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	_, _, err = client.Commits.SetCommitStatus(_repo.ID, build.Commit, &gitlab.SetCommitStatusOptions{
		Ref:         gitlab.String(strings.ReplaceAll(build.Ref, "refs/heads/", "")),
		State:       getStatus(build.Status),
		Description: gitlab.String(getDesc(build.Status)),
		TargetURL:   &link,
		Name:        nil,
		Context:     gitlab.String(statusContext),
	}, gitlab.WithContext(ctx))

	return err
}

// Netrc returns a netrc file capable of authenticating Gitlab requests and
// cloning Gitlab repositories. The netrc will use the global machine account
// when configured.
func (g *Gitlab) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	if g.Password != "" {
		return &model.Netrc{
			Login:    g.Username,
			Password: g.Password,
			Machine:  g.Machine,
		}, nil
	}
	return &model.Netrc{
		Login:    "oauth2",
		Password: u.Token,
		Machine:  g.Machine,
	}, nil
}

// Activate activates a repository by adding a Post-commit hook and
// a Public Deploy key, if applicable.
func (g *Gitlab) Activate(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}
	uri, err := url.Parse(link)
	if err != nil {
		return err
	}
	token := uri.Query().Get("access_token")
	webURL := fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return err
	}
	// TODO: "WoodpeckerCIService"
	_, err = client.Services.SetDroneCIService(_repo.ID, &gitlab.SetDroneCIServiceOptions{
		Token:                 &token,
		DroneURL:              &webURL,
		EnableSSLVerification: gitlab.Bool(!g.SkipVerify),
	}, gitlab.WithContext(ctx))
	return err
}

// Deactivate removes a repository by removing all the post-commit hooks
// which are equal to link and removing the SSH deploy key.
func (g *Gitlab) Deactivate(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return err
	}
	// TODO: "WoodpeckerCIService"
	_, err = client.Services.DeleteDroneCIService(_repo.ID, gitlab.WithContext(ctx))

	return err
}

// Branches returns the names of all branches for the named repository.
func (g *Gitlab) Branches(ctx context.Context, user *model.User, repo *model.Repo) ([]string, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}

	gitlabBranches, _, err := client.Branches.ListBranches(_repo.ID, &gitlab.ListBranchesOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range gitlabBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// Hook parses the post-commit hook from the Request body
// and returns the required data in a standard format.
func (g *Gitlab) Hook(req *http.Request) (*model.Repo, *model.Build, error) {
	defer req.Body.Close()
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}

	eventType := gitlab.WebhookEventType(req)
	// TODO: Fix Upstream: We get `Service Hook` - which the library do not understand
	if eventType == "Service Hook" {
		e := struct {
			ObjectKind string `json:"object_kind"`
		}{}
		if err := json.Unmarshal(payload, &e); err != nil {
			return nil, nil, err
		}
		switch e.ObjectKind {
		case "push":
			eventType = gitlab.EventTypePush
		case "tag_push":
			eventType = gitlab.EventTypeTagPush
		case "merge_request":
			eventType = gitlab.EventTypeMergeRequest
		}
	}

	parsed, err := gitlab.ParseWebhook(eventType, payload)
	if err != nil {
		return nil, nil, err
	}

	switch event := parsed.(type) {
	case *gitlab.MergeEvent:
		return convertMergeRequestHock(event, req)
	case *gitlab.PushEvent:
		return convertPushHock(event)
	case *gitlab.TagEvent:
		return convertTagHock(event)
	default:
		return nil, nil, nil
	}
}

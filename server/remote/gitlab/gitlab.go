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
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/shared/oauth2"

	"github.com/xanzy/go-gitlab"
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
func (g *Gitlab) Login(res http.ResponseWriter, req *http.Request) (*model.User, error) {
	var config = &oauth2.Config{
		ClientId:     g.ClientID,
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
	var token_, err = trans.Exchange(code)
	if err != nil {
		return nil, fmt.Errorf("Error exchanging token. %s", err)
	}

	client, err := newClient(g.URL, token_.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	login, _, err := client.Users.CurrentUser()
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Login:  login.Username,
		Email:  login.Email,
		Avatar: login.AvatarURL,
		Token:  token_.AccessToken,
		Secret: token_.RefreshToken,
	}
	if !strings.HasPrefix(user.Avatar, "http") {
		user.Avatar = g.URL + "/" + login.AvatarURL
	}

	return user, nil
}

// Auth authenticates the session and returns the remote user login for the given token
func (g *Gitlab) Auth(token, _ string) (string, error) {
	client, err := newClient(g.URL, token, g.SkipVerify)
	if err != nil {
		return "", err
	}

	login, _, err := client.Users.CurrentUser()
	if err != nil {
		return "", err
	}
	return login.Username, nil
}

// Teams fetches a list of team memberships from the remote system.
func (g *Gitlab) Teams(user *model.User) ([]*model.Team, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	teams := make([]*model.Team, 0, perPage)

	for i := 1; true; i++ {
		batch, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions:    gitlab.ListOptions{Page: i, PerPage: perPage},
			AllAvailable:   gitlab.Bool(false),
			MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions), // TODO: check whats best here
		})
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

// Repo fetches the named repository from the remote system.
func (g *Gitlab) Repo(user *model.User, owner, name string) (*model.Repo, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%s/%s", owner, name) // TODO: support nested repos
	repo_, _, err := client.Projects.GetProject(id, nil)
	if err != nil {
		return nil, err
	}

	return g.convertGitlabRepo(repo_)
}

// Repos fetches a list of repos from the remote system.
func (g *Gitlab) Repos(user *model.User) ([]*model.Repo, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	repos := make([]*model.Repo, 0, perPage)
	opts := &gitlab.ListProjectsOptions{
		ListOptions:    gitlab.ListOptions{PerPage: perPage},
		MinAccessLevel: gitlab.AccessLevel(gitlab.DeveloperPermissions), // TODO: check whats best here
	}
	if g.HideArchives {
		opts.Archived = gitlab.Bool(false)
	}

	for i := 1; true; i++ {
		opts.Page = i
		batch, _, err := client.Projects.ListProjects(opts)
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
func (g *Gitlab) Perm(user *model.User, owner, name string) (*model.Perm, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%s/%s", owner, name) // TODO: support nested repos
	repo, _, err := client.Projects.GetProject(id, nil)
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
// TODO: use io.Reader
func (g *Gitlab) File(user *model.User, repo *model.Repo, build *model.Build, fileName string) ([]byte, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	id := fmt.Sprintf("%s/%s", repo.Owner, repo.Name) // TODO: support nested repos
	file, _, err := client.RepositoryFiles.GetRawFile(id, fileName, &gitlab.GetRawFileOptions{Ref: &build.Commit})
	return file, err
}

// Dir fetches a folder from the remote repository
func (g *Gitlab) Dir(user *model.User, repo *model.Repo, build *model.Build, path string) ([]*remote.FileMeta, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	files := make([]*remote.FileMeta, 0, perPage)
	id := fmt.Sprintf("%s/%s", repo.Owner, repo.Name) // TODO: support nested repos
	opts := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{PerPage: perPage},
		Path:        &path,
		Ref:         &build.Commit,
		Recursive:   gitlab.Bool(false),
	}

	for i := 1; true; i++ {
		opts.Page = 1
		batch, _, err := client.Repositories.ListTree(id, opts)
		if err != nil {
			return nil, err
		}

		for i := range batch {
			if batch[i].Type != "blob" { // no file
				continue
			}
			data, err := g.File(user, repo, build, batch[i].Path)
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

	return nil, nil
}

// Status sends the commit status back to gitlab.
func (g *Gitlab) Status(user *model.User, repo *model.Repo, build *model.Build, link string, proc *model.Proc) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	var procID *int
	if proc != nil {
		*procID = int(proc.ID)
	}

	id := fmt.Sprintf("%s/%s", repo.Owner, repo.Name) // TODO: support nested repos
	_, _, err = client.Commits.SetCommitStatus(id, build.Commit, &gitlab.SetCommitStatusOptions{
		Ref:         gitlab.String(strings.ReplaceAll(build.Ref, "refs/heads/", "")),
		State:       getStatus(build.Status),
		Description: gitlab.String(getDesc(build.Status)),
		TargetURL:   &link,
		PipelineID:  procID,
		Name:        nil,
		Context:     gitlab.String(statusContext),
	})

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
func (g *Gitlab) Activate(user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}
	uri, err := url.Parse(link)
	if err != nil {
		return err
	}
	token := uri.Query().Get("access_token")
	webUrl := fmt.Sprintf("%s://%s", uri.Scheme, uri.Host)

	id := fmt.Sprintf("%s/%s", repo.Owner, repo.Name) // TODO: support nested repos
	// TODO: "WoodpeckerCIService"
	_, err = client.Services.SetDroneCIService(id, &gitlab.SetDroneCIServiceOptions{
		Token:                 &token,
		DroneURL:              &webUrl,
		EnableSSLVerification: gitlab.Bool(!g.SkipVerify),
	})
	return err
}

// Deactivate removes a repository by removing all the post-commit hooks
// which are equal to link and removing the SSH deploy key.
func (g *Gitlab) Deactivate(user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s/%s", repo.Owner, repo.Name) // TODO: support nested repos
	// TODO: "WoodpeckerCIService"
	_, err = client.Services.DeleteDroneCIService(id)

	return err
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

	switch eventType {
	case gitlab.EventTypeMergeRequest:
		event := parsed.(*gitlab.MergeEvent)
		return convertMergeRequestHock(event, req)
	case gitlab.EventTypePush:
		event := parsed.(*gitlab.PushEvent)
		return convertPushHock(event)
	case gitlab.EventTypeTagPush:
		event := parsed.(*gitlab.TagEvent)
		return convertTagHock(event)
	default:
		return nil, nil, nil
	}
}

const (
	DescPending  = "the build is pending"
	DescRunning  = "the buils is running"
	DescSuccess  = "the build was successful"
	DescFailure  = "the build failed"
	DescCanceled = "the build canceled"
	DescBlocked  = "the build is pending approval"
	DescDeclined = "the build was rejected"
)

// getStatus is a helper that converts a Woodpecker status to a Gitlab status.
func getStatus(status string) gitlab.BuildStateValue {
	switch status {
	case model.StatusPending, model.StatusBlocked:
		return gitlab.Pending
	case model.StatusRunning:
		return gitlab.Running
	case model.StatusSuccess:
		return gitlab.Success
	case model.StatusFailure, model.StatusError:
		return gitlab.Failed
	case model.StatusKilled:
		return gitlab.Canceled
	default:
		return gitlab.Failed
	}
}

// getDesc is a helper function that generates a description
// message for the build based on the status.
func getDesc(status string) string {
	switch status {
	case model.StatusPending:
		return DescPending
	case model.StatusRunning:
		return DescRunning
	case model.StatusSuccess:
		return DescSuccess
	case model.StatusFailure, model.StatusError:
		return DescFailure
	case model.StatusKilled:
		return DescCanceled
	case model.StatusBlocked:
		return DescBlocked
	case model.StatusDeclined:
		return DescDeclined
	default:
		return DescFailure
	}
}

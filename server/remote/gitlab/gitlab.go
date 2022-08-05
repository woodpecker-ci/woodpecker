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
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/xanzy/go-gitlab"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/common"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/shared/oauth2"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
)

const (
	defaultScope = "api"
	perPage      = 100
)

// Opts defines configuration options.
type Opts struct {
	URL          string // Gitlab server url.
	ClientID     string // Oauth2 client id.
	ClientSecret string // Oauth2 client secret.
	SkipVerify   bool   // Skip ssl verification.
}

// Gitlab implements "Remote" interface
type Gitlab struct {
	URL          string
	ClientID     string
	ClientSecret string
	SkipVerify   bool
	HideArchives bool
	Search       bool
}

// New returns a Remote implementation that integrates with Gitlab, an open
// source Git service. See https://gitlab.com
func New(opts Opts) (remote.Remote, error) {
	return &Gitlab{
		URL:          opts.URL,
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		SkipVerify:   opts.SkipVerify,
	}, nil
}

// Name returns the string name of this driver
func (g *Gitlab) Name() string {
	return "gitlab"
}

// Login authenticates the session and returns the
// remote user details.
func (g *Gitlab) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	config := &oauth2.Config{
		ClientID:     g.ClientID,
		ClientSecret: g.ClientSecret,
		Scope:        defaultScope,
		AuthURL:      fmt.Sprintf("%s/oauth/authorize", g.URL),
		TokenURL:     fmt.Sprintf("%s/oauth/token", g.URL),
		RedirectURL:  fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
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
	code := req.FormValue("code")
	if len(code) == 0 {
		authCodeURL, err := config.AuthCodeURL("drone")
		if err != nil {
			return nil, fmt.Errorf("authCodeURL error: %v", err)
		}
		http.Redirect(res, req, authCodeURL, http.StatusSeeOther)
		return nil, nil
	}

	trans := &oauth2.Transport{Config: config, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: g.SkipVerify},
		Proxy:           http.ProxyFromEnvironment,
	}}
	token, err := trans.Exchange(code)
	if err != nil {
		return nil, fmt.Errorf("Error exchanging token. %s", err)
	}

	client, err := newClient(g.URL, token.AccessToken, g.SkipVerify)
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
		Token:  token.AccessToken,
		Secret: token.RefreshToken,
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

// Repo fetches the repository from the remote system.
func (g *Gitlab) Repo(ctx context.Context, user *model.User, id, owner, name string) (*model.Repo, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	intID, err := strconv.ParseInt(id, 10, 64)
	if intID > 0 && err == nil {
		_repo, err := g.getProject(ctx, client, owner, name)
		if err != nil {
			return nil, err
		}

		return g.convertGitlabRepo(_repo)
	} else {
		repo, _, err := client.Projects.GetProject(int(intID), nil, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		return g.convertGitlabRepo(repo)
	}
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

			// TODO(648) remove when woodpecker understands nested repos
			if strings.Count(repo.FullName, "/") > 1 {
				log.Debug().Msgf("Skipping nested repository %s for user %s, because they are not supported, yet (see #648).", repo.FullName, user.Login)
				continue
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
func (g *Gitlab) Perm(ctx context.Context, user *model.User, r *model.Repo) (*model.Perm, error) {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	repo, err := g.getProject(ctx, client, r.Owner, r.Name)
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
func (g *Gitlab) Status(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build, proc *model.Proc) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	_, _, err = client.Commits.SetCommitStatus(_repo.ID, build.Commit, &gitlab.SetCommitStatusOptions{
		State:       getStatus(proc.State),
		Description: gitlab.String(common.GetBuildStatusDescription(proc.State)),
		TargetURL:   gitlab.String(common.GetBuildStatusLink(repo, build, proc)),
		Context:     gitlab.String(common.GetBuildStatusContext(repo, build, proc)),
		PipelineID:  gitlab.Int(int(build.Number)),
	}, gitlab.WithContext(ctx))

	return err
}

// Netrc returns a netrc file capable of authenticating Gitlab requests and
// cloning Gitlab repositories. The netrc will use the global machine account
// when configured.
func (g *Gitlab) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
	login := ""
	token := ""

	if u != nil {
		login = "oauth2"
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

func (g *Gitlab) getTokenAndWebURL(link string) (token, webURL string, err error) {
	uri, err := url.Parse(link)
	if err != nil {
		return "", "", err
	}
	token = uri.Query().Get("access_token")
	webURL = fmt.Sprintf("%s://%s/api/hook", uri.Scheme, uri.Host)
	return token, webURL, nil
}

// Activate activates a repository by adding a Post-commit hook and
// a Public Deploy key, if applicable.
func (g *Gitlab) Activate(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	token, webURL, err := g.getTokenAndWebURL(link)
	if err != nil {
		return err
	}

	if len(token) == 0 {
		return fmt.Errorf("no token found")
	}

	_, _, err = client.Projects.AddProjectHook(_repo.ID, &gitlab.AddProjectHookOptions{
		URL:                   gitlab.String(webURL),
		Token:                 gitlab.String(token),
		PushEvents:            gitlab.Bool(true),
		TagPushEvents:         gitlab.Bool(true),
		MergeRequestsEvents:   gitlab.Bool(true),
		DeploymentEvents:      gitlab.Bool(true),
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

	_, webURL, err := g.getTokenAndWebURL(link)
	if err != nil {
		return err
	}

	hookID := -1
	listProjectHooksOptions := &gitlab.ListProjectHooksOptions{
		PerPage: 10,
		Page:    1,
	}
	for {
		hooks, resp, err := client.Projects.ListProjectHooks(_repo.ID, listProjectHooksOptions, gitlab.WithContext(ctx))
		if err != nil {
			return err
		}

		for _, hook := range hooks {
			if hook.URL == webURL {
				hookID = hook.ID
				break
			}
		}

		// Exit the loop when we've seen all pages
		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page
		listProjectHooksOptions.Page = resp.NextPage
	}

	if hookID == -1 {
		return fmt.Errorf("could not find hook to delete")
	}

	_, err = client.Projects.DeleteProjectHook(_repo.ID, hookID, gitlab.WithContext(ctx))

	return err
}

// Branches returns the names of all branches for the named repository.
func (g *Gitlab) Branches(ctx context.Context, user *model.User, repo *model.Repo) ([]string, error) {
	token := ""
	if user != nil {
		token = user.Token
	}
	client, err := newClient(g.URL, token, g.SkipVerify)
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
func (g *Gitlab) Hook(ctx context.Context, req *http.Request) (*model.Repo, *model.Build, error) {
	defer req.Body.Close()
	payload, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}

	parsed, err := gitlab.ParseWebhook(gitlab.WebhookEventType(req), payload)
	if err != nil {
		return nil, nil, err
	}

	switch event := parsed.(type) {
	case *gitlab.MergeEvent:
		mergeIID, repo, build, err := convertMergeRequestHook(event, req)
		if err != nil {
			return nil, nil, err
		}

		if build, err = g.loadChangedFilesFromMergeRequest(ctx, repo, build, mergeIID); err != nil {
			return nil, nil, err
		}

		return repo, build, nil
	case *gitlab.PushEvent:
		return convertPushHook(event)
	case *gitlab.TagEvent:
		return convertTagHook(event)
	default:
		return nil, nil, nil
	}
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (g *Gitlab) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client, err := newClient(g.URL, u.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	groups, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
		Search: gitlab.String(owner),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	var gid int
	for _, group := range groups {
		if group.Name == owner {
			gid = group.ID
			break
		}
	}
	if gid == 0 {
		return &model.OrgPerm{}, nil
	}

	opts := &gitlab.ListGroupMembersOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 100,
		},
	}

	for i := 1; true; i++ {
		opts.Page = i
		members, _, err := client.Groups.ListAllGroupMembers(gid, opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		for _, member := range members {
			if member.Username == u.Login {
				return &model.OrgPerm{Member: true, Admin: member.AccessLevel >= gitlab.OwnerPermissions}, nil
			}
		}

		if len(members) < opts.PerPage {
			break
		}
	}

	return &model.OrgPerm{}, nil
}

func (g *Gitlab) loadChangedFilesFromMergeRequest(ctx context.Context, tmpRepo *model.Repo, build *model.Build, mergeIID int) (*model.Build, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return build, nil
	}

	repo, err := _store.GetRepoNameFallback(tmpRepo.RemoteID, tmpRepo.FullName)
	if err != nil {
		return nil, err
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, err
	}

	client, err := newClient(g.URL, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}

	changes, _, err := client.MergeRequests.GetMergeRequestChanges(_repo.ID, mergeIID, &gitlab.GetMergeRequestChangesOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(changes.Changes)*2)
	for _, file := range changes.Changes {
		files = append(files, file.NewPath, file.OldPath)
	}
	build.ChangedFiles = utils.DedupStrings(files)

	return build, nil
}

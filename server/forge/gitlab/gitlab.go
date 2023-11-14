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

package gitlab

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/server"
	"go.woodpecker-ci.org/woodpecker/server/forge"
	"go.woodpecker-ci.org/woodpecker/server/forge/common"
	forge_types "go.woodpecker-ci.org/woodpecker/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/server/model"
	"go.woodpecker-ci.org/woodpecker/server/store"
	"go.woodpecker-ci.org/woodpecker/shared/utils"
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

// Gitlab implements "Forge" interface
type GitLab struct {
	url          string
	ClientID     string
	ClientSecret string
	SkipVerify   bool
	HideArchives bool
	Search       bool
}

// New returns a Forge implementation that integrates with Gitlab, an open
// source Git service. See https://gitlab.com
func New(opts Opts) (forge.Forge, error) {
	return &GitLab{
		url:          opts.URL,
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		SkipVerify:   opts.SkipVerify,
		HideArchives: true,
	}, nil
}

// Name returns the string name of this driver
func (g *GitLab) Name() string {
	return "gitlab"
}

// URL returns the root url of a configured forge
func (g *GitLab) URL() string {
	return g.url
}

func (g *GitLab) oauth2Config(ctx context.Context) (*oauth2.Config, context.Context) {
	return &oauth2.Config{
			ClientID:     g.ClientID,
			ClientSecret: g.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("%s/oauth/authorize", g.url),
				TokenURL: fmt.Sprintf("%s/oauth/token", g.url),
			},
			Scopes:      []string{defaultScope},
			RedirectURL: fmt.Sprintf("%s/authorize", server.Config.Server.OAuthHost),
		},

		context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: g.SkipVerify},
			Proxy:           http.ProxyFromEnvironment,
		}})
}

// Login authenticates the session and returns the
// forge user details.
func (g *GitLab) Login(ctx context.Context, res http.ResponseWriter, req *http.Request) (*model.User, error) {
	config, oauth2Ctx := g.oauth2Config(ctx)

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
		http.Redirect(res, req, config.AuthCodeURL("woodpecker"), http.StatusSeeOther)
		return nil, nil
	}

	token, err := config.Exchange(oauth2Ctx, code)
	if err != nil {
		return nil, fmt.Errorf("Error exchanging token. %w", err)
	}

	client, err := newClient(g.url, token.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	login, _, err := client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Login:         login.Username,
		Email:         login.Email,
		Avatar:        login.AvatarURL,
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(login.ID)),
		Token:         token.AccessToken,
		Secret:        token.RefreshToken,
	}
	if !strings.HasPrefix(user.Avatar, "http") {
		user.Avatar = g.url + "/" + login.AvatarURL
	}

	return user, nil
}

// Refresh refreshes the Gitlab oauth2 access token. If the token is
// refreshed the user is updated and a true value is returned.
func (g *GitLab) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config, oauth2Ctx := g.oauth2Config(ctx)
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

// Auth authenticates the session and returns the forge user login for the given token
func (g *GitLab) Auth(ctx context.Context, token, _ string) (string, error) {
	client, err := newClient(g.url, token, g.SkipVerify)
	if err != nil {
		return "", err
	}

	login, _, err := client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return "", err
	}
	return login.Username, nil
}

// Teams fetches a list of team memberships from the forge.
func (g *GitLab) Teams(ctx context.Context, user *model.User) ([]*model.Team, error) {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
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

// getProject fetches the named repository from the forge.
func (g *GitLab) getProject(ctx context.Context, client *gitlab.Client, owner, name string) (*gitlab.Project, error) {
	repo, _, err := client.Projects.GetProject(fmt.Sprintf("%s/%s", owner, name), nil, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Repo fetches the repository from the forge.
func (g *GitLab) Repo(ctx context.Context, user *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	if remoteID.IsValid() {
		intID, err := strconv.ParseInt(string(remoteID), 10, 64)
		if err != nil {
			return nil, err
		}
		_repo, _, err := client.Projects.GetProject(int(intID), nil, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}
		return g.convertGitLabRepo(_repo)
	}

	_repo, err := g.getProject(ctx, client, owner, name)
	if err != nil {
		return nil, err
	}

	return g.convertGitLabRepo(_repo)
}

// Repos fetches a list of repos from the forge.
func (g *GitLab) Repos(ctx context.Context, user *model.User) ([]*model.Repo, error) {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
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
			repo, err := g.convertGitLabRepo(batch[i])
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

func (g *GitLab) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	token := common.UserToken(ctx, r, u)
	client, err := newClient(g.url, token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, r.Owner, r.Name)
	if err != nil {
		return nil, err
	}

	state := "open"
	pullRequests, _, err := client.MergeRequests.ListProjectMergeRequests(_repo.ID, &gitlab.ListProjectMergeRequestsOptions{
		ListOptions: gitlab.ListOptions{Page: p.Page, PerPage: p.PerPage},
		State:       &state,
	})
	if err != nil {
		return nil, err
	}

	result := make([]*model.PullRequest, len(pullRequests))
	for i := range pullRequests {
		result[i] = &model.PullRequest{
			Index: model.ForgeRemoteID(strconv.Itoa((pullRequests[i].ID))),
			Title: pullRequests[i].Title,
		}
	}
	return result, err
}

// File fetches a file from the forge repository and returns in string format.
func (g *GitLab) File(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, fileName string) ([]byte, error) {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}
	file, _, err := client.RepositoryFiles.GetRawFile(_repo.ID, fileName, &gitlab.GetRawFileOptions{Ref: &pipeline.Commit}, gitlab.WithContext(ctx))
	return file, err
}

// Dir fetches a folder from the forge repository
func (g *GitLab) Dir(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, path string) ([]*forge_types.FileMeta, error) {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	files := make([]*forge_types.FileMeta, 0, perPage)
	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}
	opts := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{PerPage: perPage},
		Path:        &path,
		Ref:         &pipeline.Commit,
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
			data, err := g.File(ctx, user, repo, pipeline, batch[i].Path)
			if err != nil {
				return nil, err
			}
			files = append(files, &forge_types.FileMeta{
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
func (g *GitLab) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) error {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	_, _, err = client.Commits.SetCommitStatus(_repo.ID, pipeline.Commit, &gitlab.SetCommitStatusOptions{
		State:       getStatus(workflow.State),
		Description: gitlab.String(common.GetPipelineStatusDescription(workflow.State)),
		TargetURL:   gitlab.String(common.GetPipelineStatusLink(repo, pipeline, workflow)),
		Context:     gitlab.String(common.GetPipelineStatusContext(repo, pipeline, workflow)),
	}, gitlab.WithContext(ctx))

	return err
}

// Netrc returns a netrc file capable of authenticating Gitlab requests and
// cloning Gitlab repositories. The netrc will use the global machine account
// when configured.
func (g *GitLab) Netrc(u *model.User, r *model.Repo) (*model.Netrc, error) {
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

func (g *GitLab) getTokenAndWebURL(link string) (token, webURL string, err error) {
	uri, err := url.Parse(link)
	if err != nil {
		return "", "", err
	}
	token = uri.Query().Get("access_token")
	webURL = fmt.Sprintf("%s://%s/%s", uri.Scheme, uri.Host, strings.TrimPrefix(uri.Path, "/"))
	return token, webURL, nil
}

// Activate activates a repository by adding a Post-commit hook and
// a Public Deploy key, if applicable.
func (g *GitLab) Activate(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
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
func (g *GitLab) Deactivate(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.url, user.Token, g.SkipVerify)
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
func (g *GitLab) Branches(ctx context.Context, user *model.User, repo *model.Repo, p *model.ListOptions) ([]string, error) {
	token := common.UserToken(ctx, repo, user)
	client, err := newClient(g.url, token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}

	gitlabBranches, _, err := client.Branches.ListBranches(_repo.ID,
		&gitlab.ListBranchesOptions{ListOptions: gitlab.ListOptions{Page: p.Page, PerPage: p.PerPage}},
		gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	branches := make([]string, 0)
	for _, branch := range gitlabBranches {
		branches = append(branches, branch.Name)
	}
	return branches, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch
func (g *GitLab) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error) {
	token := common.UserToken(ctx, r, u)
	client, err := newClient(g.url, token, g.SkipVerify)
	if err != nil {
		return "", err
	}

	_repo, err := g.getProject(ctx, client, r.Owner, r.Name)
	if err != nil {
		return "", err
	}

	b, _, err := client.Branches.GetBranch(_repo.ID, branch, gitlab.WithContext(ctx))
	if err != nil {
		return "", err
	}

	return b.Commit.ID, nil
}

// Hook parses the post-commit hook from the Request body
// and returns the required data in a standard format.
func (g *GitLab) Hook(ctx context.Context, req *http.Request) (*model.Repo, *model.Pipeline, error) {
	defer req.Body.Close()
	payload, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, nil, err
	}

	eventType := gitlab.WebhookEventType(req)
	parsed, err := gitlab.ParseWebhook(eventType, payload)
	if err != nil {
		return nil, nil, err
	}

	switch event := parsed.(type) {
	case *gitlab.MergeEvent:
		mergeIID, repo, pipeline, err := convertMergeRequestHook(event, req)
		if err != nil {
			return nil, nil, err
		}

		if pipeline, err = g.loadChangedFilesFromMergeRequest(ctx, repo, pipeline, mergeIID); err != nil {
			return nil, nil, err
		}

		return repo, pipeline, nil
	case *gitlab.PushEvent:
		return convertPushHook(event)
	case *gitlab.TagEvent:
		return convertTagHook(event)
	default:
		return nil, nil, &forge_types.ErrIgnoreEvent{Event: string(eventType)}
	}
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (g *GitLab) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client, err := newClient(g.url, u.Token, g.SkipVerify)
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

func (g *GitLab) Org(ctx context.Context, u *model.User, owner string) (*model.Org, error) {
	client, err := newClient(g.url, u.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	users, _, err := client.Users.ListUsers(&gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 1,
		},
		Username: gitlab.String(owner),
	})
	if len(users) == 1 && err == nil {
		return &model.Org{
			Name:    users[0].Username,
			IsUser:  true,
			Private: users[0].PrivateProfile,
		}, nil
	}

	groups, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 1,
		},
		Search: gitlab.String(owner),
	}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	if len(groups) != 1 {
		return nil, fmt.Errorf("could not find org %s", owner)
	}

	return &model.Org{
		Name:    groups[0].FullPath,
		Private: groups[0].Visibility != gitlab.PublicVisibility,
	}, nil
}

func (g *GitLab) loadChangedFilesFromMergeRequest(ctx context.Context, tmpRepo *model.Repo, pipeline *model.Pipeline, mergeIID int) (*model.Pipeline, error) {
	_store, ok := store.TryFromContext(ctx)
	if !ok {
		log.Error().Msg("could not get store from context")
		return pipeline, nil
	}

	repo, err := _store.GetRepoNameFallback(tmpRepo.ForgeRemoteID, tmpRepo.FullName)
	if err != nil {
		return nil, err
	}

	user, err := _store.GetUser(repo.UserID)
	if err != nil {
		return nil, err
	}

	client, err := newClient(g.url, user.Token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}

	changes, _, err := client.MergeRequests.ListMergeRequestDiffs(_repo.ID, mergeIID, &gitlab.ListMergeRequestDiffsOptions{}, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	files := make([]string, 0, len(changes)*2)
	for _, file := range changes {
		files = append(files, file.NewPath, file.OldPath)
	}
	pipeline.ChangedFiles = utils.DedupStrings(files)

	return pipeline, nil
}

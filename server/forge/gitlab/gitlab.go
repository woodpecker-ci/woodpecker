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
	"errors"
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

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/common"
	forge_types "go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
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
	OAuthHost    string // Public url for oauth if different from url.
}

// Gitlab implements "Forge" interface.
type GitLab struct {
	url          string
	ClientID     string
	ClientSecret string
	SkipVerify   bool
	HideArchives bool
	Search       bool
	oAuthHost    string
}

// New returns a Forge implementation that integrates with Gitlab, an open
// source Git service. See https://gitlab.com
func New(opts Opts) (forge.Forge, error) {
	return &GitLab{
		url:          opts.URL,
		ClientID:     opts.ClientID,
		ClientSecret: opts.ClientSecret,
		oAuthHost:    opts.OAuthHost,
		SkipVerify:   opts.SkipVerify,
		HideArchives: true,
	}, nil
}

// Name returns the string name of this driver.
func (g *GitLab) Name() string {
	return "gitlab"
}

// URL returns the root url of a configured forge.
func (g *GitLab) URL() string {
	return g.url
}

func (g *GitLab) oauth2Config(ctx context.Context) (*oauth2.Config, context.Context) {
	publicOAuthURL := g.oAuthHost
	if publicOAuthURL == "" {
		publicOAuthURL = g.url
	}

	return &oauth2.Config{
			ClientID:     g.ClientID,
			ClientSecret: g.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  fmt.Sprintf("%s/oauth/authorize", publicOAuthURL),
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
func (g *GitLab) Login(ctx context.Context, req *forge_types.OAuthRequest) (*model.User, string, error) {
	config, oauth2Ctx := g.oauth2Config(ctx)
	redirectURL := config.AuthCodeURL(req.State)

	// check the OAuth code
	if len(req.Code) == 0 {
		return nil, redirectURL, nil
	}

	token, err := config.Exchange(oauth2Ctx, req.Code)
	if err != nil {
		return nil, redirectURL, fmt.Errorf("error exchanging token: %w", err)
	}

	client, err := newClient(g.url, token.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, redirectURL, err
	}

	login, _, err := client.Users.CurrentUser(gitlab.WithContext(ctx))
	if err != nil {
		return nil, redirectURL, err
	}

	user := &model.User{
		Login:         login.Username,
		Email:         login.Email,
		Avatar:        login.AvatarURL,
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(login.ID)),
		AccessToken:   token.AccessToken,
		RefreshToken:  token.RefreshToken,
		Expiry:        token.Expiry.UTC().Unix(),
	}
	if !strings.HasPrefix(user.Avatar, "http") {
		user.Avatar = g.url + "/" + login.AvatarURL
	}

	return user, redirectURL, nil
}

// Refresh refreshes the Gitlab oauth2 access token. If the token is
// refreshed the user is updated and a true value is returned.
func (g *GitLab) Refresh(ctx context.Context, user *model.User) (bool, error) {
	config, oauth2Ctx := g.oauth2Config(ctx)
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

// Auth authenticates the session and returns the forge user login for the given token.
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
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	teams := make([]*model.Team, 0, perPage)

	for i := 1; true; i++ {
		batch, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions:    gitlab.ListOptions{Page: i, PerPage: perPage},
			AllAvailable:   gitlab.Ptr(false),
			MinAccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions), // TODO: check what's best here
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
func (g *GitLab) getProject(ctx context.Context, client *gitlab.Client, forgeRemoteID model.ForgeRemoteID, owner, name string) (*gitlab.Project, error) {
	var (
		repo *gitlab.Project
		err  error
	)

	if forgeRemoteID.IsValid() {
		intID, err := strconv.Atoi(string(forgeRemoteID))
		if err != nil {
			return nil, err
		}
		repo, _, err = client.Projects.GetProject(intID, nil, gitlab.WithContext(ctx))
		return repo, err
	}

	repo, _, err = client.Projects.GetProject(fmt.Sprintf("%s/%s", owner, name), nil, gitlab.WithContext(ctx))
	return repo, err
}

func (g *GitLab) getInheritedProjectMember(ctx context.Context, client *gitlab.Client, forgeRemoteID model.ForgeRemoteID, owner, name string, userID int) (*gitlab.ProjectMember, error) {
	if forgeRemoteID.IsValid() {
		intID, err := strconv.Atoi(string(forgeRemoteID))
		if err != nil {
			return nil, err
		}
		projectMember, _, err := client.ProjectMembers.GetInheritedProjectMember(intID, userID, gitlab.WithContext(ctx))
		return projectMember, err
	}

	projectMember, _, err := client.ProjectMembers.GetInheritedProjectMember(fmt.Sprintf("%s/%s", owner, name), userID, gitlab.WithContext(ctx))
	return projectMember, err
}

// Repo fetches the repository from the forge.
func (g *GitLab) Repo(ctx context.Context, user *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, remoteID, owner, name)
	if err != nil {
		return nil, err
	}

	intUserID, err := strconv.Atoi(string(user.ForgeRemoteID))
	if err != nil {
		return nil, err
	}

	projectMember, err := g.getInheritedProjectMember(ctx, client, remoteID, owner, name, intUserID)
	if err != nil {
		return nil, err
	}

	return g.convertGitLabRepo(_repo, projectMember)
}

// Repos fetches a list of repos from the forge.
func (g *GitLab) Repos(ctx context.Context, user *model.User) ([]*model.Repo, error) {
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	repos := make([]*model.Repo, 0, perPage)
	opts := &gitlab.ListProjectsOptions{
		ListOptions:    gitlab.ListOptions{PerPage: perPage},
		MinAccessLevel: gitlab.Ptr(gitlab.DeveloperPermissions), // TODO: check what's best here
	}
	if g.HideArchives {
		opts.Archived = gitlab.Ptr(false)
	}
	intUserID, err := strconv.Atoi(string(user.ForgeRemoteID))
	if err != nil {
		return nil, err
	}

	for i := 1; true; i++ {
		opts.Page = i
		batch, _, err := client.Projects.ListProjects(opts, gitlab.WithContext(ctx))
		if err != nil {
			return nil, err
		}

		for i := range batch {
			projectMember, _, err := client.ProjectMembers.GetInheritedProjectMember(batch[i].ID, intUserID, gitlab.WithContext(ctx))
			if err != nil {
				return nil, err
			}

			repo, err := g.convertGitLabRepo(batch[i], projectMember)
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

	_repo, err := g.getProject(ctx, client, r.ForgeRemoteID, r.Owner, r.Name)
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
			Index: model.ForgeRemoteID(strconv.Itoa(pullRequests[i].ID)),
			Title: pullRequests[i].Title,
		}
	}
	return result, err
}

// File fetches a file from the forge repository and returns in string format.
func (g *GitLab) File(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, fileName string) ([]byte, error) {
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}
	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}
	file, resp, err := client.RepositoryFiles.GetRawFile(_repo.ID, fileName, &gitlab.GetRawFileOptions{Ref: &pipeline.Commit}, gitlab.WithContext(ctx))
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		return nil, errors.Join(err, &forge_types.ErrConfigNotFound{Configs: []string{fileName}})
	}
	return file, err
}

// Dir fetches a folder from the forge repository.
func (g *GitLab) Dir(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, path string) ([]*forge_types.FileMeta, error) {
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	files := make([]*forge_types.FileMeta, 0, perPage)
	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		return nil, err
	}

	opts := &gitlab.ListTreeOptions{
		ListOptions: gitlab.ListOptions{PerPage: perPage},
		Path:        &path,
		Ref:         &pipeline.Commit,
		Recursive:   gitlab.Ptr(false),
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
				if errors.Is(err, &forge_types.ErrConfigNotFound{}) {
					return nil, fmt.Errorf("git tree reported existence of file but we got: %s", err.Error())
				}
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
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	_, _, err = client.Commits.SetCommitStatus(_repo.ID, pipeline.Commit, &gitlab.SetCommitStatusOptions{
		State:       getStatus(workflow.State),
		Description: gitlab.Ptr(common.GetPipelineStatusDescription(workflow.State)),
		TargetURL:   gitlab.Ptr(common.GetPipelineStatusURL(repo, pipeline, workflow)),
		Context:     gitlab.Ptr(common.GetPipelineStatusContext(repo, pipeline, workflow)),
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
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
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
		URL:                   gitlab.Ptr(webURL),
		Token:                 gitlab.Ptr(token),
		PushEvents:            gitlab.Ptr(true),
		TagPushEvents:         gitlab.Ptr(true),
		MergeRequestsEvents:   gitlab.Ptr(true),
		DeploymentEvents:      gitlab.Ptr(true),
		EnableSSLVerification: gitlab.Ptr(!g.SkipVerify),
	}, gitlab.WithContext(ctx))

	return err
}

// Deactivate removes a repository by removing all the post-commit hooks
// which are equal to link and removing the SSH deploy key.
func (g *GitLab) Deactivate(ctx context.Context, user *model.User, repo *model.Repo, link string) error {
	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return err
	}

	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
	if err != nil {
		return err
	}

	_, webURL, err := g.getTokenAndWebURL(link)
	if err != nil {
		return err
	}

	listProjectHooksOptions := &gitlab.ListProjectHooksOptions{
		PerPage: perPage,
		Page:    1,
	}
	for {
		hooks, resp, err := client.Projects.ListProjectHooks(_repo.ID, listProjectHooksOptions, gitlab.WithContext(ctx))
		if err != nil {
			return err
		}

		for _, hook := range hooks {
			if strings.Contains(hook.URL, webURL) {
				_, err = client.Projects.DeleteProjectHook(_repo.ID, hook.ID, gitlab.WithContext(ctx))
				log.Info().Msg(fmt.Sprintf("successfully deleted hook with ID %d for repo %s", hook.ID, repo.FullName))
				if err != nil {
					return err
				}
			}
		}

		if resp.CurrentPage >= resp.TotalPages {
			break
		}

		// Update the page number to get the next page
		listProjectHooksOptions.Page = resp.NextPage
	}

	return nil
}

// Branches returns the names of all branches for the named repository.
func (g *GitLab) Branches(ctx context.Context, user *model.User, repo *model.Repo, p *model.ListOptions) ([]string, error) {
	token := common.UserToken(ctx, repo, user)
	client, err := newClient(g.url, token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
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

// BranchHead returns the sha of the head (latest commit) of the specified branch.
func (g *GitLab) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error) {
	token := common.UserToken(ctx, r, u)
	client, err := newClient(g.url, token, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, r.ForgeRemoteID, r.Owner, r.Name)
	if err != nil {
		return nil, err
	}

	b, _, err := client.Branches.GetBranch(_repo.ID, branch, gitlab.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return &model.Commit{
		SHA:      b.Commit.ID,
		ForgeURL: b.Commit.WebURL,
	}, nil
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
		// https://docs.gitlab.com/ee/user/project/integrations/webhook_events.html#merge-request-events
		if event.ObjectAttributes.OldRev == "" && event.ObjectAttributes.Action != "open" && event.ObjectAttributes.Action != "close" && event.ObjectAttributes.Action != "merge" {
			return nil, nil, &forge_types.ErrIgnoreEvent{Event: string(eventType), Reason: "no code changes"}
		}
		mergeIID, repo, pipeline, err := convertMergeRequestHook(event, req)
		if err != nil {
			return nil, nil, err
		}

		if pipeline, err = g.loadChangedFilesFromMergeRequest(ctx, repo, pipeline, mergeIID); err != nil {
			return nil, nil, err
		}

		return repo, pipeline, nil
	case *gitlab.PushEvent:
		if event.TotalCommitsCount == 0 {
			return nil, nil, &forge_types.ErrIgnoreEvent{Event: string(eventType), Reason: "no commits"}
		}
		return convertPushHook(event)
	case *gitlab.TagEvent:
		return convertTagHook(event)
	case *gitlab.ReleaseEvent:
		return convertReleaseHook(event)
	default:
		return nil, nil, &forge_types.ErrIgnoreEvent{Event: string(eventType)}
	}
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (g *GitLab) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	client, err := newClient(g.url, u.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	groups, _, err := client.Groups.ListGroups(&gitlab.ListGroupsOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: perPage,
		},
		Search: gitlab.Ptr(owner),
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
			PerPage: perPage,
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
	client, err := newClient(g.url, u.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	users, _, err := client.Users.ListUsers(&gitlab.ListUsersOptions{
		ListOptions: gitlab.ListOptions{
			Page:    1,
			PerPage: 1,
		},
		Username: gitlab.Ptr(owner),
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
		Search: gitlab.Ptr(owner),
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

	client, err := newClient(g.url, user.AccessToken, g.SkipVerify)
	if err != nil {
		return nil, err
	}

	_repo, err := g.getProject(ctx, client, repo.ForgeRemoteID, repo.Owner, repo.Name)
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
	pipeline.ChangedFiles = utils.DeduplicateStrings(files)

	return pipeline, nil
}

package main

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	forge_types "github.com/woodpecker-ci/woodpecker/server/forge/types"
	"github.com/woodpecker-ci/woodpecker/server/model"
	addon_types "github.com/woodpecker-ci/woodpecker/shared/addon/types"
)

var Type = addon_types.TypeForge

func Addon(logger zerolog.Logger, env []string) (forge.Forge, error) {
	logger.Error().Msg("hello world from addon")
	return &config{l: logger}, nil
}

type config struct {
	l zerolog.Logger
}

// Name returns the string name of this driver
func (c *config) Name() string {
	c.l.Error().Msg("call Name")
	return "addon-forge"
}

// URL returns the root url of a configured forge
func (c *config) URL() string {
	c.l.Error().Msg("call URL")
	return ""
}

// Login authenticates an account with Bitbucket using the oauth2 protocol. The
// Bitbucket account details are returned when the user is successfully authenticated.
func (c *config) Login(ctx context.Context, w http.ResponseWriter, req *http.Request) (*model.User, error) {
	c.l.Error().Msg("call Login")
	return nil, nil
}

// Auth uses the Bitbucket oauth2 access token and refresh token to authenticate
// a session and return the Bitbucket account login.
func (c *config) Auth(ctx context.Context, token, secret string) (string, error) {
	c.l.Error().Msg("call Auth")
	return "", nil
}

// Teams returns a list of all team membership for the Bitbucket account.
func (c *config) Teams(ctx context.Context, u *model.User) ([]*model.Team, error) {
	c.l.Error().Msg("call Teams")
	return nil, nil
}

// Repo returns the named Bitbucket repository.
func (c *config) Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error) {
	c.l.Error().Msg("call Repo")
	return nil, nil
}

// Repos returns a list of all repositories for Bitbucket account, including
// organization repositories.
func (c *config) Repos(ctx context.Context, u *model.User) ([]*model.Repo, error) {
	c.l.Error().Msg("call Repos")
	return nil, nil
}

// File fetches the file from the Bitbucket repository and returns its contents.
func (c *config) File(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]byte, error) {
	c.l.Error().Msg("call File")
	return nil, nil
}

// Dir fetches a folder from the bitbucket repository
func (c *config) Dir(ctx context.Context, u *model.User, r *model.Repo, p *model.Pipeline, f string) ([]*forge_types.FileMeta, error) {
	c.l.Error().Msg("call Dir")
	return nil, nil
}

// Status creates a pipeline status for the Bitbucket commit.
func (c *config) Status(ctx context.Context, user *model.User, repo *model.Repo, pipeline *model.Pipeline, _ *model.Workflow) error {
	c.l.Error().Msg("call Status")
	return nil
}

// Activate activates the repository by registering repository push hooks with
// the Bitbucket repository. Prior to registering hook, previously created hooks
// are deleted.
func (c *config) Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	c.l.Error().Msg("call Activate")
	return nil
}

// Deactivate deactivates the repository be removing repository push hooks from
// the Bitbucket repository.
func (c *config) Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error {
	c.l.Error().Msg("call Deactivate")
	return nil
}

// Netrc returns a netrc file capable of authenticating Bitbucket requests and
// cloning Bitbucket repositories.
func (c *config) Netrc(u *model.User, _ *model.Repo) (*model.Netrc, error) {
	c.l.Error().Msg("call Netrc")
	return nil, nil
}

// Branches returns the names of all branches for the named repository.
func (c *config) Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error) {
	c.l.Error().Msg("call Branches")
	return nil, nil
}

// BranchHead returns the sha of the head (latest commit) of the specified branch
func (c *config) BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error) {
	c.l.Error().Msg("call BranchHead")
	return "", nil
}

// PullRequests returns the pull requests of the named repository.
func (c *config) PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error) {
	c.l.Error().Msg("call PullRequests")
	return nil, nil
}

// Hook parses the incoming Bitbucket hook and returns the Repository and
// Pipeline details. If the hook is unsupported nil values are returned.
func (c *config) Hook(_ context.Context, req *http.Request) (*model.Repo, *model.Pipeline, error) {
	c.l.Error().Msg("call Hook")
	return nil, nil, nil
}

// OrgMembership returns if user is member of organization and if user
// is admin/owner in this organization.
func (c *config) OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error) {
	c.l.Error().Msg("call OrgMembership")
	return nil, nil
}

func (c *config) Org(ctx context.Context, u *model.User, owner string) (*model.Org, error) {
	c.l.Error().Msg("call Org")
	return nil, nil
}

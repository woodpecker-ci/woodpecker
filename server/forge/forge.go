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

package forge

//go:generate mockery --name Forge --output mocks --case underscore --note "+build test"

import (
	"context"
	"net/http"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// TODO: use pagination

type Forge interface {
	// Name returns the string name of this driver
	Name() string

	// URL returns the root url of a configured forge
	URL() string

	// Login authenticates the session and returns the
	// forge user details and the URL to redirect to if not authorized yet.
	Login(ctx context.Context, r *types.OAuthRequest) (*model.User, string, error)

	// Auth authenticates the session and returns the forge user
	// login for the given token and secret
	Auth(ctx context.Context, token, secret string) (string, error)

	// Teams fetches a list of team memberships from the forge.
	Teams(ctx context.Context, u *model.User) ([]*model.Team, error)

	// Repo fetches the repository from the forge, preferred is using the ID, fallback is owner/name.
	Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error)

	// Repos fetches a list of repos from the forge.
	Repos(ctx context.Context, u *model.User) ([]*model.Repo, error)

	// File fetches a file from the forge repository and returns it in string
	// format.
	File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error)

	// Dir fetches a folder from the forge repository
	Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*types.FileMeta, error)

	// Status sends the commit status to the forge.
	// An example would be the GitHub pull request status.
	Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Workflow) error

	// Netrc returns a .netrc file that can be used to clone
	// private repositories from a forge.
	Netrc(u *model.User, r *model.Repo) (*model.Netrc, error)

	// Activate activates a repository by creating the post-commit hook.
	Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error

	// Deactivate deactivates a repository by removing all previously created
	// post-commit hooks matching the given link.
	Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error

	// Branches returns the names of all branches for the named repository.
	Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error)

	// BranchHead returns the sha of the head (latest commit) of the specified branch
	BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error)

	// PullRequests returns all pull requests for the named repository.
	PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error)

	// Hook parses the post-commit hook from the Request body and returns the
	// required data in a standard format.
	Hook(ctx context.Context, r *http.Request) (repo *model.Repo, pipeline *model.Pipeline, err error)

	// OrgMembership returns if user is member of organization and if user
	// is admin/owner in that organization.
	OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error)

	// Org fetches the organization from the forge by name. If the name is a user an org with type user is returned.
	Org(ctx context.Context, u *model.User, org string) (*model.Org, error)
}

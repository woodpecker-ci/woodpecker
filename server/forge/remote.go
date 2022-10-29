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

//go:generate go install github.com/vektra/mockery/v2@latest
//go:generate mockery --name Forge --output mocks --case underscore

import (
	"context"
	"net/http"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// TODO: use pagination
// TODO: add Driver() who return source forge back

type Forge interface {
	// Name returns the string name of this driver
	Name() string

	// Login authenticates the session and returns the
	// forge user details.
	Login(ctx context.Context, w http.ResponseWriter, r *http.Request) (*model.User, error)

	// Auth authenticates the session and returns the forge user
	// login for the given token and secret
	Auth(ctx context.Context, token, secret string) (string, error)

	// Teams fetches a list of team memberships from the forge.
	Teams(ctx context.Context, u *model.User) ([]*model.Team, error)

	// Repo fetches the repository from the forge, preferred is using the ID, fallback is owner/name.
	Repo(ctx context.Context, u *model.User, id model.ForgeID, owner, name string) (*model.Repo, error)

	// Repos fetches a list of repos from the forge.
	Repos(ctx context.Context, u *model.User) ([]*model.Repo, error)

	// Perm fetches the named repository permissions from
	// the forge for the specified user.
	Perm(ctx context.Context, u *model.User, r *model.Repo) (*model.Perm, error)

	// File fetches a file from the forge repository and returns in string
	// format.
	File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]byte, error)

	// Dir fetches a folder from the forge repository
	Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, f string) ([]*FileMeta, error)

	// Status sends the commit status to the forge.
	// An example would be the GitHub pull request status.
	Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Step) error

	// Netrc returns a .netrc file that can be used to clone
	// private repositories from a forge.
	Netrc(u *model.User, r *model.Repo) (*model.Netrc, error)

	// Activate activates a repository by creating the post-commit hook.
	Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error

	// Deactivate deactivates a repository by removing all previously created
	// post-commit hooks matching the given link.
	Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error

	// Branches returns the names of all branches for the named repository.
	// TODO: Add proper pagination handling and remove workaround in gitea forge
	Branches(ctx context.Context, u *model.User, r *model.Repo) ([]string, error)

	// BranchHead returns the sha of the head (lastest commit) of the specified branch
	BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (string, error)

	// Hook parses the post-commit hook from the Request body and returns the
	// required data in a standard format.
	Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error)

	// OrgMembership returns if user is member of organization and if user
	// is admin/owner in that organization.
	OrgMembership(ctx context.Context, u *model.User, owner string) (*model.OrgPerm, error)
}

// FileMeta represents a file in version control
type FileMeta struct {
	Name string
	Data []byte
}

type ByName []*FileMeta

func (a ByName) Len() int           { return len(a) }
func (a ByName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a ByName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// Refresher refreshes an oauth token and expiration for the given user. It
// returns true if the token was refreshed, false if the token was not refreshed,
// and error if it failed to refersh.
type Refresher interface {
	Refresh(context.Context, *model.User) (bool, error)
}

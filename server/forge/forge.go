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

// Package forge defines the Forge interface for integrating with Git hosting
// platforms (GitHub, GitLab, Gitea, Forgejo, Bitbucket, etc.).
//
// The Forge interface provides a unified abstraction for OAuth authentication,
// repository management, webhook processing, and status reporting.
package forge

import (
	"context"
	"net/http"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// Forge defines the interface for integrating with Git hosting platforms.
//
// Architecture:
// A Forge instance represents a single forge provider. Woodpecker supports
// multiple forge instances simultaneously through ForgeManager.
// Each User and Repo has a ForgeID field associating them with a specific forge.
//
// Thread Safety:
// Implementations must be safe for concurrent use. Methods receive context.Context
// for cancellation/timeout. Do not maintain user-specific state; user context is
// passed via *model.User parameter.
//
// Authentication:
// OAuth2-based authentication is assumed. Tokens are refreshed 30 minutes before
// expiry via the optional Refresher interface.
//
// Configuration Fetching:
// Pipeline configurations retrieved via File() or Dir() from Repo.Config path
// with fallback to defaults.
//
// Error Handling:
// - types.ErrIgnoreEvent: Skippable webhook events
// - types.RecordNotExist: Resource not found
// - nil Repo/Pipeline: "No action needed" (not an error).
type Forge interface {
	// Name returns the unique identifier of this forge driver.
	// Examples: "github", "gitlab", "gitea", "forgejo", "bitbucket"
	// Must be unique and constant across all implementations.
	Name() string

	// URL returns the root URL of the forge instance.
	// Examples: "https://github.com", "https://gitlab.example.com"
	URL() string

	// Login authenticates a user via OAuth2.
	//
	// OAuth Flow:
	//  1. Initial call with empty OAuthRequest.Code returns (nil, redirectURL, nil)
	//  2. User authorizes at redirectURL
	//  3. Second call with OAuthRequest.Code returns (User, redirectURL, nil)
	//
	// Returned User must contain: Login, Email, Avatar, AccessToken, RefreshToken, Expiry, ForgeRemoteID
	Login(ctx context.Context, r *types.OAuthRequest) (*model.User, string, error)

	// Auth validates an access token and returns the associated username.
	Auth(ctx context.Context, token, secret string) (string, error)

	// Teams fetches all team/organization memberships for a user.
	// May return empty slice if forge doesn't support teams/organizations.
	// Used to determine if an user is member of an team/organization.
	// Should support pagination via ListOptions.
	Teams(ctx context.Context, u *model.User, p *model.ListOptions) ([]*model.Team, error)

	// Repo fetches a single repository.
	//
	// Lookup Strategy:
	// - Prefer lookup by remoteID (forge's internal ID) if provided (more reliable as repos can be renamed)
	// - Fallback to owner/name if remoteID empty
	//
	// Must verify user has at least read access.
	// Caller must make sure ForgeID is set.
	Repo(ctx context.Context, u *model.User, remoteID model.ForgeRemoteID, owner, name string) (*model.Repo, error)

	// Repos fetches all repositories accessible to the user.
	// Should include user's permission level in Repo.Perm.
	// Should support pagination via ListOptions.
	// Caller must make sure ForgeID is set.
	Repos(ctx context.Context, u *model.User, p *model.ListOptions) ([]*model.Repo, error)

	// File fetches a single file at a specific commit.
	// Primary method for retrieving pipeline configuration files.
	// Must fetch at specific commit (b.Commit), not branch head.
	File(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, fileName string) ([]byte, error)

	// Dir fetches all files in a directory at a specific commit.
	// Supports pipeline configurations split across multiple files.
	// Should return files only.
	Dir(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, dirName string) ([]*types.FileMeta, error)

	// Status sends workflow status updates to the forge.
	// Provides visual feedback in forge UI (commit checks, PR status).
	// Failures should be logged but not block pipeline execution.
	Status(ctx context.Context, u *model.User, r *model.Repo, b *model.Pipeline, p *model.Workflow) error

	// Netrc generates .netrc credentials for cloning private repositories.
	// May receive nil user for public repos.
	Netrc(u *model.User, r *model.Repo) (*model.Netrc, error)

	// Activate creates a webhook pointing to Woodpecker.
	// Called when user activates a repository.
	// Must verify user has admin access. Should set webhook secret from r.Hash.
	// Configure webhook for all events Hook() can parse.
	Activate(ctx context.Context, u *model.User, r *model.Repo, link string) error

	// Deactivate removes the webhook.
	// Should ignore if webhook doesn't exist anymore.
	Deactivate(ctx context.Context, u *model.User, r *model.Repo, link string) error

	// Branches returns all branch names in the repository.
	// Should support pagination via ListOptions.
	Branches(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]string, error)

	// BranchHead returns the latest commit SHA for a branch.
	BranchHead(ctx context.Context, u *model.User, r *model.Repo, branch string) (*model.Commit, error)

	// PullRequests returns all open pull requests.
	// Should support pagination via ListOptions.
	PullRequests(ctx context.Context, u *model.User, r *model.Repo, p *model.ListOptions) ([]*model.PullRequest, error)

	// Hook parses incoming webhook and returns pipeline data.
	//
	// Webhook Processing Flow:
	//  1. HTTP request arrives at /api/hook with forge-specific format
	//  2. Webhook token verified against repo.Hash
	//  3. Hook() parses webhook and returns (Repo, Pipeline, error)
	//
	// Return Semantics:
	// - (repo, pipeline, nil): Execute pipeline for this event
	// - (repo, nil, nil): Valid webhook, no pipeline should run
	// - (nil, nil, types.ErrIgnoreEvent): Event ignored (logged)
	// - (nil, nil, error): Invalid webhook or parsing error
	//
	// Must verify webhook signature to prevent spoofing.
	// Should return types.ErrIgnoreEvent for non-pipeline events
	// (e.g. repository settings changed).
	Hook(ctx context.Context, r *http.Request) (*model.Repo, *model.Pipeline, error)

	// OrgMembership checks if user is member of organization and their permission.
	// Should return (Member: false, Admin: false) if not a member.
	OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error)

	// Org fetches organization details.
	// If identifier is a user, return org with IsUser: true.
	Org(ctx context.Context, u *model.User, org string) (*model.Org, error)
}

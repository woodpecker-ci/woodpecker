// Copyright 2024 Woodpecker Authors
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

package bitbucketdatacenter

import (
	"context"
	"fmt"

	bb "github.com/neticdk/go-bitbucket/bitbucket"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// OrgMembership checks if a user is a member of an organization and their permission level.
//
// Returns:
//   - Member: true, Admin: true  - if the user has admin permissions on any repository in the organization
//   - Member: true, Admin: false - if the user has write permissions (but not admin) on any repository in the organization
//   - Member: false, Admin: false - if the user has no repositories or only read permissions in the organization
func (c *client) OrgMembership(ctx context.Context, u *model.User, org string) (*model.OrgPerm, error) {
	bc, err := c.newClient(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("unable to create bitbucket client: %w", err)
	}

	return checkUserOrgPermissions(ctx, org, bc)
}

func checkUserOrgPermissions(ctx context.Context, org string, client *bb.Client) (*model.OrgPerm, error) {
	if hasAdmin, err := hasRepositoriesWithPermissionLevel(ctx, org, bb.PermissionRepoAdmin, client); err != nil {
		return nil, fmt.Errorf("failed to check for admin access: %w", err)
	} else if hasAdmin {
		return &model.OrgPerm{Member: true, Admin: true}, nil
	}

	if hasWrite, err := hasRepositoriesWithPermissionLevel(ctx, org, bb.PermissionRepoWrite, client); err != nil {
		return nil, fmt.Errorf("failed to check for write access: %w", err)
	} else if hasWrite {
		return &model.OrgPerm{Member: true, Admin: false}, nil
	}

	return &model.OrgPerm{Member: false, Admin: false}, nil
}

func hasRepositoriesWithPermissionLevel(ctx context.Context, org string, permission bb.Permission, client *bb.Client) (bool, error) {
	opts := &bb.RepositorySearchOptions{
		Archived:   bb.RepositoryArchivedActive,
		ProjectKey: org,
		Permission: permission,
	}
	repos, _, err := client.Projects.SearchRepositories(ctx, opts)
	if err != nil {
		return false, fmt.Errorf("failed to search repositories: %w", err)
	}

	return len(repos) > 0, nil
}

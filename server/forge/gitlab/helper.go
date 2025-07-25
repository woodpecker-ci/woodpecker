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
	"crypto/tls"
	"net/http"

	gitlab "gitlab.com/gitlab-org/api/client-go"
	"golang.org/x/oauth2"
)

const (
	gravatarBase = "https://www.gravatar.com/avatar"
)

// newClient is a helper function that returns a new GitHub
// client using the provided OAuth token.
func newClient(url, accessToken string, skipVerify bool) (*gitlab.Client, error) {
	return gitlab.NewAuthSourceClient(gitlab.OAuthTokenSource{
		TokenSource: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
	}, gitlab.WithBaseURL(url), gitlab.WithHTTPClient(&http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipVerify},
			Proxy:           http.ProxyFromEnvironment,
		},
	}))
}

// isRead is a helper function that returns true if the
// user has Read-only access to the repository.
func isRead(proj *gitlab.Project, projectMember *gitlab.ProjectMember) bool {
	return proj.Visibility == gitlab.InternalVisibility || proj.Visibility == gitlab.PrivateVisibility || projectMember != nil && projectMember.AccessLevel >= gitlab.ReporterPermissions
}

// isWrite is a helper function that returns true if the
// user has Read-Write access to the repository.
func isWrite(projectMember *gitlab.ProjectMember) bool {
	return projectMember != nil && projectMember.AccessLevel >= gitlab.DeveloperPermissions
}

// isAdmin is a helper function that returns true if the
// user has Admin access to the repository.
func isAdmin(projectMember *gitlab.ProjectMember) bool {
	return projectMember != nil && projectMember.AccessLevel >= gitlab.MaintainerPermissions
}

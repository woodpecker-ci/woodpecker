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

package github

import (
	"fmt"

	"github.com/google/go-github/v66/github"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

const (
	statusPending = "pending"
	statusSuccess = "success"
	statusFailure = "failure"
	statusError   = "error"
)

const (
	descPending  = "this pipeline is pending"
	descSuccess  = "the pipeline was successful"
	descFailure  = "the pipeline failed"
	descBlocked  = "the pipeline requires approval"
	descDeclined = "the pipeline was rejected"
	descError    = "oops, something went wrong"
)

const (
	headRefs  = "refs/pull/%d/head"  // pull request unmerged
	mergeRefs = "refs/pull/%d/merge" // pull request merged with base
	refSpec   = "%s:%s"
)

// convertStatus is a helper function used to convert a Woodpecker status to a
// GitHub commit status.
func convertStatus(status model.StatusValue) string {
	switch status {
	case model.StatusPending, model.StatusRunning, model.StatusBlocked, model.StatusSkipped:
		return statusPending
	case model.StatusFailure, model.StatusDeclined:
		return statusFailure
	case model.StatusSuccess:
		return statusSuccess
	default:
		return statusError
	}
}

// convertDesc is a helper function used to convert a Woodpecker status to a
// GitHub status description.
func convertDesc(status model.StatusValue) string {
	switch status {
	case model.StatusPending, model.StatusRunning:
		return descPending
	case model.StatusSuccess:
		return descSuccess
	case model.StatusFailure:
		return descFailure
	case model.StatusBlocked:
		return descBlocked
	case model.StatusDeclined:
		return descDeclined
	default:
		return descError
	}
}

// convertRepo is a helper function used to convert a GitHub repository
// structure to the common Woodpecker repository structure.
func convertRepo(from *github.Repository) *model.Repo {
	repo := &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(from.GetID())),
		Name:          from.GetName(),
		FullName:      from.GetFullName(),
		ForgeURL:      from.GetHTMLURL(),
		IsSCMPrivate:  from.GetPrivate(),
		Clone:         from.GetCloneURL(),
		CloneSSH:      from.GetSSHURL(),
		Branch:        from.GetDefaultBranch(),
		Owner:         from.GetOwner().GetLogin(),
		Avatar:        from.GetOwner().GetAvatarURL(),
		Perm:          convertPerm(from.GetPermissions()),
		SCMKind:       model.RepoGit,
		PREnabled:     true,
	}
	return repo
}

// convertPerm is a helper function used to convert a GitHub repository
// permissions to the common Woodpecker permissions structure.
func convertPerm(perm map[string]bool) *model.Perm {
	return &model.Perm{
		Admin: perm["admin"],
		Push:  perm["push"],
		Pull:  perm["pull"],
	}
}

// convertRepoList is a helper function used to convert a GitHub repository
// list to the common Woodpecker repository structure.
func convertRepoList(from []*github.Repository) []*model.Repo {
	var repos []*model.Repo
	for _, repo := range from {
		repos = append(repos, convertRepo(repo))
	}
	return repos
}

// convertTeamList is a helper function used to convert a GitHub team list to
// the common Woodpecker repository structure.
func convertTeamList(from []*github.Organization) []*model.Team {
	var teams []*model.Team
	for _, team := range from {
		teams = append(teams, convertTeam(team))
	}
	return teams
}

// convertTeam is a helper function used to convert a GitHub team structure
// to the common Woodpecker repository structure.
func convertTeam(from *github.Organization) *model.Team {
	return &model.Team{
		Login:  from.GetLogin(),
		Avatar: from.GetAvatarURL(),
	}
}

// convertRepoHook is a helper function used to extract the Repository details
// from a webhook and convert to the common Woodpecker repository structure.
func convertRepoHook(eventRepo *github.PushEventRepository) *model.Repo {
	repo := &model.Repo{
		ForgeRemoteID: model.ForgeRemoteID(fmt.Sprint(eventRepo.GetID())),
		Owner:         eventRepo.GetOwner().GetLogin(),
		Name:          eventRepo.GetName(),
		FullName:      eventRepo.GetFullName(),
		ForgeURL:      eventRepo.GetHTMLURL(),
		IsSCMPrivate:  eventRepo.GetPrivate(),
		Clone:         eventRepo.GetCloneURL(),
		CloneSSH:      eventRepo.GetSSHURL(),
		Branch:        eventRepo.GetDefaultBranch(),
		SCMKind:       model.RepoGit,
		PREnabled:     true,
	}
	if repo.FullName == "" {
		repo.FullName = repo.Owner + "/" + repo.Name
	}
	return repo
}

// convertLabels is a helper function used to convert a GitHub label list to
// the common Woodpecker label structure.
func convertLabels(from []*github.Label) []string {
	labels := make([]string, len(from))
	for i, label := range from {
		labels[i] = *label.Name
	}
	return labels
}

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

package gitea

import (
	"io"
	"net/http"
	"strings"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	hookEvent       = "X-Gitea-Event"
	hookPush        = "push"
	hookCreated     = "create"
	hookPullRequest = "pull_request"

	actionOpen = "opened"
	actionSync = "synchronized"

	stateOpen = "open"

	refBranch = "branch"
	refTag    = "tag"
)

// parseHook parses a Gitea hook from an http.Request request and returns
// Repo and Build detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request) (*model.Repo, *model.Build, error) {
	switch r.Header.Get(hookEvent) {
	case hookPush:
		return parsePushHook(r.Body)
	case hookCreated:
		return parseCreatedHook(r.Body)
	case hookPullRequest:
		return parsePullRequestHook(r.Body)
	}
	return nil, nil, nil
}

// parsePushHook parses a push hook and returns the Repo and Build details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(payload io.Reader) (repo *model.Repo, build *model.Build, err error) {
	push, err := parsePush(payload)
	if err != nil {
		return nil, nil, err
	}

	// ignore push events for tags
	if strings.HasPrefix(push.Ref, "refs/tags/") {
		return nil, nil, nil
	}

	// is this even needed?
	if push.RefType == refBranch {
		return nil, nil, nil
	}

	repo = repoFromPush(push)
	build = buildFromPush(push)
	return repo, build, err
}

// parseCreatedHook parses a push hook and returns the Repo and Build details.
// If the commit type is unsupported nil values are returned.
func parseCreatedHook(payload io.Reader) (repo *model.Repo, build *model.Build, err error) {
	push, err := parsePush(payload)
	if err != nil {
		return nil, nil, err
	}

	if push.RefType != refTag {
		return nil, nil, nil
	}

	repo = repoFromPush(push)
	build = buildFromTag(push)
	return repo, build, nil
}

// parsePullRequestHook parses a pull_request hook and returns the Repo and Build details.
func parsePullRequestHook(payload io.Reader) (*model.Repo, *model.Build, error) {
	var (
		repo  *model.Repo
		build *model.Build
	)

	pr, err := parsePullRequest(payload)
	if err != nil {
		return nil, nil, err
	}

	// Don't trigger builds for non-code changes, or if PR is not open
	if pr.Action != actionOpen && pr.Action != actionSync {
		return nil, nil, nil
	}
	if pr.PullRequest.State != stateOpen {
		return nil, nil, nil
	}

	repo = repoFromPullRequest(pr)
	build = buildFromPullRequest(pr)
	return repo, build, err
}

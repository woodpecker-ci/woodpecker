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

package gogs

import (
	"io"
	"net/http"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	hookEvent       = "X-Gogs-Event"
	hookPush        = "push"
	hookCreated     = "create"
	hookPullRequest = "pull_request"

	actionOpen = "opened"
	actionSync = "synchronized"

	stateOpen = "open"

	refBranch = "branch"
	refTag    = "tag"
)

// parseHook parses a Bitbucket hook from an http.Request request and returns
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request, privateMode bool) (*model.Repo, *model.Pipeline, error) {
	switch r.Header.Get(hookEvent) {
	case hookPush:
		return parsePushHook(r.Body, privateMode)
	case hookCreated:
		return parseCreatedHook(r.Body, privateMode)
	case hookPullRequest:
		return parsePullRequestHook(r.Body, privateMode)
	}
	return nil, nil, nil
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(payload io.Reader, privateMode bool) (*model.Repo, *model.Pipeline, error) {
	var (
		repo  *model.Repo
		build *model.Pipeline
	)

	push, err := parsePush(payload)
	if err != nil {
		return nil, nil, err
	}

	// is this even needed?
	if push.RefType == refBranch {
		return nil, nil, nil
	}

	repo = toRepo(push.Repo, privateMode)
	build = buildFromPush(push)
	return repo, build, err
}

// parseCreatedHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parseCreatedHook(payload io.Reader, privateMode bool) (*model.Repo, *model.Pipeline, error) {
	var (
		repo  *model.Repo
		build *model.Pipeline
	)

	push, err := parsePush(payload)
	if err != nil {
		return nil, nil, err
	}

	if push.RefType != refTag {
		return nil, nil, nil
	}

	repo = toRepo(push.Repo, privateMode)
	build = buildFromTag(push)
	return repo, build, err
}

// parsePullRequestHook parses a pull_request hook and returns the Repo and Pipeline details.
func parsePullRequestHook(payload io.Reader, privateMode bool) (*model.Repo, *model.Pipeline, error) {
	var (
		repo  *model.Repo
		build *model.Pipeline
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

	repo = toRepo(pr.Repo, privateMode)
	build = buildFromPullRequest(pr)
	return repo, build, err
}

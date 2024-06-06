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

package forgejo

import (
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

const (
	hookEvent       = "X-Forgejo-Event"
	hookPush        = "push"
	hookCreated     = "create"
	hookPullRequest = "pull_request"
	hookRelease     = "release"

	actionOpen  = "opened"
	actionSync  = "synchronized"
	actionClose = "closed"

	refBranch = "branch"
	refTag    = "tag"
)

// parseHook parses a Forgejo hook from an http.Request and returns
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request) (*model.Repo, *model.Pipeline, error) {
	hookType := r.Header.Get(hookEvent)
	switch hookType {
	case hookPush:
		return parsePushHook(r.Body)
	case hookCreated:
		return parseCreatedHook(r.Body)
	case hookPullRequest:
		return parsePullRequestHook(r.Body)
	case hookRelease:
		return parseReleaseHook(r.Body)
	}
	log.Debug().Msgf("unsupported hook type: '%s'", hookType)
	return nil, nil, &types.ErrIgnoreEvent{Event: hookType}
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(payload io.Reader) (repo *model.Repo, pipeline *model.Pipeline, err error) {
	push, err := parsePush(payload)
	if err != nil {
		return nil, nil, err
	}

	// ignore push events for tags
	if strings.HasPrefix(push.Ref, "refs/tags/") {
		return nil, nil, nil
	}

	// TODO is this even needed?
	if push.RefType == refBranch {
		return nil, nil, nil
	}

	repo = toRepo(push.Repo)
	pipeline = pipelineFromPush(push)
	return repo, pipeline, err
}

// parseCreatedHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parseCreatedHook(payload io.Reader) (repo *model.Repo, pipeline *model.Pipeline, err error) {
	push, err := parsePush(payload)
	if err != nil {
		return nil, nil, err
	}

	if push.RefType != refTag {
		return nil, nil, nil
	}

	repo = toRepo(push.Repo)
	pipeline = pipelineFromTag(push)
	return repo, pipeline, nil
}

// parsePullRequestHook parses a pull_request hook and returns the Repo and Pipeline details.
func parsePullRequestHook(payload io.Reader) (*model.Repo, *model.Pipeline, error) {
	var (
		repo     *model.Repo
		pipeline *model.Pipeline
	)

	pr, err := parsePullRequest(payload)
	if err != nil {
		return nil, nil, err
	}

	// Don't trigger pipelines for non-code changes ...
	if pr.Action != actionOpen && pr.Action != actionSync && pr.Action != actionClose {
		log.Debug().Msgf("pull_request action is '%s' and no open or sync", pr.Action)
		return nil, nil, nil
	}

	repo = toRepo(pr.Repo)
	pipeline = pipelineFromPullRequest(pr)
	return repo, pipeline, err
}

// parseReleaseHook parses a release hook and returns the Repo and Pipeline details.
func parseReleaseHook(payload io.Reader) (*model.Repo, *model.Pipeline, error) {
	var (
		repo     *model.Repo
		pipeline *model.Pipeline
	)

	release, err := parseRelease(payload)
	if err != nil {
		return nil, nil, err
	}

	repo = toRepo(release.Repo)
	pipeline = pipelineFromRelease(release)
	return repo, pipeline, err
}

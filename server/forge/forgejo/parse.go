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

package forgejo

import (
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

const (
	hookEvent       = "X-Forgejo-Event"
	hookPush        = "push"
	hookCreated     = "create"
	hookPullRequest = "pull_request"

	actionOpen = "opened"
	actionSync = "synchronized"

	stateOpen = "open"

	refBranch = "branch"
	refTag    = "tag"
)

// parseHook parses a Forgejo hook from an http.Request request and returns
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request) (*model.Repo, *model.Pipeline, error) {
	event := hookEvent
	hookType := r.Header.Get(event)
	if hookType == "" {
		// Gitea backward compatibility
		event = "X-Gitea-Event"
		hookType = r.Header.Get(event)
	}
	switch hookType {
	case hookPush:
		return parsePushHook(r.Body)
	case hookCreated:
		return parseCreatedHook(r.Body)
	case hookPullRequest:
		return parsePullRequestHook(r.Body)
	}
	log.Debug().Msgf("unsuported hook type from %s: '%s'", event, hookType)
	return nil, nil, nil
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
	if pr.Action != actionOpen && pr.Action != actionSync {
		log.Debug().Msgf("pull_request action is '%s' and no open or sync", pr.Action)
		return nil, nil, nil
	}
	// ... or if PR is not open
	if pr.PullRequest.State != stateOpen {
		log.Debug().Msg("pull_request is closed")
		return nil, nil, nil
	}

	repo = toRepo(pr.Repo)
	pipeline = pipelineFromPullRequest(pr)
	return repo, pipeline, err
}

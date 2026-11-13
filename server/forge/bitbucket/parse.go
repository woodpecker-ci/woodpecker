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

package bitbucket

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	hookEvent        = "X-Event-Key"
	hookPush         = "repo:push"
	hookPullCreated  = "pullrequest:created"
	hookPullUpdated  = "pullrequest:updated"
	hookPullMerged   = "pullrequest:fulfilled"
	hookPullDeclined = "pullrequest:rejected"
	stateClosed      = "MERGED"
	stateDeclined    = "DECLINED"
)

var errIncompleteHook = errors.New("incomplete webhook payload")

// validateRepo rejects a payload whose repository is missing the fields the
// conversion needs. The full name carries the owner and the repository name,
// so without a separator there is nothing to split.
func validateRepo(repo *internal.Repo) error {
	if !strings.Contains(repo.FullName, "/") {
		return errIncompleteHook
	}
	return nil
}

// parseHook parses a Bitbucket hook from an http.Request request and returns Pull Request,
// Repo and Pipeline details: one pipeline for a pull request, one per pushed branch or tag for a
// push. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request) (*internal.PullRequestHook, *model.Repo, []*model.Pipeline, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, nil, err
	}

	hookType := r.Header.Get(hookEvent)
	switch hookType {
	case hookPush:
		r, pls, err := parsePushHook(payload)
		return nil, r, pls, err
	case hookPullCreated, hookPullUpdated, hookPullMerged, hookPullDeclined:
		pr, r, pl, err := parsePullHook(payload)
		if err != nil {
			return nil, nil, nil, err
		}
		return pr, r, []*model.Pipeline{pl}, nil
	default:
		return nil, nil, nil, &types.ErrIgnoreEvent{Event: hookType}
	}
}

// parsePushHook parses a push hook and returns the Repo and one Pipeline per change.
//
// Bitbucket sends a single repo:push webhook for everything one `git push` moves, with one entry in
// push.changes per branch or tag. Taking only the first entry would silently drop every other ref
// pushed with it, so each change with a new commit becomes a pipeline of its own. A change without
// one (a deleted branch or tag) starts nothing. If no change has a new commit it returns an
// ErrIgnoreEvent error.
func parsePushHook(payload []byte) (*model.Repo, []*model.Pipeline, error) {
	hook := internal.PushHook{}

	err := json.Unmarshal(payload, &hook)
	if err != nil {
		return nil, nil, err
	}

	var pipelines []*model.Pipeline
	for _, change := range hook.Push.Changes {
		if change.New.Target.Hash == "" {
			continue
		}
		pipelines = append(pipelines, convertPushHook(&hook, &change))
	}
	if len(pipelines) == 0 {
		return nil, nil, &types.ErrIgnoreEvent{Event: "push", Reason: "BB reports no Changes"}
	}
	if err := validateRepo(&hook.Repo); err != nil {
		return nil, nil, err
	}
	return convertRepo(&hook.Repo, &internal.RepoPerm{}), pipelines, nil
}

// parsePullHook parses a pull request hook and returns the Pull Request, Repo and Pipeline
// details.
func parsePullHook(payload []byte) (*internal.PullRequestHook, *model.Repo, *model.Pipeline, error) {
	hook := internal.PullRequestHook{}

	if err := json.Unmarshal(payload, &hook); err != nil {
		return nil, nil, nil, err
	}
	if err := validateRepo(&hook.Repo); err != nil {
		return nil, nil, nil, err
	}

	return &hook, convertRepo(&hook.Repo, &internal.RepoPerm{}), convertPullHook(&hook), nil
}

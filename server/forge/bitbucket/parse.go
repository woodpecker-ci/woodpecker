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
	"io"
	"net/http"
	"slices"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/bitbucket/internal"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	hookEvent = "X-Event-Key"

	hookPush                      = "repo:push"
	hookPullCreated               = "pullrequest:created"
	hookPullUpdated               = "pullrequest:updated"
	hookPullMerged                = "pullrequest:fulfilled"
	hookPullDeclined              = "pullrequest:rejected"
	hookPullApproved              = "pullrequest:approved"
	hookPullChangesRequestCreated = "pullrequest:changes_request_created"
	hookPullChangesRequestRemoved = "pullrequest:changes_request_removed"
	hookPullCommentCreated        = "pullrequest:comment_created"
	hookPullCommentDeleted        = "pullrequest:comment_deleted"
	hookPullCommentReopened       = "pullrequest:comment_reopened"
	hookPullCommentResolved       = "pullrequest:comment_resolved"
	hookPullCommentUpdated        = "pullrequest:comment_updated"
	hookPullPush                  = "pullrequest:push"
	hookPullUnapproved            = "pullrequest:unapproved"

	stateOpen     = "OPEN"
	stateClosed   = "MERGED"
	stateDeclined = "DECLINED"
)

var supportedHookEvents = []string{
	hookPush,
	hookPullApproved,
	hookPullChangesRequestCreated,
	hookPullChangesRequestRemoved,
	hookPullCommentCreated,
	hookPullCommentDeleted,
	hookPullCommentReopened,
	hookPullCommentResolved,
	hookPullCommentUpdated,
	hookPullCreated,
	hookPullMerged,
	hookPullPush,
	hookPullDeclined,
	hookPullUnapproved,
	hookPullUpdated,
}

// parseHook parses a Bitbucket hook from an http.Request request and returns
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request) (*model.Repo, *model.Pipeline, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	hookType := r.Header.Get(hookEvent)
	if !slices.Contains(supportedHookEvents, hookType) {
		return nil, nil, &types.ErrIgnoreEvent{Event: hookType}
	}

	switch hookType {
	case hookPush:
		return parsePushHook(payload)
	default:
		return parsePullHook(payload)
	}
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(payload []byte) (*model.Repo, *model.Pipeline, error) {
	hook := internal.PushHook{}

	err := json.Unmarshal(payload, &hook)
	if err != nil {
		return nil, nil, err
	}

	for _, change := range hook.Push.Changes {
		if change.New.Target.Hash == "" {
			continue
		}
		return convertWebhookRepo(&hook.Repo, &internal.RepoPerm{}), convertPushHook(&hook, &change), nil
	}
	return nil, nil, nil
}

// parsePullHook parses a pull request hook and returns the Repo and Pipeline
// details.
func parsePullHook(payload []byte) (*model.Repo, *model.Pipeline, error) {
	hook := internal.PullRequestHook{}

	if err := json.Unmarshal(payload, &hook); err != nil {
		return nil, nil, err
	}

	return convertWebhookRepo(&hook.Repo, &internal.RepoPerm{}), convertPullHook(&hook), nil
}

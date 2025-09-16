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
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

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

	stateMerged = "MERGED"
	stateOpen   = "OPEN"
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

type parsedHookMetadata struct {
	RepoUUID,
	RepoOwner,
	RepoName,
	RepoFullName string

	NeedPostProcessing bool

	DiffStatApi string
}

// parseHook parses a Bitbucket hook from an http.Request request and returns
// Repo and Pipeline detail.
// NOTE: the "pullrequest:updated" event will be pre-parsed but needs post processing via API query and DB query!
func parseHook(r *http.Request) (*parsedHookMetadata, *model.Pipeline, error) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, nil, err
	}

	hookType := r.Header.Get(hookEvent)
	if !slices.Contains(supportedHookEvents, hookType) {
		return nil, nil, &types.ErrIgnoreEvent{Event: hookType}
	}

	if hookType == hookPush {
		return parsePushHook(payload)
	}
	// ok it must be an pullrequest:* event

	hookRepo, hookPull, p, err := parsePullHook(payload)
	if err != nil {
		return nil, nil, err
	}

	needPostProcessing := false

	// detect pull event type
	switch hookType {
	case hookPullCreated:
		p.Event = model.EventPull

	case hookPullMerged, hookPullDeclined:
		p.Event = model.EventPullClosed

	case hookPullUpdated:
		// first we only care about open pulls get updated
		if hookPull.PullRequest.State != stateOpen {
			return nil, nil, &types.ErrIgnoreEvent{
				Event:  hookType,
				Reason: fmt.Sprintf("pull state not %s but %s", stateOpen, hookPull.PullRequest.State),
			}
		}

		p.Event = model.EventPull
		// we need more info via api so we just pass it as task onto our caller
		needPostProcessing = true
	default:
		// first we only care about open pulls
		if hookPull.PullRequest.State != stateOpen {
			return nil, nil, &types.ErrIgnoreEvent{
				Event:  hookType,
				Reason: fmt.Sprintf("pull state not %s but %s", stateOpen, hookPull.PullRequest.State),
			}
		}

		p.Event = model.EventPullMetadata
		p.EventReason = []string{strings.TrimPrefix(hookType, "pullrequest:")}
	}

	return &parsedHookMetadata{
		RepoUUID:     hookRepo.UUID,
		RepoOwner:    hookRepo.Owner.Nickname,
		RepoName:     hookRepo.Name,
		RepoFullName: hookRepo.FullName,

		DiffStatApi: getLink(hookPull.PullRequest.Links, linkKeyDiffStat),

		NeedPostProcessing: needPostProcessing,
	}, p, nil
}

// parsePushHook parses a push hook and returns the Repo and Pipeline details.
// If the commit type is unsupported nil values are returned.
func parsePushHook(payload []byte) (*parsedHookMetadata, *model.Pipeline, error) {
	hook := internal.PushHook{}

	err := json.Unmarshal(payload, &hook)
	if err != nil {
		return nil, nil, err
	}

	for _, change := range hook.Push.Changes {
		if change.New.Target.Hash == "" {
			continue
		}
		return &parsedHookMetadata{
			RepoUUID:     hook.Repo.UUID,
			RepoOwner:    hook.Repo.Owner.Nickname,
			RepoName:     hook.Repo.Name,
			RepoFullName: hook.Repo.FullName,
		}, convertPushHook(&hook, &change), nil
	}
	return nil, nil, &types.ErrIgnoreEvent{Event: hookPush, Reason: "no changes detected"}
}

// parsePullHook parses a pull request hook and returns the Repo and Pipeline
// details.
func parsePullHook(payload []byte) (*internal.WebhookRepo, *internal.PullRequestHook, *model.Pipeline, error) {
	hook := internal.PullRequestHook{}

	if err := json.Unmarshal(payload, &hook); err != nil {
		return nil, nil, nil, err
	}

	return &hook.Repo, &hook, convertPullHook(&hook), nil
}

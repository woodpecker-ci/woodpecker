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

package atomgit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	hookEvent        = "X-AtomGit-Event"
	hookPush         = "push"
	hookTagPush      = "tag_push"
	hookMergeRequest = "merge_request"

	actionOpen   = "merge_request_open"
	actionClose  = "merge_request_close"
	actionMerge  = "merge_request_merge"
	actionUpdate = "merge_request_update"
	actionReopen = "merge_request_reopen"

	refBranch = "branch"
	refTag    = "tag"
)

// parseHook parses a AtomGit webhook from an http.Request and returns the
// Repo and Pipeline detail. If a hook type is unsupported nil values are returned.
func parseHook(r *http.Request) (*model.Repo, *model.Pipeline, error) {
	hookType := r.Header.Get(hookEvent)
	switch hookType {
	case hookPush:
		return parsePushHook(r.Body)
	case hookTagPush:
		return parseTagPushHook(r.Body)
	case hookMergeRequest:
		return parseMergeRequestHook(r.Body)
	}
	log.Debug().Msgf("unsupported atomgit hook type: '%s'", hookType)
	return nil, nil, &types.ErrIgnoreEvent{Event: hookType}
}

func parsePushHook(payload io.Reader) (*model.Repo, *model.Pipeline, error) {
	hook := new(pushHook)
	if err := json.NewDecoder(payload).Decode(hook); err != nil {
		return nil, nil, err
	}

	if hook.Repository == nil {
		return nil, nil, fmt.Errorf("parsed push webhook does not contain repository info")
	}

	// ignore tag pushes handled by tag_push event
	if strings.HasPrefix(hook.Ref, "refs/tags/") {
		return nil, nil, nil
	}

	repo := toRepo(hook.Repository)
	pipeline := pipelineFromPush(hook)
	return repo, pipeline, nil
}

func parseTagPushHook(payload io.Reader) (*model.Repo, *model.Pipeline, error) {
	hook := new(tagPushHook)
	if err := json.NewDecoder(payload).Decode(hook); err != nil {
		return nil, nil, err
	}

	if hook.Repository == nil {
		return nil, nil, fmt.Errorf("parsed tag push webhook does not contain repository info")
	}

	repo := toRepo(hook.Repository)
	pipeline := pipelineFromTag(hook)
	return repo, pipeline, nil
}

func parseMergeRequestHook(payload io.Reader) (*model.Repo, *model.Pipeline, error) {
	hook := new(mergeRequestHook)
	if err := json.NewDecoder(payload).Decode(hook); err != nil {
		return nil, nil, err
	}

	if hook.ObjectAttributes == nil {
		return nil, nil, fmt.Errorf("parsed merge_request webhook does not contain merge request info")
	}

	if !supportedMergeRequestAction(hook.EventType) {
		log.Debug().Msgf("merge_request action '%s' is not supported, ignoring", hook.EventType)
		return nil, nil, nil
	}

	repo := toRepo(hook.Project)
	pipeline := pipelineFromMergeRequest(hook)
	return repo, pipeline, nil
}

func supportedMergeRequestAction(action string) bool {
	switch action {
	case actionOpen, actionUpdate, actionReopen, actionClose, actionMerge:
		return true
	default:
		return false
	}
}

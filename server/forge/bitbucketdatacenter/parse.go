// Copyright 2025 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bitbucketdatacenter

import (
	"fmt"
	"net/http"

	"github.com/neticdk/go-bitbucket/bitbucket"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type HookResult struct {
	Repo     *model.Repo
	Pipeline *model.Pipeline
	Event    any
	Payload  []byte
}

func parseHook(r *http.Request, baseURL string) (*HookResult, string, string, error) {
	ev, payload, err := bitbucket.ParsePayloadWithoutSignature(r)
	if err != nil {
		return nil, "", "", fmt.Errorf("unable to parse payload from webhook invocation: %w", err)
	}

	result := &HookResult{
		Event:   ev,
		Payload: payload,
	}

	switch e := ev.(type) {
	case *bitbucket.RepositoryPushEvent:
		result.Repo = convertRepo(&e.Repository, nil, "")
		result.Pipeline = convertRepositoryPushEvent(e, baseURL)
		currCommit, prevCommit := convertGetCommitRange(e)
		return result, currCommit, prevCommit, nil
	case *bitbucket.PullRequestEvent:
		result.Repo = convertRepo(&e.PullRequest.Target.Repository, nil, "")
		result.Pipeline = convertPullRequestEvent(e, baseURL)
		return result, "", "", nil
	default:
		return nil, "", "", &types.ErrIgnoreEvent{Event: fmt.Sprintf("%T", e), Reason: "unsupported webhook event type"}
	}
}

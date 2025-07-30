// Copyright 2025 Woodpecker Authors
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

package fixtures

import (
	_ "embed"
	"net/http"
	"net/url"
)

var (
	ServiceHookMethod = http.MethodPost
	ServiceHookURL, _ = url.Parse(
		"http://10.40.8.5:8000/hook?owner=test&name=woodpecker&access_token=dummyToken." +
			"eyJ0ZXh0IjoidGVzdC93b29kcGVja2VyIiwidHlwZSI6Imhvb2sifQ.x3kPnmZtxZQ_9_eMhfQ1HSmj_SLhdT_Lu2hMczWjKh0")
	ServiceHookHeaders = http.Header{
		"Content-Type":   []string{"application/json"},
		"User-Agent":     []string{"GitLab/14.3.0"},
		"X-Gitlab-Event": []string{"Service Hook"},
	}
	ReleaseHookHeaders = http.Header{
		"Content-Type":   []string{"application/json"},
		"User-Agent":     []string{"GitLab/14.3.0"},
		"X-Gitlab-Event": []string{"Release Hook"},
	}
	MergeRequestHookHeaders = http.Header{
		"Content-Type":   []string{"application/json"},
		"User-Agent":     []string{"GitLab/18.3.0-pre"},
		"X-Gitlab-Event": []string{"Merge Request Hook"},
	}
)

// HookPush is payload of a push event
//
//go:embed HookPush.json
var HookPush []byte

// HookTag is payload of a TAG event
//
//go:embed HookTag.json
var HookTag []byte

// HookPullRequest is payload of a PULL_REQUEST event
//
//go:embed HookPullRequest.json
var HookPullRequest []byte

//go:embed HookPullRequestWithoutChanges.json
var HookPullRequestWithoutChanges []byte

//go:embed HookPullRequestApproved.json
var HookPullRequestApproved []byte

//go:embed HookPullRequestEdited.json
var HookPullRequestEdited []byte

//go:embed HookPullRequestClosed.json
var HookPullRequestClosed []byte

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged []byte

//go:embed WebhookReleaseBody.json
var WebhookReleaseBody []byte

//go:embed HookPullRequestReviewAck.json
var HookPullRequestReviewAck []byte

/// HookPullRequestReviewAck:
///   Content-Type: application/json
///   User-Agent: GitLab/18.3.0-pre
///   Idempotency-Key: 72e93f1d-b856-4989-b8aa-cf711c58f8fc
///   X-Gitlab-Event: Merge Request Hook
///   X-Gitlab-Webhook-UUID: 416c6b60-1350-4d1c-abc1-fa2b807ef434
///   X-Gitlab-Instance: https://gitlab.com
///   X-Gitlab-Token: [REDACTED]
///   X-Gitlab-Event-UUID: c321aa65-ee6d-4c2e-9570-847bd8f40fe3

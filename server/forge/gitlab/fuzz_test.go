// Copyright 2026 Woodpecker Authors
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

package gitlab

import (
	"net/http"
	"testing"

	gitlab "gitlab.com/gitlab-org/api/client-go/v2"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/gitlab/fixtures"
)

// FuzzParseWebhook feeds untrusted webhook payloads of arbitrary event types
// into the webhook decoding and the pure hook conversion functions. The
// property checked is that neither decoding nor conversion ever panics.
func FuzzParseWebhook(f *testing.F) {
	f.Add("Push Hook", fixtures.HookPush)
	f.Add("Tag Push Hook", fixtures.HookTag)
	f.Add("Merge Request Hook", fixtures.HookPullRequestOpened)
	f.Add("Merge Request Hook", fixtures.HookPullRequestMerged)
	f.Add("Release Hook", fixtures.WebhookReleaseBody)

	f.Fuzz(func(_ *testing.T, eventType string, payload []byte) {
		parsed, err := gitlab.ParseWebhook(gitlab.EventType(eventType), payload)
		if err != nil {
			return
		}
		switch event := parsed.(type) {
		case *gitlab.MergeEvent:
			_, _, _, _, _ = convertMergeRequestHook(event, &http.Request{})
		case *gitlab.PushEvent:
			_, _, _ = convertPushHook(event)
		case *gitlab.TagEvent:
			_, _, _, _ = convertTagHook(event)
		case *gitlab.ReleaseEvent:
			_, _, _ = convertReleaseHook(event)
		}
	})
}

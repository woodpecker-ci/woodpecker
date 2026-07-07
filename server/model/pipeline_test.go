// Copyright 2026 Woodpecker Authors
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

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipelineToAPIModel(t *testing.T) {
	tests := []struct {
		name             string
		pipeline         Pipeline
		wantTitle        string
		wantMessage      string
		wantSender       string
		wantIsPrerelease bool
		wantCommit       string
		wantTimestamp    int64
		wantEmail        string
	}{
		{
			name:        "cron uses cron name as message and sender",
			pipeline:    Pipeline{Event: EventCron, Cron: "nightly"},
			wantMessage: "nightly",
			wantSender:  "nightly",
		},
		{
			name:        "tag uses tag title in message",
			pipeline:    Pipeline{Event: EventTag, TagTitle: "v1.2.3"},
			wantMessage: "created tag v1.2.3",
		},
		{
			name:        "release without release object falls back to tag title",
			pipeline:    Pipeline{Event: EventRelease, TagTitle: "v2.0.0"},
			wantMessage: "created release v2.0.0",
		},
		{
			name: "release with release object uses release title and prerelease flag",
			pipeline: Pipeline{
				Event:    EventRelease,
				TagTitle: "v2.0.0",
				Release:  &Release{Title: "Release 2.0", IsPrerelease: true},
			},
			wantTitle:        "Release 2.0",
			wantMessage:      "created release Release 2.0",
			wantIsPrerelease: true,
		},
		{
			name:     "push leaves derived fields untouched",
			pipeline: Pipeline{Event: EventPush, Commit: &Commit{Message: "fix bug"}},
			// message is the stored commit message, not overwritten
			wantMessage: "fix bug",
		},
		{
			name: "commit substruct is exposed via the deprecated fields",
			pipeline: Pipeline{
				Event: EventPush,
				Commit: &Commit{
					SHA:       "cafe1234",
					Message:   "fix bug",
					Timestamp: 1700000000,
					Author:    CommitAuthor{Name: "alice", Email: "alice@example.com"},
				},
			},
			wantCommit:    "cafe1234",
			wantMessage:   "fix bug",
			wantTimestamp: 1700000000,
			wantEmail:     "alice@example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.pipeline
			ap := p.ToAPIModel()
			assert.Equal(t, tc.wantTitle, ap.Title)
			assert.Equal(t, tc.wantMessage, ap.Message)
			assert.Equal(t, tc.wantSender, ap.Sender)
			assert.Equal(t, tc.wantIsPrerelease, ap.IsPrerelease)
			assert.Equal(t, tc.wantCommit, ap.Commit)
			assert.Equal(t, tc.wantTimestamp, ap.Timestamp)
			assert.Equal(t, tc.wantEmail, ap.Email)
		})
	}
}

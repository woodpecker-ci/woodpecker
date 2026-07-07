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

func TestFeedToAPIModel(t *testing.T) {
	tests := []struct {
		name        string
		feed        Feed
		wantTitle   string
		wantMessage string
	}{
		{
			name:        "tag uses tag title in message",
			feed:        Feed{Event: string(EventTag), TagTitle: "v1.0"},
			wantMessage: "created tag v1.0",
		},
		{
			name:        "release without release object falls back to tag title",
			feed:        Feed{Event: string(EventRelease), TagTitle: "v3.0"},
			wantMessage: "created release v3.0",
		},
		{
			name: "release with release object uses release title",
			feed: Feed{
				Event:    string(EventRelease),
				TagTitle: "v3.0",
				Release:  &Release{Title: "My Release"},
			},
			wantTitle:   "My Release",
			wantMessage: "created release My Release",
		},
		{
			name:        "push leaves derived fields untouched",
			feed:        Feed{Event: string(EventPush), Commit: &Commit{Message: "some commit"}},
			wantTitle:   "some commit",
			wantMessage: "some commit",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := tc.feed
			af := f.ToAPIModel()
			assert.Equal(t, tc.wantTitle, af.Title)
			assert.Equal(t, tc.wantMessage, af.Message)
		})
	}
}

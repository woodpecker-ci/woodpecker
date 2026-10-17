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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPullRequestJSON(t *testing.T) {
	tests := []struct {
		name string
		pr   PullRequest
		want string
	}{
		{
			name: "branches are omitted when unknown",
			pr:   PullRequest{Index: "1", Title: "title"},
			want: `{"index":"1","title":"title"}`,
		},
		{
			name: "branches are included when set",
			pr:   PullRequest{Index: "1", Title: "title", SourceBranch: "feature/x", TargetBranch: "main"},
			want: `{"index":"1","title":"title","source_branch":"feature/x","target_branch":"main"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.pr)
			assert.NoError(t, err)
			assert.JSONEq(t, tt.want, string(got))
		})
	}
}

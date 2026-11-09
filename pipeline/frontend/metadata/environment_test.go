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

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnviron(t *testing.T) {
	m := Metadata{
		Sys: System{Name: "wp"},
		Repo: Repo{
			ID: 42,
		},
		Curr: Pipeline{
			Event: EventRelease,
			Release: Release{
				Title:        "v1.2.3",
				IsPrerelease: true,
			},
			Commit: Commit{
				Ref:       "refs/tags/v1.2.3",
				Timestamp: 1722617519,
			},
		},
		Prev: Pipeline{
			Event: EventPullMetadata,
			Commit: Commit{
				Refspec:   "branch-a:branch-b",
				Timestamp: 1722610173,
			},
		},
	}

	envs := m.Environ()
	assert.Equal(t, "wp", envs["CI"])
	assert.Equal(t, "42", envs["CI_REPO_ID"])
	assert.Equal(t, "release", envs["CI_PIPELINE_EVENT"])
	assert.Equal(t, "pull_request_metadata", envs["CI_PREV_PIPELINE_EVENT"])
	assert.Equal(t, "v1.2.3", envs["CI_PIPELINE_RELEASE_TITLE"])
	assert.Equal(t, "true", envs["CI_PIPELINE_RELEASE_PRE"])
	assert.Equal(t, "true", envs["CI_COMMIT_PRERELEASE"]) // deprecated alias
	assert.Equal(t, "branch-a", envs["CI_PREV_COMMIT_SOURCE_BRANCH"])
	assert.Equal(t, "branch-b", envs["CI_PREV_COMMIT_TARGET_BRANCH"])
	assert.Equal(t, "[]", envs["CI_PIPELINE_FILES"])
	assert.Equal(t, "v1.2.3", envs["CI_COMMIT_TAG"])
	assert.Equal(t, "1722617519", envs["CI_COMMIT_TIMESTAMP"])
	assert.Equal(t, "1722610173", envs["CI_PREV_COMMIT_TIMESTAMP"])

	m = Metadata{
		Sys: System{Name: "wp"},
		Curr: Pipeline{
			Event: EventPull,
			Commit: Commit{
				ChangedFiles:     []string{"readme", "license"},
				Refspec:          "branch-a:branch-b",
				PullRequestDraft: true,
			},
		},
		Prev: Pipeline{
			Event: EventPull,
			Commit: Commit{
				Refspec: "branch-a:branch-b",
			},
		},
	}

	envs = m.Environ()

	_, ok := envs["CI_COMMIT_TAG"]
	assert.False(t, ok)

	assert.Equal(t, `["readme","license"]`, envs["CI_PIPELINE_FILES"])
	assert.Equal(t, "true", envs["CI_COMMIT_PULL_REQUEST_DRAFT"])

	m = Metadata{
		Sys: System{Name: "wp"},
		Curr: Pipeline{
			Event: EventPull,
			Commit: Commit{
				Refspec: "branch-a:branch-b",
			},
		},
	}

	envs = m.Environ()
	assert.Equal(t, "false", envs["CI_COMMIT_PULL_REQUEST_DRAFT"])
}

func TestEnvironForgeHost(t *testing.T) {
	tests := []struct {
		name  string
		forge Forge
		want  string
	}{
		{name: "host of the forge", forge: Forge{Type: "gitea", URL: "https://git.example.com", Host: "git.example.com"}, want: "git.example.com"},
		{name: "port is kept", forge: Forge{Type: "forgejo", URL: "http://forgejo.internal:3000", Host: "forgejo.internal:3000"}, want: "forgejo.internal:3000"},
		{name: "no forge configured", forge: Forge{}, want: ""},
		{name: "an url on its own adds no host", forge: Forge{Type: "gitea", URL: "https://git.example.com"}, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := Metadata{Forge: tt.forge}
			envs := m.Environ()
			assert.Equal(t, tt.want, envs["CI_FORGE_HOST"])
			_, ok := envs["CI_FORGE_HOST"]
			assert.Equal(t, tt.want != "", ok)
		})
	}
}

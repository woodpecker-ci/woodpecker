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
		Curr: Pipeline{
			Event: EventRelease,
			Commit: Commit{
				Ref:          "refs/tags/v1.2.3",
				IsPrerelease: true,
			},
		},
		Prev: Pipeline{
			Event: EventPullMetadata,
			Commit: Commit{
				Refspec: "branch-a:branch-b",
			},
		},
	}

	envs := m.Environ()
	assert.Equal(t, "wp", envs["CI"])
	assert.Equal(t, "release", envs["CI_PIPELINE_EVENT"])
	assert.Equal(t, "pull_request_metadata", envs["CI_PREV_PIPELINE_EVENT"])
	assert.Equal(t, "true", envs["CI_COMMIT_PRERELEASE"])
	assert.Equal(t, "branch-a", envs["CI_PREV_COMMIT_SOURCE_BRANCH"])
	assert.Equal(t, "branch-b", envs["CI_PREV_COMMIT_TARGET_BRANCH"])
	assert.Equal(t, "[]", envs["CI_PIPELINE_FILES"])
	assert.Equal(t, "v1.2.3", envs["CI_COMMIT_TAG"])

	m = Metadata{
		Sys: System{Name: "wp"},
		Curr: Pipeline{
			Event: EventPull,
			Commit: Commit{
				ChangedFiles: []string{"readme", "license"},
				Refspec:      "branch-a:branch-b",
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
}

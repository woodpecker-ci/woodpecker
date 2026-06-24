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

package metadata

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestMetadataPipelineFromModelPipelineRelease(t *testing.T) {
	tests := []struct {
		name             string
		pipeline         *model.Pipeline
		wantRelease      string
		wantIsPrerelease bool
	}{
		{
			name:     "no release leaves release fields empty and does not panic",
			pipeline: &model.Pipeline{Number: 1, Event: model.EventPush},
		},
		{
			name: "release populates release title and prerelease flag",
			pipeline: &model.Pipeline{
				Number:  2,
				Event:   model.EventRelease,
				Release: &model.Release{Title: "v1.0.0", IsPrerelease: true},
			},
			wantRelease:      "v1.0.0",
			wantIsPrerelease: true,
		},
		{
			name: "stable release keeps prerelease false",
			pipeline: &model.Pipeline{
				Number:  3,
				Event:   model.EventRelease,
				Release: &model.Release{Title: "v2.0.0"},
			},
			wantRelease: "v2.0.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := metadataPipelineFromModelPipeline(tc.pipeline, false)
			assert.Equal(t, tc.wantRelease, result.Release)
			assert.Equal(t, tc.wantIsPrerelease, result.Commit.IsPrerelease)
		})
	}
}

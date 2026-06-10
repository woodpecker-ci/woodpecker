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

package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func pipelineWithSteps(status model.StatusValue, retryCount int64, steps ...*model.Step) *model.Pipeline {
	return &model.Pipeline{
		Status:          status,
		InfraRetryCount: retryCount,
		Workflows:       []*model.Workflow{{Children: steps}},
	}
}

func step(state model.StatusValue, infra bool, failure string) *model.Step {
	return &model.Step{State: state, InfraFailure: infra, Failure: failure}
}

func TestShouldRetryInfraFailure(t *testing.T) {
	tests := []struct {
		name        string
		pipeline    *model.Pipeline
		maxAttempts int64
		want        bool
	}{
		{
			name:        "infra-only failure is retried",
			pipeline:    pipelineWithSteps(model.StatusFailure, 0, step(model.StatusFailure, true, model.FailureFail)),
			maxAttempts: 2,
			want:        true,
		},
		{
			name: "infra failure plus fail-fast killed siblings is retried",
			pipeline: pipelineWithSteps(model.StatusFailure, 0,
				step(model.StatusFailure, true, model.FailureFail),
				step(model.StatusKilled, false, model.FailureFail),
				step(model.StatusSuccess, false, model.FailureFail),
			),
			maxAttempts: 2,
			want:        true,
		},
		{
			name: "genuine failure alongside infra failure is not retried",
			pipeline: pipelineWithSteps(model.StatusFailure, 0,
				step(model.StatusFailure, true, model.FailureFail),
				step(model.StatusFailure, false, model.FailureFail),
			),
			maxAttempts: 2,
			want:        false,
		},
		{
			name:        "non-infra failure is not retried",
			pipeline:    pipelineWithSteps(model.StatusFailure, 0, step(model.StatusFailure, false, model.FailureFail)),
			maxAttempts: 2,
			want:        false,
		},
		{
			name:        "feature disabled (maxAttempts 0) never retries",
			pipeline:    pipelineWithSteps(model.StatusFailure, 0, step(model.StatusFailure, true, model.FailureFail)),
			maxAttempts: 0,
			want:        false,
		},
		{
			name:        "attempt budget exhausted is not retried",
			pipeline:    pipelineWithSteps(model.StatusFailure, 2, step(model.StatusFailure, true, model.FailureFail)),
			maxAttempts: 2,
			want:        false,
		},
		{
			name:        "successful pipeline is not retried",
			pipeline:    pipelineWithSteps(model.StatusSuccess, 0, step(model.StatusSuccess, false, model.FailureFail)),
			maxAttempts: 2,
			want:        false,
		},
		{
			name:        "killed (canceled) pipeline is not retried",
			pipeline:    pipelineWithSteps(model.StatusKilled, 0, step(model.StatusKilled, false, model.FailureFail)),
			maxAttempts: 2,
			want:        false,
		},
		{
			name:        "error (config) pipeline is not retried",
			pipeline:    pipelineWithSteps(model.StatusError, 0, step(model.StatusError, false, model.FailureFail)),
			maxAttempts: 2,
			want:        false,
		},
		{
			name: "ignored failing step does not block an infra retry",
			pipeline: pipelineWithSteps(model.StatusFailure, 0,
				step(model.StatusFailure, true, model.FailureFail),
				step(model.StatusFailure, false, model.FailureIgnore),
			),
			maxAttempts: 2,
			want:        true,
		},
		{
			name: "infra-flagged step reported killed still retries",
			pipeline: pipelineWithSteps(model.StatusFailure, 0,
				step(model.StatusKilled, true, model.FailureFail),
			),
			maxAttempts: 2,
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, shouldRetryInfraFailure(tt.pipeline, tt.maxAttempts))
		})
	}
}

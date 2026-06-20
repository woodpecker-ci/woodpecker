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

package gitlab

import (
	"testing"

	"github.com/stretchr/testify/assert"
	gitlab "gitlab.com/gitlab-org/api/client-go/v2"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func TestGetStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		status model.StatusValue
		want   gitlab.BuildStateValue
	}{
		{model.StatusPending, gitlab.Pending},
		{model.StatusBlocked, gitlab.Pending},
		{model.StatusRunning, gitlab.Running},
		{model.StatusSuccess, gitlab.Success},
		{model.StatusFailure, gitlab.Failed},
		{model.StatusError, gitlab.Failed},
		{model.StatusKilled, gitlab.Canceled},
		// unknown statuses fall back to failed
		{model.StatusDeclined, gitlab.Failed},
	}

	for _, tt := range tests {
		assert.Equalf(t, tt.want, getStatus(tt.status), "status %q", tt.status)
	}
}

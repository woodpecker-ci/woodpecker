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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func blockedPipeline(env *statusEnv) *model.Pipeline {
	p := pushPipeline()
	p.ID, p.Number, p.RepoID, p.Status = 1, 1, env.repo.ID, model.StatusBlocked
	env.pipeline = *p
	return p
}

// Approving posts the statuses once, from start().
func TestApprovePostsStatusesOnce(t *testing.T) {
	env := newStatusEnv(t, nil)

	approved, err := Approve(t.Context(), env.store, blockedPipeline(env), env.user, env.repo)
	require.NoError(t, err)
	require.Len(t, approved.Workflows, 2)

	assert.ElementsMatch(t, env.round(model.StatusPending, model.StatusPending), env.postedStatuses())
}

func TestApproveStartFailureErrorsWorkflowsAndSteps(t *testing.T) {
	env := newStatusEnv(t, errors.New("queue unavailable"))

	_, err := Approve(t.Context(), env.store, blockedPipeline(env), env.user, env.repo)
	require.Error(t, err)

	env.assertErrored(t)
}

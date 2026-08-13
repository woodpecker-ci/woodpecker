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
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	store_mocks "go.woodpecker-ci.org/woodpecker/v3/server/store/mocks"
)

func compileWorkflow(name string, state model.StatusValue, result *model.CompileResult) *model.Workflow {
	return &model.Workflow{
		Name:          name,
		State:         state,
		Phase:         model.WorkflowPhaseCompile,
		CompileResult: result,
	}
}

func emitted(configs ...model.CompileConfig) *model.CompileResult {
	return &model.CompileResult{Configs: configs}
}

func TestCollectCompileResults(t *testing.T) {
	t.Parallel()

	t.Run("gathers in workflow order", func(t *testing.T) {
		configs, err := collectCompileResults([]*model.Workflow{
			compileWorkflow("a", model.StatusSuccess, emitted(model.CompileConfig{Name: "first", Data: []byte("x")})),
			compileWorkflow("b", model.StatusSuccess, emitted(model.CompileConfig{Name: "second", Data: []byte("y")})),
		})
		require.NoError(t, err)

		require.Len(t, configs, 2)
		assert.Equal(t, "first", configs[0].Name)
		assert.Equal(t, "second", configs[1].Name)
	})

	t.Run("ignores run workflows", func(t *testing.T) {
		configs, err := collectCompileResults([]*model.Workflow{
			compileWorkflow("a", model.StatusSuccess, emitted(model.CompileConfig{Name: "first", Data: []byte("x")})),
			{Name: "b", State: model.StatusSuccess, Phase: model.WorkflowPhaseRun},
		})
		require.NoError(t, err)
		assert.Len(t, configs, 1)
	})

	t.Run("an empty config list means proceed unchanged", func(t *testing.T) {
		configs, err := collectCompileResults([]*model.Workflow{
			compileWorkflow("a", model.StatusSuccess, new(model.CompileResult)),
		})
		require.NoError(t, err)
		assert.Empty(t, configs)
	})

	t.Run("a failed compile workflow fails the phase", func(t *testing.T) {
		_, err := collectCompileResults([]*model.Workflow{
			compileWorkflow("a", model.StatusFailure, emitted()),
		})
		assert.ErrorContains(t, err, `compile workflow "a" failed`)
	})

	t.Run("no result at all fails the phase", func(t *testing.T) {
		// distinct from an empty config list: a generator that crashed before
		// printing anything must not read as one that decided to do nothing
		_, err := collectCompileResults([]*model.Workflow{
			compileWorkflow("a", model.StatusSuccess, nil),
		})
		assert.ErrorContains(t, err, "reported no config response")
	})

	t.Run("an unreadable response fails the phase", func(t *testing.T) {
		_, err := collectCompileResults([]*model.Workflow{
			compileWorkflow("a", model.StatusSuccess, &model.CompileResult{Error: "malformed block"}),
		})
		assert.ErrorContains(t, err, "malformed block")
	})
}

func TestShiftPIDsPast(t *testing.T) {
	t.Parallel()

	// Run workflows are appended to a pipeline that already has compile
	// workflows and their steps, so they have to continue the numbering rather
	// than collide with it.
	existing := []*model.Workflow{
		{PID: 1, Children: []*model.Step{{PID: 2}, {PID: 3}}},
	}
	items := []*builder.Item{
		{Workflow: &builder.Workflow{PID: 1}},
		{Workflow: &builder.Workflow{PID: 2}},
	}

	shiftPIDsPast(items, existing)

	assert.Equal(t, 4, items[0].Workflow.PID)
	assert.Equal(t, 5, items[1].Workflow.PID)
}

func TestCompilePhaseDoneWaitsForEveryCompileWorkflow(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)
	currentPipeline := &model.Pipeline{ID: 1, CompileState: model.CompileStateCompiling}

	mockStore.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
		compileWorkflow("a", model.StatusSuccess, emitted()),
		compileWorkflow("b", model.StatusRunning, nil),
	}, nil)

	// No compare-and-swap is expected: the mock fails the test if one happens,
	// which is the point. Claiming the merge while a workflow is still
	// producing configs would merge half a result.
	err := CompilePhaseDone(t.Context(), nil, mockStore, currentPipeline, &model.Repo{}, &model.User{})
	assert.ErrorIs(t, err, ErrCompilePhaseNotDone)
}

func TestCompilePhaseDoneRunsTheMergeExactlyOnce(t *testing.T) {
	t.Parallel()

	mockStore := store_mocks.NewMockStore(t)
	currentPipeline := &model.Pipeline{ID: 1, CompileState: model.CompileStateCompiling}

	mockStore.On("WorkflowGetTree", currentPipeline).Return([]*model.Workflow{
		compileWorkflow("a", model.StatusSuccess, emitted()),
	}, nil)

	// Every compile workflow has finished, so this caller gets as far as the
	// claim, but another Done call already took it.
	mockStore.On("PipelineCompileStateCompareAndSwap",
		int64(1), model.CompileStateCompiling, model.CompileStateMerging,
	).Return(false, nil)

	err := CompilePhaseDone(t.Context(), nil, mockStore, currentPipeline, &model.Repo{}, &model.User{})
	assert.ErrorIs(t, err, ErrCompilePhaseNotDone,
		"only the caller that moved the state may merge")
}

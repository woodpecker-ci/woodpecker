// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// TestCheckParentState covers checkParentState for both pipeline-level
// (isStep=false) and workflow-level (isStep=true) checks in a single
// table-driven test, eliminating duplicate cases shared by the two levels.
func TestCheckParentState(t *testing.T) {
	t.Parallel()

	// States that always allow a child to proceed, regardless of level.
	allowedParents := []model.StatusValue{
		model.StatusCreated,
		model.StatusPending,
		model.StatusRunning,
	}

	// Terminal parent states that reject a running child as an illegal re-run.
	terminalParents := []model.StatusValue{
		model.StatusSuccess,
		model.StatusFailure,
		model.StatusKilled,
		model.StatusError,
		model.StatusSkipped,
	}

	// Parents whose terminal children are exempt (allowed through).
	exemptParents := []model.StatusValue{
		model.StatusCanceled,
		model.StatusFailure,
		model.StatusKilled,
	}

	// Child states considered exempt under a terminal/canceled parent.
	exemptChildren := []model.StatusValue{
		model.StatusCanceled,
		model.StatusKilled,
		model.StatusSkipped,
	}

	// Error sentinels per level.
	type levelConfig struct {
		isStep       bool
		blockedErr   error
		reRunErr     error
		extraExempt  []model.StatusValue // additional exempt child states beyond the shared set
		extraRejects []model.StatusValue // additional terminal parents beyond the shared set
	}

	levels := []levelConfig{
		{
			isStep:       false,
			blockedErr:   ErrAgentIllegalPipelineWorkflowRun,
			reRunErr:     ErrAgentIllegalPipelineWorkflowReRunStateChange,
			extraExempt:  []model.StatusValue{model.StatusFailure, model.StatusSuccess},
			extraRejects: []model.StatusValue{model.StatusDeclined},
		},
		{
			isStep:      true,
			blockedErr:  ErrAgentIllegalWorkflowRun,
			reRunErr:    ErrAgentIllegalWorkflowReRunStateChange,
			extraExempt: []model.StatusValue{model.StatusFailure, model.StatusSuccess},
		},
	}

	for _, lc := range levels {
		label := "pipeline"
		if lc.isStep {
			label = "step"
		}

		t.Run(label, func(t *testing.T) {
			t.Parallel()

			// Allowed parent states.
			for _, ps := range allowedParents {
				t.Run(fmt.Sprintf("%s allows", ps), func(t *testing.T) {
					t.Parallel()
					assert.NoError(t, checkParentState(ps, model.StatusRunning, lc.isStep))
				})
			}

			// Blocked parent.
			t.Run("blocked rejects", func(t *testing.T) {
				t.Parallel()
				assert.ErrorIs(t, checkParentState(model.StatusBlocked, model.StatusRunning, lc.isStep), lc.blockedErr)
			})

			// Terminal parents reject a running child.
			allTerminal := append(terminalParents, lc.extraRejects...)
			for _, ps := range allTerminal {
				t.Run(fmt.Sprintf("%s running child rejected", ps), func(t *testing.T) {
					t.Parallel()
					assert.ErrorIs(t, checkParentState(ps, model.StatusRunning, lc.isStep), lc.reRunErr)
				})
			}

			// Canceled parent with running child is also rejected.
			t.Run("canceled running child rejected", func(t *testing.T) {
				t.Parallel()
				assert.ErrorIs(t, checkParentState(model.StatusCanceled, model.StatusRunning, lc.isStep), lc.reRunErr)
			})

			// Exempt parent + exempt child combinations → allowed.
			allExemptChildren := append(exemptChildren, lc.extraExempt...)
			for _, ps := range exemptParents {
				for _, cs := range allExemptChildren {
					t.Run(fmt.Sprintf("%s parent %s child allowed", ps, cs), func(t *testing.T) {
						t.Parallel()
						assert.NoError(t, checkParentState(ps, cs, lc.isStep))
					})
				}
			}
		})
	}
}

func TestAllowAppendingLogs(t *testing.T) {
	t.Parallel()

	recentFinish := time.Now().Add(-30 * time.Second).Unix()
	staleFinish := time.Now().Add(-10 * time.Minute).Unix()

	// Step running always allows logs, regardless of pipeline state or age.
	t.Run("step running always allowed", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			name   string
			status model.StatusValue
			finish int64
		}{
			{"pipeline running", model.StatusRunning, 0},
			{"pipeline success stale", model.StatusSuccess, staleFinish},
			{"pipeline failure stale", model.StatusFailure, staleFinish},
			{"pipeline killed stale", model.StatusKilled, staleFinish},
		} {
			t.Run(tc.name, func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: tc.status, Finished: tc.finish}
				assert.NoError(t, allowAppendingLogs(p, &model.Step{State: model.StatusRunning}))
			})
		}
	})

	// Pipeline running allows logs for any step state.
	t.Run("pipeline running any step allowed", func(t *testing.T) {
		t.Parallel()

		for _, ss := range []model.StatusValue{
			model.StatusSuccess, model.StatusFailure, model.StatusPending, model.StatusKilled,
		} {
			t.Run(string(ss), func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: model.StatusRunning}
				assert.NoError(t, allowAppendingLogs(p, &model.Step{State: ss}))
			})
		}
	})

	// Recent finish → drain window allows logs.
	t.Run("recent finish drain allowed", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			pStatus model.StatusValue
			sState  model.StatusValue
		}{
			{model.StatusSuccess, model.StatusSuccess},
			{model.StatusFailure, model.StatusFailure},
			{model.StatusKilled, model.StatusPending},
		} {
			t.Run(fmt.Sprintf("%s/%s", tc.pStatus, tc.sState), func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: tc.pStatus, Finished: recentFinish}
				assert.NoError(t, allowAppendingLogs(p, &model.Step{State: tc.sState}))
			})
		}
	})

	// Stale finish → drain window expired → reject.
	t.Run("stale finish drain rejected", func(t *testing.T) {
		t.Parallel()

		for _, tc := range []struct {
			pStatus model.StatusValue
			sState  model.StatusValue
			finish  int64
		}{
			{model.StatusSuccess, model.StatusSuccess, staleFinish},
			{model.StatusFailure, model.StatusFailure, staleFinish},
			{model.StatusKilled, model.StatusPending, staleFinish},
			{model.StatusError, model.StatusCreated, staleFinish},
			{model.StatusSuccess, model.StatusSuccess, 0}, // zero = never recorded
		} {
			t.Run(fmt.Sprintf("%s/%s/fin=%d", tc.pStatus, tc.sState, tc.finish), func(t *testing.T) {
				t.Parallel()
				p := &model.Pipeline{Status: tc.pStatus, Finished: tc.finish}
				assert.ErrorIs(t, allowAppendingLogs(p, &model.Step{State: tc.sState}), ErrAgentIllegalLogStreaming)
			})
		}
	})
}

// TestAllowAppendingLogsDrainBoundary guards the exact edge of the 5-minute
// drain window against off-by-one errors.
func TestAllowAppendingLogsDrainBoundary(t *testing.T) {
	t.Parallel()

	step := &model.Step{State: model.StatusSuccess}

	t.Run("just inside drain window allowed", func(t *testing.T) {
		t.Parallel()
		p := &model.Pipeline{
			Status:   model.StatusSuccess,
			Finished: time.Now().Add(-(logStreamDelayAllowed - time.Second)).Unix(),
		}
		assert.NoError(t, allowAppendingLogs(p, step))
	})

	t.Run("just outside drain window rejected", func(t *testing.T) {
		t.Parallel()
		p := &model.Pipeline{
			Status:   model.StatusSuccess,
			Finished: time.Now().Add(-(logStreamDelayAllowed + time.Second)).Unix(),
		}
		assert.ErrorIs(t, allowAppendingLogs(p, step), ErrAgentIllegalLogStreaming)
	})
}

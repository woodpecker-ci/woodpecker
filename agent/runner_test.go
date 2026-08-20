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

//go:build test

package agent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types/mocks"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	rpc_mocks "go.woodpecker-ci.org/woodpecker/v3/rpc/mocks"
)

func dummyWorkflow() *rpc.Workflow {
	return &rpc.Workflow{
		ID: "1",
		Config: &backend_types.Config{
			Stages: []*backend_types.Stage{{
				Steps: []*backend_types.Step{{
					Name:      "build",
					UUID:      "step-uuid",
					OnSuccess: true,
					Environment: map[string]string{
						"CI_REPO":            "octocat/hello-world",
						"CI_PIPELINE_NUMBER": "1",
					},
				}},
			}},
		},
	}
}

// A workflow the server canceled did not fail, it was stopped on purpose. The
// cancellation is reported through WorkflowState.Canceled, so no error text
// must be reported next to it: the server stores that text as the workflow
// error and the web UI shows it as a runtime error.
func TestRunReportsCancelWithoutError(t *testing.T) {
	engine := mocks.NewMockBackend(t)
	engine.On("SetupWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	engine.On("DestroyWorkflow", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// The step start is aborted mid-flight once the workflow is canceled, the
	// backend surfacing whatever its transport made of the cancel cause.
	engine.On("StartStep", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			ctx, ok := args.Get(0).(context.Context)
			if !ok {
				t.Error("StartStep called without a context")
				return
			}
			<-ctx.Done()
		}).
		Return(errors.New(`error during connect: Post "http://.../containers/x/start": Canceled`))

	var done rpc.WorkflowState
	peer := rpc_mocks.NewMockPeer(t)
	peer.On("Next", mock.Anything, mock.Anything).Return(dummyWorkflow(), nil)
	peer.On("Init", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	peer.On("Wait", mock.Anything, mock.Anything).Return(true, nil)
	peer.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	peer.On("Done", mock.Anything, mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			state, ok := args.Get(2).(rpc.WorkflowState)
			if !ok {
				t.Error("Done called without a workflow state")
				return
			}
			done = state
		}).
		Return(nil)

	counter := &State{Metadata: map[string]Info{}}
	runner := NewRunner(peer, rpc.Filter{}, "test-agent", counter, engine)

	assert.NoError(t, runner.Run(t.Context()))

	assert.True(t, done.Canceled, "the workflow must be reported as canceled")
	assert.Empty(t, done.Error, "a cancellation is not a workflow error")
}

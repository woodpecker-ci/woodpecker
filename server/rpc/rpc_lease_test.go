//go:build test

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
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server/logging"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
	"go.woodpecker-ci.org/woodpecker/v3/server/queue"
	"go.woodpecker-ci.org/woodpecker/v3/server/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/server/store/datastore"
	"go.woodpecker-ci.org/woodpecker/v3/shared/constant"
)

// A slow agent whose lease expired before it got past Next must not run the
// workflow next to the agent that picked it up again. Regression for the
// duplicate run reported on #7152: with the real store, RPC and persistent
// queue, agent A's Wait and Init are rejected once the task is pending again,
// agent B runs it alone, and A cannot finish it.
func TestRPCStaleAgentLosesLeaseAfterExpiry(t *testing.T) {
	orig := constant.TaskTimeout
	constant.TaskTimeout = 300 * time.Millisecond
	t.Cleanup(func() { constant.TaskTimeout = orig })

	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()

	s, err := datastore.NewEngine(&store.Opts{Driver: "sqlite3", Config: filepath.Join(t.TempDir(), "lease.sqlite"), XORM: store.XORM{MaxOpenConns: 1, MaxIdleConns: 1}})
	require.NoError(t, err)
	t.Cleanup(func() { _ = s.Close() })
	require.NoError(t, s.Migrate(ctx, true))

	for _, a := range []*model.Agent{{Name: "A", OrgID: model.IDNotSet, Token: "a"}, {Name: "B", OrgID: model.IDNotSet, Token: "b"}} {
		require.NoError(t, s.AgentCreate(a))
	}
	repo := &model.Repo{UserID: 4242, ForgeRemoteID: "1", FullName: "o/r", Owner: "o", Name: "r", OrgID: 1}
	require.NoError(t, s.CreateRepo(repo))
	pl := &model.Pipeline{RepoID: repo.ID, Status: model.StatusPending, Number: 1}
	require.NoError(t, s.CreatePipeline(pl))
	wf := &model.Workflow{PipelineID: pl.ID, PID: 1, State: model.StatusPending, Name: "wf", Children: []*model.Step{
		{UUID: "step-1", PipelineID: pl.ID, PID: 2, PPID: 1, Name: "build", State: model.StatusPending},
	}}
	require.NoError(t, s.WorkflowsCreate([]*model.Workflow{wf}))

	q, err := queue.New(ctx, queue.Config{Backend: queue.TypeMemory, Store: s})
	require.NoError(t, err)
	r := RPC{
		store:         s,
		scheduler:     scheduler.NewScheduler(ctx, s, q, memory.New()),
		logger:        logging.New(),
		pipelineTime:  prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "lease_t"}, []string{"repo", "branch", "status", "pipeline"}),
		pipelineCount: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "lease_c"}, []string{"repo", "branch", "status", "pipeline"}),
	}

	wfID := "1"
	data, err := json.Marshal(rpc.Workflow{ID: wfID})
	require.NoError(t, err)
	require.NoError(t, q.PushAtOnce(ctx, []*model.Task{{ID: wfID, Data: data, Labels: map[string]string{"platform": "x"}}}))

	ctxA := context.WithValue(ctx, agentIDKey, int64(1))
	ctxB := context.WithValue(ctx, agentIDKey, int64(2))
	filter := rpc.Filter{Labels: map[string]string{"platform": "x"}}

	got, err := r.Next(ctxA, filter)
	require.NoError(t, err)
	require.NotNil(t, got)

	// A's Next took longer than the lease: the task is pending again before A
	// continues with its run.
	require.Eventually(t, func() bool { return q.Info(ctx).Stats.Pending == 1 }, 5*time.Second, 10*time.Millisecond)

	waitCtx, waitCancel := context.WithTimeout(ctxA, 2*time.Second)
	canceled, err := r.Wait(waitCtx, wfID)
	waitCancel()
	assert.False(t, canceled)
	require.ErrorIs(t, err, ErrAgentLostLease)
	require.ErrorIs(t, err, queue.ErrTaskExpired)

	err = r.Init(ctxA, wfID, rpc.WorkflowState{Started: time.Now().Unix()})
	require.ErrorIs(t, err, ErrAgentLostLease)

	// The store still locks the workflow to A, only the queue knows the lease
	// is gone: A can neither report a step nor finish the workflow.
	err = r.Update(ctxA, wfID, rpc.StepState{StepUUID: "step-1", Started: time.Now().Unix()})
	require.ErrorIs(t, err, ErrAgentLostLease)
	err = r.Done(ctxA, wfID, rpc.WorkflowState{Started: 1, Finished: time.Now().Unix()})
	require.ErrorIs(t, err, ErrAgentLostLease)

	nextCtx, nextCancel := context.WithTimeout(ctxB, 5*time.Second)
	gotB, err := r.Next(nextCtx, filter)
	nextCancel()
	require.NoError(t, err)
	require.NotNil(t, gotB)
	assert.Equal(t, wfID, gotB.ID)
	require.NoError(t, r.Init(ctxB, wfID, rpc.WorkflowState{Started: time.Now().Unix()}))

	// B holds the lease and the workflow lock now: A can neither wait on nor
	// finish the workflow. The store check fires first, the queue check backs it.
	waitCtx, waitCancel = context.WithTimeout(ctxA, 2*time.Second)
	_, err = r.Wait(waitCtx, wfID)
	waitCancel()
	require.ErrorIs(t, err, ErrAgentIllegalWorkflowAgentID)
	require.ErrorIs(t, r.scheduler.Leased(ctxA, 1, wfID), queue.ErrAgentMissMatch)
	err = r.Done(ctxA, wfID, rpc.WorkflowState{Started: 1, Finished: time.Now().Unix()})
	require.ErrorIs(t, err, ErrAgentIllegalWorkflowAgentID)

	require.NoError(t, r.Done(ctxB, wfID, rpc.WorkflowState{Started: 1, Finished: time.Now().Unix()}))
	w, err := s.WorkflowLoad(1)
	require.NoError(t, err)
	assert.Equal(t, model.StatusSuccess, w.State)
	assert.Equal(t, int64(2), w.AgentID)
	assert.Equal(t, 0, q.Info(ctx).Stats.Running)
	assert.Equal(t, 0, q.Info(ctx).Stats.Pending)
}

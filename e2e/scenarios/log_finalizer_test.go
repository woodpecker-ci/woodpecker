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

//go:build test

package scenarios

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/api"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
	service_log "go.woodpecker-ci.org/woodpecker/v3/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v3/server/services/log/finalizer"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// Black box tests of the log finalizer: a real agent runs a pipeline and streams its log lines
// over gRPC, then the finalizer runs with a clock past the late-log window.

var logFinalizerYAML = []byte(`
steps:
  - name: build
    image: dummy
    commands:
      - echo building
  - name: test
    image: dummy
    commands:
      - echo testing
  - name: on-failure
    image: dummy
    commands:
      - echo never runs
    when:
      status: failure
`)

var logFinalizerCancelYAML = []byte(`
steps:
  - name: long-running
    image: dummy
    commands:
      - echo starting long job
    environment:
      SLEEP: "30s"
  - name: after-cancel
    image: dummy
    commands:
      - echo never runs
`)

// recordingLogStore keeps logs in the database store and records the log StepFinished saw.
type recordingLogStore struct {
	service_log.Service

	mu       sync.Mutex
	finished map[string][]string // step name -> log data at each StepFinished call
}

func (r *recordingLogStore) StepFinished(ctx context.Context, step *model.Step) error {
	data, err := logData(ctx, r.Service, step)
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finished[step.Name] = append(r.finished[step.Name], data)
	return r.Service.StepFinished(ctx, step)
}

func (r *recordingLogStore) calls() map[string][]string {
	r.mu.Lock()
	defer r.mu.Unlock()
	calls := make(map[string][]string, len(r.finished))
	for name, data := range r.finished {
		calls[name] = append([]string(nil), data...)
	}
	return calls
}

// assertFinalizedOnce asserts every step of a pipeline is finalized and was handed to StepFinished
// exactly once with its complete log, unless it never started and so has no log.
func assertFinalizedOnce(t *testing.T, env *logFinalizerEnv, pipelineID int64) {
	t.Helper()
	calls := env.logs.calls()
	for name, step := range steps(t, env.Store, pipelineID) {
		assert.Equal(t, model.LogStateFinalized, step.LogState, "step %s", name)
		if step.Started == 0 {
			assert.NotContains(t, calls, name, "step %s never started", name)
			continue
		}
		data, err := logData(t.Context(), env.logs, step)
		require.NoError(t, err)
		assert.Equal(t, []string{data}, calls[name], "step %s", name)
	}
}

func logData(ctx context.Context, logs service_log.Service, step *model.Step) (string, error) {
	entries, err := logs.LogFind(ctx, step)
	if err != nil {
		return "", err
	}
	data := ""
	for _, entry := range entries {
		data += string(entry.Data)
	}
	return data, nil
}

type logFinalizerEnv struct {
	*setup.ServerEnv
	agent *setup.AgentEnv
	logs  *recordingLogStore
}

func startLogFinalizerEnv(t *testing.T, config []byte) *logFinalizerEnv {
	t.Helper()
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{{Name: ".woodpecker.yaml", Data: config}})
	logs := &recordingLogStore{Service: server.Config.Services.LogStore, finished: make(map[string][]string)}
	server.Config.Services.LogStore = logs
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)
	return &logFinalizerEnv{ServerEnv: env, agent: agent, logs: logs}
}

func (env *logFinalizerEnv) runPipeline(t *testing.T) *model.Pipeline {
	t.Helper()
	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err)
	return setup.WaitForPipeline(t, env.Store, created.ID)
}

func (env *logFinalizerEnv) finalizer() *finalizer.Finalizer {
	return &finalizer.Finalizer{Store: env.Store, Logs: env.logs, Interval: time.Minute, Batch: 10}
}

// steps returns the steps of a pipeline by name, as stored now.
func steps(t *testing.T, s store.Store, pipelineID int64) map[string]*model.Step {
	t.Helper()
	list, err := s.StepList(pipelineID)
	require.NoError(t, err)
	byName := make(map[string]*model.Step, len(list))
	for _, step := range list {
		byName[step.Name] = step
	}
	return byName
}

// waitLogsSettled waits until the agent sent every log line: the agent batches lines and may send
// them after the pipeline finished, which the late-log window allows for.
func waitLogsSettled(t *testing.T, env *logFinalizerEnv, pipelineID int64, names ...string) {
	t.Helper()
	last := ""
	require.Eventually(t, func() bool {
		all := ""
		for _, name := range names {
			data, err := logData(t.Context(), env.logs, steps(t, env.Store, pipelineID)[name])
			require.NoError(t, err)
			if data == "" {
				return false
			}
			all += name + ":" + data
		}
		settled := all == last
		last = all
		return settled
	}, 20*time.Second, 1500*time.Millisecond, "log lines of %v did not settle", names)
}

// afterWindow is a time when every log of a pipeline that finished now is complete.
func afterWindow() time.Time {
	return time.Now().Add(service_log.StreamDelayAllowed + 2*time.Minute)
}

func TestLogFinalizerFinalizesCompleteLogs(t *testing.T) {
	env := startLogFinalizerEnv(t, logFinalizerYAML)
	finished := env.runPipeline(t)
	require.Equal(t, model.StatusSuccess, finished.Status)
	waitLogsSettled(t, env, finished.ID, "build", "test")
	fin := env.finalizer()

	t.Run("nothing is finalized inside the late-log window", func(t *testing.T) {
		res, err := fin.Tick(t.Context(), time.Now())
		require.NoError(t, err)
		assert.Equal(t, finalizer.Result{}, res)
		assert.Empty(t, env.logs.calls())
	})

	t.Run("every step is finalized once with its complete log", func(t *testing.T) {
		res, err := fin.Tick(t.Context(), afterWindow())
		require.NoError(t, err)
		assert.Equal(t, finalizer.Result{Done: len(steps(t, env.Store, finished.ID))}, res)
		assertFinalizedOnce(t, env, finished.ID)
		// the skipped step never ran and has no log: it is only marked
		assert.Zero(t, steps(t, env.Store, finished.ID)["on-failure"].Started)
	})

	t.Run("a finalized step is not handed over again", func(t *testing.T) {
		res, err := fin.Tick(t.Context(), afterWindow())
		require.NoError(t, err)
		assert.Equal(t, finalizer.Result{}, res)
	})

	t.Run("a late line of the agent does not reach a finalized log", func(t *testing.T) {
		build := steps(t, env.Store, finished.ID)["build"]
		before, err := logData(t.Context(), env.logs, build)
		require.NoError(t, err)

		// inside the late-log window by time, but the log is finalized
		env.agent.Peer.EnqueueLog(&rpc.LogEntry{StepUUID: build.UUID, Line: 1000, Data: []byte("late line")})
		assert.Never(t, func() bool {
			after, err := logData(t.Context(), env.logs, build)
			return err != nil || after != before
		}, 3*time.Second, 500*time.Millisecond)
	})
}

func TestLogFinalizerCanceledPipeline(t *testing.T) {
	env := startLogFinalizerEnv(t, logFinalizerCancelYAML)
	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err)
	setup.WaitForStepRunning(t, env.Store, created.ID, "long-running")
	running, err := env.Store.GetPipeline(created.ID)
	require.NoError(t, err)
	forge, err := env.Manager.ForgeByID(env.Fixtures.Forge.ID)
	require.NoError(t, err)
	require.NoError(t, pipeline.Cancel(t.Context(), forge, env.Store, env.Fixtures.Repo, env.Fixtures.Owner, running, nil))
	killed := setup.WaitForPipeline(t, env.Store, created.ID)
	setup.WaitForStepStatus(t, env.Store, killed, "long-running", model.StatusKilled, 30*time.Second)
	waitLogsSettled(t, env, created.ID, "long-running")

	res, err := env.finalizer().Tick(t.Context(), afterWindow())
	require.NoError(t, err)
	assert.Equal(t, finalizer.Result{Done: len(steps(t, env.Store, created.ID))}, res)
	assertFinalizedOnce(t, env, created.ID)
	// canceled before it started: no log, only marked
	assert.Zero(t, steps(t, env.Store, created.ID)["after-cancel"].Started)
}

func TestLogFinalizerDeletedLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	env := startLogFinalizerEnv(t, logFinalizerYAML)
	kept := env.runPipeline(t)
	deleted := env.runPipeline(t)
	waitLogsSettled(t, env, kept.ID, "build", "test")
	waitLogsSettled(t, env, deleted.ID, "build", "test")

	// delete the logs of one step, then the whole second pipeline, through the API
	build := steps(t, env.Store, kept.ID)["build"]
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("store", env.Store)
	c.Set("step", build)
	api.DeleteStepLogs(c)
	require.Equal(t, http.StatusNoContent, c.Writer.Status())

	c, _ = gin.CreateTestContext(httptest.NewRecorder())
	c.Set("store", env.Store)
	c.Set("pipeline", deleted)
	api.DeletePipeline(c)
	require.Equal(t, http.StatusNoContent, c.Writer.Status())

	res, err := env.finalizer().Tick(t.Context(), afterWindow())
	require.NoError(t, err)
	// only the steps of the kept pipeline are left, steps of the deleted one are not handed over
	assert.Equal(t, finalizer.Result{Done: len(steps(t, env.Store, kept.ID))}, res)
	assertFinalizedOnce(t, env, kept.ID)
	assert.Equal(t, []string{""}, env.logs.calls()["build"], "the deleted log is handed over empty")
	data, err := logData(t.Context(), env.logs, build)
	require.NoError(t, err)
	assert.Empty(t, data, "the deleted log stays deleted")
}

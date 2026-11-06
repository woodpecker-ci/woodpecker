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
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/e2e/setup"
	forge_types "go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pipeline"
)

// The flaky workflows below fail unless the run gets FLAKY_EXIT=0, which a
// restart can pass as a pipeline parameter. This mimics a step that failed for
// a transient reason and passes when it is retried.

var restartFailedBuildYAML = []byte(`
skip_clone: true
steps:
  - name: compile
    image: dummy
    commands:
      - echo compile
`)

var restartFailedLintYAML = []byte(`
skip_clone: true
steps:
  - name: lint
    image: dummy
    commands:
      - echo lint
`)

var restartFailedTestYAML = []byte(`
skip_clone: true
depends_on:
  - build
steps:
  - name: unit
    image: dummy
    environment:
      STEP_EXIT_CODE: '${FLAKY_EXIT:-1}'
    commands:
      - echo test
`)

var restartFailedDeployYAML = []byte(`
skip_clone: true
depends_on:
  - test
steps:
  - name: ship
    image: dummy
    commands:
      - echo deploy
`)

var restartFailedNotifyYAML = []byte(`
skip_clone: true
depends_on:
  - test
when:
  - status: [failure]
steps:
  - name: notify
    image: dummy
    commands:
      - echo notify
`)

// restartFailedGateYAML only runs when the build fails, so it is skipped in
// the original pipeline and is not rerun.
var restartFailedGateYAML = []byte(`
skip_clone: true
depends_on:
  - build
when:
  - status: [failure]
steps:
  - name: gate
    image: dummy
    commands:
      - echo gate
`)

// restartFailedGuardedYAML needs both the gate and the test to succeed. It was
// skipped in the original pipeline because the gate was skipped.
var restartFailedGuardedYAML = []byte(`
skip_clone: true
depends_on:
  - gate
  - test
steps:
  - name: guarded
    image: dummy
    commands:
      - echo guarded
`)

var restartFailedMatrixYAML = []byte(`
skip_clone: true
matrix:
  include:
    - NAME: good
      CODE: 0
    - NAME: bad
      CODE: 1
steps:
  - name: run
    image: dummy
    environment:
      STEP_EXIT_CODE: '${CODE}'
    commands:
      - echo ${NAME}
`)

func restartFailedFiles(extra ...*forge_types.FileMeta) []*forge_types.FileMeta {
	return append([]*forge_types.FileMeta{
		{Name: ".woodpecker/build.yaml", Data: restartFailedBuildYAML},
		{Name: ".woodpecker/lint.yaml", Data: restartFailedLintYAML},
		{Name: ".woodpecker/test.yaml", Data: restartFailedTestYAML},
		{Name: ".woodpecker/deploy.yaml", Data: restartFailedDeployYAML},
	}, extra...)
}

func workflowStates(t *testing.T, env *setup.ServerEnv, p *model.Pipeline) map[string]model.StatusValue {
	t.Helper()
	workflows, err := env.Store.WorkflowGetTree(p)
	require.NoError(t, err)
	states := make(map[string]model.StatusValue, len(workflows))
	for _, wf := range workflows {
		states[wf.Name] = wf.State
	}
	return states
}

// statusReportedWorkflows returns the names of the workflows whose status was
// reported to the forge for the given pipeline.
func statusReportedWorkflows(env *setup.ServerEnv, pipelineID int64) []string {
	var names []string
	for _, call := range env.Forge.Calls {
		if call.Method != "Status" {
			continue
		}
		p, _ := call.Arguments.Get(3).(*model.Pipeline)
		wf, _ := call.Arguments.Get(4).(*model.Workflow)
		if p == nil || wf == nil || p.ID != pipelineID {
			continue
		}
		if !slices.Contains(names, wf.Name) {
			names = append(names, wf.Name)
		}
	}
	return names
}

func runFailedPipeline(t *testing.T, env *setup.ServerEnv) *model.Pipeline {
	t.Helper()
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	original, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err, "create original pipeline")
	finished := setup.WaitForPipeline(t, env.Store, original.ID)
	require.Equal(t, model.StatusFailure, finished.Status, "original pipeline should fail")
	return finished
}

// TestRestartFailedWorkflows verifies that only the failed workflow and the
// workflows depending on it run again, and that the workflows which already
// passed are neither rerun nor reported to the forge again.
func TestRestartFailedWorkflows(t *testing.T) {
	env := setup.StartServer(t.Context(), t, restartFailedFiles())
	original := runFailedPipeline(t, env)

	assert.Equal(t, map[string]model.StatusValue{
		"build":  model.StatusSuccess,
		"lint":   model.StatusSuccess,
		"test":   model.StatusFailure,
		"deploy": model.StatusSkipped,
	}, workflowStates(t, env, original))

	restarted, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"FLAKY_EXIT": "0"})
	require.NoError(t, err, "restart failed workflows")
	require.NotNil(t, restarted)

	assert.NotEqual(t, original.ID, restarted.ID, "restart should create a new pipeline")
	assert.Equal(t, original.Number, restarted.Parent, "restart.Parent should point at original.Number")
	assert.Equal(t, int64(1), restarted.RerunCount)
	assert.True(t, restarted.PartialRestart, "the pipeline should be marked as a partial restart")

	finished := setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status, "restarted pipeline should succeed")
	assert.Equal(t, map[string]model.StatusValue{
		"test":   model.StatusSuccess,
		"deploy": model.StatusSuccess,
	}, workflowStates(t, env, finished), "only the failed workflow and its dependents should be rerun")

	reported := statusReportedWorkflows(env, restarted.ID)
	assert.NotContains(t, reported, "build", "untouched workflows must keep their forge status")
	assert.NotContains(t, reported, "lint", "untouched workflows must keep their forge status")
	assert.Contains(t, reported, "test")

	// The original pipeline is left as it was.
	assert.Equal(t, model.StatusFailure, workflowStates(t, env, original)["test"])

	// A full restart of the partial restart runs everything again and is no
	// partial restart itself.
	full, err := pipeline.Restart(t.Context(), env.Store, finished, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"FLAKY_EXIT": "0"})
	require.NoError(t, err)
	assert.False(t, full.PartialRestart, "a full restart is no partial restart")
	full = setup.WaitForPipeline(t, env.Store, full.ID)
	assert.Len(t, workflowStates(t, env, full), 4, "a full restart runs all workflows")
}

// TestRestartFailedWorkflowsChain verifies that a restart of a restart, which
// failed again, reruns the same workflows and links to its own parent.
func TestRestartFailedWorkflowsChain(t *testing.T) {
	env := setup.StartServer(t.Context(), t, restartFailedFiles())
	original := runFailedPipeline(t, env)

	first, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.NoError(t, err)
	first = setup.WaitForPipeline(t, env.Store, first.ID)
	require.Equal(t, model.StatusFailure, first.Status, "without FLAKY_EXIT the test fails again")
	require.Equal(t, map[string]model.StatusValue{
		"test":   model.StatusFailure,
		"deploy": model.StatusSkipped,
	}, workflowStates(t, env, first))

	second, err := pipeline.RestartFailed(t.Context(), env.Store, first, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"FLAKY_EXIT": "0"})
	require.NoError(t, err)
	assert.Equal(t, first.Number, second.Parent)
	assert.Equal(t, int64(2), second.RerunCount)

	second = setup.WaitForPipeline(t, env.Store, second.ID)
	assert.Equal(t, model.StatusSuccess, second.Status)
	assert.Equal(t, map[string]model.StatusValue{
		"test":   model.StatusSuccess,
		"deploy": model.StatusSuccess,
	}, workflowStates(t, env, second))
}

// TestRestartFailedWorkflowsOnFailureDependent verifies that a workflow which
// only runs on failure is rerun with the failed workflow and is skipped once
// the failure is gone.
func TestRestartFailedWorkflowsOnFailureDependent(t *testing.T) {
	env := setup.StartServer(t.Context(), t, restartFailedFiles(
		&forge_types.FileMeta{Name: ".woodpecker/notify.yaml", Data: restartFailedNotifyYAML},
	))
	original := runFailedPipeline(t, env)
	require.Equal(t, model.StatusSuccess, workflowStates(t, env, original)["notify"], "notify runs on the failure")

	restarted, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"FLAKY_EXIT": "0"})
	require.NoError(t, err)

	finished := setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)
	assert.Equal(t, map[string]model.StatusValue{
		"test":   model.StatusSuccess,
		"deploy": model.StatusSuccess,
		"notify": model.StatusSkipped,
	}, workflowStates(t, env, finished))
}

// TestRestartFailedWorkflowsKeepsDependencyState verifies that a rerun
// workflow is still gated on the outcome of the dependencies that are not
// rerun: the skipped gate runs again, but only on a failure of build, which
// passed in the original pipeline. So the gate and the guarded workflow
// behind it stay skipped even though the rerun test now passes.
func TestRestartFailedWorkflowsKeepsDependencyState(t *testing.T) {
	env := setup.StartServer(t.Context(), t, restartFailedFiles(
		&forge_types.FileMeta{Name: ".woodpecker/gate.yaml", Data: restartFailedGateYAML},
		&forge_types.FileMeta{Name: ".woodpecker/guarded.yaml", Data: restartFailedGuardedYAML},
	))
	original := runFailedPipeline(t, env)
	states := workflowStates(t, env, original)
	require.Equal(t, model.StatusSkipped, states["gate"])
	require.Equal(t, model.StatusSkipped, states["guarded"])

	restarted, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"FLAKY_EXIT": "0"})
	require.NoError(t, err)

	finished := setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, map[string]model.StatusValue{
		"test":    model.StatusSuccess,
		"deploy":  model.StatusSuccess,
		"gate":    model.StatusSkipped,
		"guarded": model.StatusSkipped,
	}, workflowStates(t, env, finished))
}

// restartFailedSlowYAML keeps the agent busy until the pipeline is canceled,
// unless the run gets WORK_TIME.
var restartFailedSlowYAML = []byte(`
skip_clone: true
steps:
  - name: work
    image: dummy
    environment:
      SLEEP: '${WORK_TIME:-30s}'
    commands:
      - echo working
`)

// TestRestartFailedWorkflowsCanceled verifies that canceling a pipeline makes
// all workflows that did not finish run again, including the ones that never
// started: canceling marks those as skipped.
func TestRestartFailedWorkflowsCanceled(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/one.yaml", Data: restartFailedSlowYAML},
		{Name: ".woodpecker/two.yaml", Data: restartFailedSlowYAML},
	})
	// One workflow at a time, so one runs and the other waits in the queue.
	agent := setup.StartAgent(t, env.GRPCAddr, setup.WithCapacity(1))
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err)
	setup.WaitForStepRunning(t, env.Store, created.ID, "work")

	forge, err := env.Manager.ForgeByID(env.Fixtures.Forge.ID)
	require.NoError(t, err)
	running, err := env.Store.GetPipeline(created.ID)
	require.NoError(t, err)
	require.NoError(t, pipeline.Cancel(t.Context(), forge, env.Store, env.Fixtures.Repo, env.Fixtures.Owner, running, nil))
	canceled := setup.WaitForPipeline(t, env.Store, created.ID)

	// The pipeline is marked as canceled right away, the running workflow once
	// the agent stopped it.
	var states map[string]model.StatusValue
	require.Eventually(t, func() bool {
		states = workflowStates(t, env, canceled)
		return states["one"] != model.StatusRunning && states["two"] != model.StatusRunning
	}, 10*time.Second, 50*time.Millisecond)
	require.ElementsMatch(t, []model.StatusValue{model.StatusKilled, model.StatusSkipped},
		[]model.StatusValue{states["one"], states["two"]}, "one workflow was running, the other never started")

	restarted, err := pipeline.RestartFailed(t.Context(), env.Store, canceled, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"WORK_TIME": "0s"})
	require.NoError(t, err)
	finished := setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, model.StatusSuccess, finished.Status)
	assert.Equal(t, map[string]model.StatusValue{
		"one": model.StatusSuccess,
		"two": model.StatusSuccess,
	}, workflowStates(t, env, finished), "the workflow that never started is not taken over as passed")
}

// TestRestartFailedWorkflowsMatrix verifies that only the failed variant of a
// matrix workflow is rerun.
func TestRestartFailedWorkflowsMatrix(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/matrix.yaml", Data: restartFailedMatrixYAML},
	})
	original := runFailedPipeline(t, env)

	restarted, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.NoError(t, err)
	finished := setup.WaitForPipeline(t, env.Store, restarted.ID)

	workflows, err := env.Store.WorkflowGetTree(finished)
	require.NoError(t, err)
	require.Len(t, workflows, 1, "only the failed variant should be rerun")
	assert.Equal(t, "bad", workflows[0].Environ["NAME"])
	assert.Equal(t, model.StatusFailure, workflows[0].State)
}

// TestRestartFailedWorkflowsRejected verifies the cases where there is
// nothing to restart.
func TestRestartFailedWorkflowsRejected(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/build.yaml", Data: restartFailedBuildYAML},
	})
	agent := setup.StartAgent(t, env.GRPCAddr)
	setup.WaitForAgentRegistered(t, env.Store, agent)

	created, err := pipeline.Create(t.Context(), env.Store, env.Fixtures.Repo, env.DummyPipeline(model.EventPush))
	require.NoError(t, err)
	succeeded := setup.WaitForPipeline(t, env.Store, created.ID)
	require.Equal(t, model.StatusSuccess, succeeded.Status)

	_, err = pipeline.RestartFailed(t.Context(), env.Store, succeeded, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	var badRequest *pipeline.ErrBadRequest
	require.ErrorAs(t, err, &badRequest, "a pipeline without failed workflows has nothing to restart")

	blocked := *succeeded
	blocked.Status = model.StatusBlocked
	_, err = pipeline.RestartFailed(t.Context(), env.Store, &blocked, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.ErrorAs(t, err, &badRequest, "a blocked pipeline cannot be restarted")
}

// restartFailedReportYAML only runs its step if the partial restart tells it
// that build and lint were taken over from pipeline #1.
var restartFailedReportYAML = []byte(`
skip_clone: true
depends_on:
  - test
steps:
  - name: report
    image: dummy
    when:
      - evaluate: 'CI_PIPELINE_REUSED_WORKFLOWS == "build=1,lint=1"'
    commands:
      - echo report
`)

func stepState(t *testing.T, env *setup.ServerEnv, p *model.Pipeline, name string) model.StatusValue {
	t.Helper()
	workflows, err := env.Store.WorkflowGetTree(p)
	require.NoError(t, err)
	for _, wf := range workflows {
		for _, step := range wf.Children {
			if step.Name == name {
				return step.State
			}
		}
	}
	return ""
}

// TestRestartFailedWorkflowsReusedEnv verifies that the steps of a partial
// restart learn where the workflows that were not rerun ran, also through a
// chain of partial restarts.
func TestRestartFailedWorkflowsReusedEnv(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/build.yaml", Data: restartFailedBuildYAML},
		{Name: ".woodpecker/lint.yaml", Data: restartFailedLintYAML},
		{Name: ".woodpecker/test.yaml", Data: restartFailedTestYAML},
		{Name: ".woodpecker/report.yaml", Data: restartFailedReportYAML},
	})
	original := runFailedPipeline(t, env)
	require.Equal(t, int64(1), original.Number)

	first, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.NoError(t, err)
	first = setup.WaitForPipeline(t, env.Store, first.ID)
	require.Equal(t, model.StatusFailure, first.Status, "without FLAKY_EXIT the test fails again")

	second, err := pipeline.RestartFailed(t.Context(), env.Store, first, env.Fixtures.Owner, env.Fixtures.Repo,
		map[string]string{"FLAKY_EXIT": "0"})
	require.NoError(t, err)
	second = setup.WaitForPipeline(t, env.Store, second.ID)
	assert.Equal(t, model.StatusSuccess, second.Status)
	assert.Equal(t, map[string]model.StatusValue{
		"test":   model.StatusSuccess,
		"report": model.StatusSuccess,
	}, workflowStates(t, env, second), "build and lint passed in #1 and are not rerun")
	assert.Equal(t, model.StatusSuccess, stepState(t, env, second, "report"),
		"the step sees build and lint as taken over from #1")
}

// restartFailedMatrixEnvYAML checks, in the variant that is rerun, that the
// variant which passed is named by its axis.
var restartFailedMatrixEnvYAML = []byte(`
skip_clone: true
matrix:
  include:
    - NAME: good
      CODE: 0
    - NAME: bad
      CODE: 1
steps:
  - name: check
    image: dummy
    when:
      - evaluate: 'CI_PIPELINE_REUSED_WORKFLOWS == "matrix/1=1"'
    commands:
      - echo check
  - name: run
    image: dummy
    environment:
      STEP_EXIT_CODE: '${CODE}'
    commands:
      - echo ${NAME}
`)

// TestRestartFailedWorkflowsReusedEnvMatrix verifies that the variants of a
// matrix workflow are listed one by one, as they can run in different
// pipelines.
func TestRestartFailedWorkflowsReusedEnvMatrix(t *testing.T) {
	env := setup.StartServer(t.Context(), t, []*forge_types.FileMeta{
		{Name: ".woodpecker/matrix.yaml", Data: restartFailedMatrixEnvYAML},
	})
	original := runFailedPipeline(t, env)
	require.Equal(t, int64(1), original.Number)

	restarted, err := pipeline.RestartFailed(t.Context(), env.Store, original, env.Fixtures.Owner, env.Fixtures.Repo, nil)
	require.NoError(t, err)
	restarted = setup.WaitForPipeline(t, env.Store, restarted.ID)
	assert.Equal(t, model.StatusSuccess, stepState(t, env, restarted, "check"),
		"the rerun variant sees the passed variant as matrix/1 from #1")
}

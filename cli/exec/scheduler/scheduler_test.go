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

package scheduler

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/constraint"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
)

func item(name string, dependsOn ...string) *builder.Item {
	it := &builder.Item{Workflow: &builder.Workflow{Name: name}}
	for _, dep := range dependsOn {
		it.DependsOn = append(it.DependsOn, constraint.Dependency{Name: dep})
	}
	return it
}

// fakeRunner records what the scheduler does and ends each workflow with the
// status configured for its name (success if none).
type fakeRunner struct {
	sync.Mutex
	results map[string]status.Value
	// block, if set, is waited on by every run
	block chan struct{}
	// onRun, if set, is called at the start of every run
	onRun func(name string)

	events     []string
	running    int
	maxRunning int
}

func (r *fakeRunner) Run(_ context.Context, item *builder.Item) status.Value {
	r.Lock()
	r.events = append(r.events, "run "+item.Workflow.Name)
	r.running++
	r.maxRunning = max(r.maxRunning, r.running)
	r.Unlock()

	if r.onRun != nil {
		r.onRun(item.Workflow.Name)
	}
	if r.block != nil {
		<-r.block
	}

	r.Lock()
	defer r.Unlock()
	r.running--
	r.events = append(r.events, "end "+item.Workflow.Name)
	if s, ok := r.results[item.Workflow.Name]; ok {
		return s
	}
	return status.Success
}

func (r *fakeRunner) Skip(item *builder.Item) {
	r.Lock()
	defer r.Unlock()
	r.events = append(r.events, "skip "+item.Workflow.Name)
}

func TestRunOrdersByDependencies(t *testing.T) {
	r := &fakeRunner{}
	Run(t.Context(), []*builder.Item{
		item("deploy", "build", "test"),
		item("test", "build"),
		item("build"),
	}, 4, r)

	assert.Equal(t, []string{
		"run build", "end build",
		"run test", "end test",
		"run deploy", "end deploy",
	}, r.events)
}

func TestRunSkipsDependentsOfFailedWorkflow(t *testing.T) {
	r := &fakeRunner{results: map[string]status.Value{"build": status.Failure}}
	Run(t.Context(), []*builder.Item{
		item("build"),
		item("test", "build"),
		item("deploy", "test"),
	}, 1, r)

	assert.Equal(t, []string{"run build", "end build", "skip test", "skip deploy"}, r.events)
}

func TestRunRespectsRunsOn(t *testing.T) {
	notify := item("notify", "build")
	notify.RunsOn = []string{"failure"}
	cleanup := item("cleanup", "build")
	cleanup.RunsOn = []string{"success", "failure"}
	items := []*builder.Item{item("build"), notify, cleanup}

	t.Run("dependency failed", func(t *testing.T) {
		r := &fakeRunner{results: map[string]status.Value{"build": status.Failure}}
		Run(t.Context(), items, 1, r)
		assert.Equal(t, []string{"run build", "end build", "run notify", "end notify", "run cleanup", "end cleanup"}, r.events)
	})

	t.Run("dependency succeeded", func(t *testing.T) {
		r := &fakeRunner{}
		Run(t.Context(), items, 1, r)
		assert.Equal(t, []string{"run build", "end build", "skip notify", "run cleanup", "end cleanup"}, r.events)
	})
}

func TestRunWaitsForAllMatrixWorkflowsOfADependency(t *testing.T) {
	r := &fakeRunner{results: map[string]status.Value{}}
	// the second "test" workflow fails, so "deploy" must not run
	calls := 0
	r.onRun = func(name string) {
		if name != "test" {
			return
		}
		r.Lock()
		defer r.Unlock()
		calls++
		if calls == 2 {
			r.results["test"] = status.Failure
		}
	}
	Run(t.Context(), []*builder.Item{item("test"), item("test"), item("deploy", "test")}, 1, r)

	assert.Equal(t, []string{"run test", "end test", "run test", "end test", "skip deploy"}, r.events)
}

func TestRunLimitsParallelWorkflows(t *testing.T) {
	r := &fakeRunner{}
	Run(t.Context(), []*builder.Item{item("a"), item("b"), item("c"), item("d"), item("e")}, 2, r)

	assert.Len(t, r.events, 10)
	assert.LessOrEqual(t, r.maxRunning, 2)
}

func TestRunRespectsConcurrencyLimit(t *testing.T) {
	limited := func(name, group string) *builder.Item {
		it := item(name)
		it.ConcurrencyLimit = 1
		it.ConcurrencyGroup = group
		return it
	}

	t.Run("same workflow name", func(t *testing.T) {
		r := &fakeRunner{}
		Run(t.Context(), []*builder.Item{limited("test", ""), limited("test", ""), limited("test", "")}, 4, r)
		assert.Equal(t, 1, r.maxRunning)
		assert.Len(t, r.events, 6)
	})

	t.Run("same group", func(t *testing.T) {
		r := &fakeRunner{}
		Run(t.Context(), []*builder.Item{limited("a", "deploy"), limited("b", "deploy")}, 4, r)
		assert.Equal(t, 1, r.maxRunning)
	})

	t.Run("different groups", func(t *testing.T) {
		started := make(chan struct{}, 2)
		r := &fakeRunner{block: make(chan struct{}), onRun: func(string) { started <- struct{}{} }}
		go func() {
			// both must run at the same time
			<-started
			<-started
			close(r.block)
		}()
		Run(t.Context(), []*builder.Item{limited("a", "x"), limited("b", "y")}, 4, r)
		assert.Equal(t, 2, r.maxRunning)
	})
}

func TestRunStartsNothingAfterCancel(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	r := &fakeRunner{onRun: func(string) { cancel(nil) }}
	Run(ctx, []*builder.Item{item("build"), item("test", "build"), item("lint")}, 1, r)

	assert.Equal(t, []string{"run build", "end build"}, r.events)
}

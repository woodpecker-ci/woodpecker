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

package exec

import (
	"context"
	"fmt"
	"io"
	"maps"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/cli/exec/scheduler"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline"
	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_errors "go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	pipeline_runtime "go.woodpecker-ci.org/woodpecker/v3/pipeline/runtime"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/state"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/tracing"
	pipeline_utils "go.woodpecker-ci.org/woodpecker/v3/pipeline/utils"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// execBackend is a backend to execute workflows on, with the labels an agent
// running that backend would have.
type execBackend struct {
	backend_types.Backend
	labels map[string]string
}

// selectBackend returns the backend whose labels fit the ones of a workflow
// best, the first one if several fit equally well, or nil if none fits.
func selectBackend(backends []execBackend, workflowLabels map[string]string) backend_types.Backend {
	var best backend_types.Backend
	bestScore := -1
	for _, b := range backends {
		if matches, score := pipeline.MatchLabels(workflowLabels, b.labels); matches && score > bestScore {
			best, bestScore = b.Backend, score
		}
	}
	return best
}

// output is where a pipeline run reports to.
type output interface {
	// Workflow is called with a copy of a workflow each time the status of it
	// or of one of its steps changed.
	Workflow(workflow *woodpecker.Workflow)
	// Log is called for each log line of a step.
	Log(entry *woodpecker.LogEntry)
	// Message is called with what the user should know about the run itself.
	Message(text string)
}

// pipelineRun executes the workflows of a pipeline and keeps their status.
// It does to the local runtime what an agent and the server do together:
// its methods mirror the rpc calls of an agent and how the server handles them.
type pipelineRun struct {
	items []*builder.Item
	// the backend each workflow is executed on, none if no backend fits it
	backends map[*builder.Item]backend_types.Backend
	// what to tell about the workflows no backend fits
	unfit map[*builder.Item]string
	// timeout of a single workflow
	timeout time.Duration
	out     output
	// cancelWorkflows stops all running workflows
	cancelWorkflows context.CancelCauseFunc

	// steps by their uuid
	steps map[string]*woodpecker.Step
	// what a failure of a step should cause, by step id
	failure map[int64]string

	// guards the status and the content of the workflows and steps
	mu        sync.Mutex
	status    status.Value
	workflows map[*builder.Item]*woodpecker.Workflow
}

// newPipelineRun returns a run for the given workflows, all pending.
// The given cancelWorkflows has to cancel the context the workflows are run with.
func newPipelineRun(items []*builder.Item, backends []execBackend, timeout time.Duration, cancelWorkflows context.CancelCauseFunc) *pipelineRun {
	r := &pipelineRun{
		items:           items,
		backends:        make(map[*builder.Item]backend_types.Backend, len(items)),
		unfit:           make(map[*builder.Item]string),
		timeout:         timeout,
		cancelWorkflows: cancelWorkflows,
		steps:           make(map[string]*woodpecker.Step),
		failure:         make(map[int64]string),
		status:          status.Pending,
		workflows:       make(map[*builder.Item]*woodpecker.Workflow, len(items)),
	}

	var pid int
	for _, item := range items {
		pid = max(pid, item.Workflow.PID)
	}

	for _, item := range items {
		workflow := &woodpecker.Workflow{
			ID:      int64(item.Workflow.PID),
			PID:     item.Workflow.PID,
			Name:    item.Workflow.Name,
			State:   string(status.Pending),
			Environ: item.Workflow.Environ,
		}
		for _, stage := range item.Config.Stages {
			for _, s := range stage.Steps {
				pid++
				step := &woodpecker.Step{
					ID:    int64(pid),
					PID:   pid,
					PPID:  item.Workflow.PID,
					Name:  s.Name,
					State: string(status.Pending),
					Type:  woodpecker.StepType(s.Type),
				}
				workflow.Children = append(workflow.Children, step)
				r.steps[s.UUID] = step
				r.failure[step.ID] = s.Failure
			}
		}
		r.workflows[item] = workflow

		if backend := selectBackend(backends, item.Labels); backend != nil {
			r.backends[item] = backend
		} else {
			r.unfit[item] = unfitMessage(item, backends)
		}
	}

	return r
}

// unfitMessage tells why a workflow is not executed on any of the backends.
func unfitMessage(item *builder.Item, backends []execBackend) string {
	var labels []string
	for _, label := range slices.Sorted(maps.Keys(item.Labels)) {
		if !strings.HasPrefix(label, pipeline.InternalLabelPrefix) && item.Labels[label] != "" {
			labels = append(labels, label+"="+item.Labels[label])
		}
	}
	names := make([]string, 0, len(backends))
	for _, backend := range backends {
		names = append(names, backend.Name())
	}
	return fmt.Sprintf("workflow %s is skipped: its labels (%s) match none of the backends in use (%s), run it anyway with --ignore-labels",
		item.Workflow.Name, strings.Join(labels, ", "), strings.Join(names, ", "))
}

// execute runs all workflows and returns the status of the pipeline.
// The context has to be the one the run can cancel.
func (r *pipelineRun) execute(ctx context.Context) status.Value {
	// report everything as pending first
	r.mu.Lock()
	for _, item := range r.items {
		r.changed(r.workflows[item])
	}
	r.mu.Unlock()

	// A server would let these wait for an agent that fits, here none will come.
	for _, item := range r.items {
		if message, unfit := r.unfit[item]; unfit {
			r.out.Message(message)
			r.Skip(item)
		}
	}

	scheduler.Run(ctx, r.items, runtime.NumCPU(), r)

	// if workflows are left, the context got canceled
	r.cancel()

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.status
}

// Run executes a workflow like an agent does.
func (r *pipelineRun) Run(ctx context.Context, item *builder.Item) status.Value {
	workflow := r.workflows[item]

	state := status.WorkflowState{Started: time.Now().Unix()}
	if !r.init(workflow, state) {
		return status.Skipped
	}

	workflowCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	// the runtime needs a context that outlives the cancellation to clean up
	err := pipeline_runtime.New(
		item.Config, r.backends[item],
		pipeline_runtime.WithContext(workflowCtx),
		pipeline_runtime.WithLogger(r.log),
		pipeline_runtime.WithTracer(tracing.TraceFunc(r.update)),
		pipeline_runtime.WithDescription(map[string]string{
			"CLI": "exec",
		}),
	).Run(context.WithoutCancel(ctx))

	state.Finish(err)
	return r.done(workflow, state)
}

// Skip ends a workflow that must not run because of its dependencies.
func (r *pipelineRun) Skip(item *builder.Item) {
	r.done(r.workflows[item], status.WorkflowState{})
}

// init marks a workflow as running. It returns false if the workflow must not
// start anymore as the pipeline got canceled.
func (r *pipelineRun) init(workflow *woodpecker.Workflow, state status.WorkflowState) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if workflow.State != string(status.Pending) {
		return false
	}

	r.status = status.Running
	workflow.State = string(status.Running)
	workflow.Started = state.Started
	r.changed(workflow)
	return true
}

// update applies the state the runtime reports for a step.
func (r *pipelineRun) update(s *state.State) error {
	cancelPipeline, err := func() (bool, error) {
		r.mu.Lock()
		defer r.mu.Unlock()

		step, ok := r.steps[s.CurrStep.UUID]
		if !ok {
			return false, nil
		}
		updated, cancelPipeline, err := r.stepStatus(step).Update(status.NewStepState(s))
		if err != nil {
			return false, err
		}
		setStepStatus(step, updated)
		r.changed(r.workflowOf(step))
		return cancelPipeline, nil
	}()

	if cancelPipeline {
		r.cancel()
	}
	return err
}

// done ends a workflow with the state its runtime finished with.
func (r *pipelineRun) done(workflow *woodpecker.Workflow, state status.WorkflowState) status.Value {
	r.mu.Lock()
	defer r.mu.Unlock()

	steps := make([]status.Step, 0, len(workflow.Children))
	for _, step := range workflow.Children {
		s := r.stepStatus(step)
		// end what the runtime did not report an end for, e.g. services
		if s.State.IsActive() {
			s = s.End(status.Killed, state.Finished)
			setStepStatus(step, s)
		}
		steps = append(steps, s)
	}

	workflow.State = string(status.WorkflowDone(steps, state))
	workflow.Stopped = state.Finished
	workflow.Error = state.Error
	r.changed(workflow)

	// the pipeline is done with its last workflow
	pipelineStatus := status.Success
	for _, w := range r.workflows {
		if status.Value(w.State).IsActive() {
			return status.Value(workflow.State)
		}
		pipelineStatus = status.Merge(pipelineStatus, status.Value(w.State))
	}
	r.status = pipelineStatus

	return status.Value(workflow.State)
}

// cancel stops the running workflows and ends everything not yet started.
func (r *pipelineRun) cancel() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.status.IsActive() {
		return
	}
	r.cancelWorkflows(pipeline_errors.ErrCancel)

	r.status = status.Canceled
	for _, workflow := range r.workflows {
		if workflow.State == string(status.Pending) {
			workflow.State = string(status.Skipped)
		} else {
			// running workflows end as soon as their runtime has stopped
			r.status = status.Killed
		}
		for _, step := range workflow.Children {
			if step.State == string(status.Pending) {
				setStepStatus(step, r.stepStatus(step).End(status.Canceled, 0))
			}
		}
		r.changed(workflow)
	}
}

// changed reports a changed workflow to the output.
// The caller has to hold the lock.
func (r *pipelineRun) changed(workflow *woodpecker.Workflow) {
	c := *workflow
	c.Children = make([]*woodpecker.Step, 0, len(workflow.Children))
	for _, step := range workflow.Children {
		s := *step
		c.Children = append(c.Children, &s)
	}
	r.out.Workflow(&c)
}

func (r *pipelineRun) workflowOf(step *woodpecker.Step) *woodpecker.Workflow {
	for _, workflow := range r.workflows {
		if workflow.PID == step.PPID {
			return workflow
		}
	}
	return nil
}

// stepStatus returns the part of the step its status is calculated from.
func (r *pipelineRun) stepStatus(step *woodpecker.Step) status.Step {
	return status.Step{
		State:    status.Value(step.State),
		Failure:  r.failure[step.ID],
		Started:  step.Started,
		Finished: step.Stopped,
		ExitCode: step.ExitCode,
		Error:    step.Error,
	}
}

// setStepStatus writes the calculated status back to the step.
func setStepStatus(step *woodpecker.Step, s status.Step) {
	step.State = string(s.State)
	step.Started = s.Started
	step.Stopped = s.Finished
	step.ExitCode = s.ExitCode
	step.Error = s.Error
}

// log sends the log of a step line by line to the output.
func (r *pipelineRun) log(step *backend_types.Step, rc io.ReadCloser) error {
	return pipeline_utils.CopyLineByLine(&logWriter{
		out:     r.out,
		stepID:  r.steps[step.UUID].ID,
		started: time.Now(),
	}, rc, pipeline.MaxLogLineLength)
}

// logWriter turns each write into a log entry, like the agent does.
type logWriter struct {
	out     output
	stepID  int64
	line    int
	started time.Time
}

func (w *logWriter) Write(p []byte) (int, error) {
	w.out.Log(&woodpecker.LogEntry{
		StepID: w.stepID,
		Time:   int64(time.Since(w.started).Seconds()),
		Line:   w.line,
		Data:   []byte(strings.TrimSuffix(string(p), "\n")),
	})
	w.line++
	return len(p), nil
}

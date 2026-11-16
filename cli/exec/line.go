// Copyright 2022 Woodpecker Authors
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
	"fmt"
	"io"
	"sync"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
	"go.woodpecker-ci.org/woodpecker/v3/woodpecker-go/woodpecker"
)

// lineOutput prints what a pipeline run reports line by line.
type lineOutput struct {
	out io.Writer
	// print the workflow name in front of each log line
	multiWorkflow bool

	mu sync.Mutex
	// last printed state by workflow id
	states map[int64]string
	// log line prefix by step id
	prefixes map[int64]string
}

func newLineOutput(out io.Writer, multiWorkflow bool) *lineOutput {
	return &lineOutput{
		out:           out,
		multiWorkflow: multiWorkflow,
		states:        make(map[int64]string),
		prefixes:      make(map[int64]string),
	}
}

// Workflow prints a workflow when it starts and how it ended.
func (o *lineOutput) Workflow(workflow *woodpecker.Workflow) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.states[workflow.ID] == workflow.State {
		return
	}
	o.states[workflow.ID] = workflow.State

	switch status.Value(workflow.State) {
	case status.Running:
		fmt.Fprintln(o.out, "#", workflow.Name)
		for _, step := range workflow.Children {
			o.prefixes[step.ID] = step.Name
			if o.multiWorkflow {
				o.prefixes[step.ID] = workflow.Name + "/" + step.Name
			}
		}
	case status.Pending, status.Success:
	default:
		fmt.Fprintf(o.out, "# %s: %s\n", workflow.Name, workflow.State)
		if workflow.Error != "" {
			fmt.Fprintln(o.out, workflow.Error)
		}
		for _, step := range workflow.Children {
			if status.Value(step.State) != status.Failure {
				continue
			}
			if step.Error != "" {
				fmt.Fprintf(o.out, "step %s: %s\n", step.Name, step.Error)
			} else {
				fmt.Fprintf(o.out, "step %s: exit code %d\n", step.Name, step.ExitCode)
			}
		}
	}
}

// Log prints a log line of a step.
func (o *lineOutput) Log(entry *woodpecker.LogEntry) {
	o.mu.Lock()
	defer o.mu.Unlock()
	fmt.Fprintf(o.out, "[%s:L%d:%ds] %s\n", o.prefixes[entry.StepID], entry.Line, entry.Time, entry.Data)
}

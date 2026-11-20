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

// Package scheduler runs the workflows of a pipeline in the order the server
// queue would hand them to agents.
package scheduler

import (
	"context"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/status"
)

// Runner executes the workflows the scheduler hands out.
type Runner interface {
	// Run executes a workflow and returns the status it ended with.
	// It is called concurrently.
	Run(ctx context.Context, item *builder.Item) status.Value
	// Skip is called instead of Run for a workflow that must not run because
	// of the status of its dependencies.
	Skip(item *builder.Item)
}

// Run executes all workflows and returns once none is left to start or still
// running. A workflow starts after all workflows it depends on have finished,
// while at most parallel workflows run at once. If ctx gets canceled no
// further workflow is started.
func Run(ctx context.Context, items []*builder.Item, parallel int, runner Runner) {
	states := make([]status.Value, len(items))
	for i := range states {
		states[i] = status.Pending
	}

	type result struct {
		item  int
		state status.Value
	}
	done := make(chan result)
	running := 0

	for {
		skipped := false
		for i, item := range items {
			if states[i] != status.Pending || ctx.Err() != nil {
				continue
			}

			deps, finished := dependencies(items, states, item)
			switch {
			case !finished:
			case !status.ShouldRun(item.RunsOn, deps):
				states[i] = status.Skipped
				runner.Skip(item)
				skipped = true
			case running < parallel && hasFreeSlot(items, states, item):
				states[i] = status.Running
				running++
				go func() {
					done <- result{i, runner.Run(ctx, item)}
				}()
			}
		}

		// a skipped workflow may have unblocked workflows we already looked at
		if skipped {
			continue
		}
		if running == 0 {
			return
		}
		r := <-done
		states[r.item] = r.state
		running--
	}
}

// dependencies returns the status of the workflows the item depends on
// and if all of them have finished.
func dependencies(items []*builder.Item, states []status.Value, item *builder.Item) (deps []status.Value, finished bool) {
	for _, name := range item.DependsOn.Names() {
		for i, dep := range items {
			if dep.Workflow.Name != name {
				continue
			}
			if states[i].IsActive() {
				return nil, false
			}
			deps = append(deps, states[i])
		}
	}
	return deps, true
}

// hasFreeSlot reports if starting the item respects its concurrency limit.
func hasFreeSlot(items []*builder.Item, states []status.Value, item *builder.Item) bool {
	if item.ConcurrencyLimit <= 0 {
		return true
	}

	running := 0
	for i, other := range items {
		if states[i] == status.Running && concurrencyGroup(other) == concurrencyGroup(item) {
			running++
		}
	}
	return running < item.ConcurrencyLimit
}

// concurrencyGroup returns what identifies the workflows limited against each
// other: the explicit group, or else all workflows of the same name.
func concurrencyGroup(item *builder.Item) [2]string {
	if item.ConcurrencyGroup != "" {
		return [2]string{"", item.ConcurrencyGroup}
	}
	return [2]string{item.Workflow.Name, ""}
}

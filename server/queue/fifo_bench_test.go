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

package queue

import (
	"container/list"
	"fmt"
	"testing"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// BenchmarkFilterWaiting measures steady-state filtering, including its own
// allocations, without fixture construction or the scheduler's timer.
func BenchmarkFilterWaiting(b *testing.B) {
	for _, shape := range []string{"independent", "running-fanout", "chain", "mixed"} {
		for _, size := range []int{100, 1000, 10000} {
			b.Run(fmt.Sprintf("%s/%d", shape, size), func(b *testing.B) {
				q := &fifo{
					pending:       list.New(),
					waitingOnDeps: list.New(),
					running:       make(map[string]*entry),
					queuedIDs:     make(map[string]struct{}, 128),
				}
				q.running["root"] = &entry{item: &model.Task{ID: "root"}}
				ready := 0
				for i := range size {
					task := &model.Task{ID: fmt.Sprintf("task-%d", i)}
					switch shape {
					case "running-fanout":
						task.Dependencies = []string{"root"}
					case "chain":
						if i > 0 {
							task.Dependencies = []string{fmt.Sprintf("task-%d", i-1)}
						}
					case "mixed":
						if i%2 != 0 {
							task.Dependencies = []string{"root"}
						}
					}
					if len(task.Dependencies) == 0 {
						ready++
					}
					q.pending.PushBack(task)
				}
				q.filterWaiting()
				b.ReportAllocs()
				for b.Loop() {
					q.filterWaiting()
				}
				if q.pending.Len() != ready || q.waitingOnDeps.Len() != size-ready {
					b.Fatal("filtering changed the steady-state partition")
				}
			})
		}
	}
}

// BenchmarkAssignToWorker measures one scheduling attempt on a concurrency
// group backlog that cannot be served: either no idle worker matches its
// labels, or an earlier pipeline still waiting on dependencies holds the slot.
func BenchmarkAssignToWorker(b *testing.B) {
	for _, shape := range []string{"unmatched", "reserved"} {
		for _, size := range []int{100, 1000, 10000} {
			b.Run(fmt.Sprintf("%s/%d", shape, size), func(b *testing.B) {
				matched := shape == "reserved"
				q := &fifo{
					pending:       list.New(),
					waitingOnDeps: list.New(),
					running:       make(map[string]*entry),
					workers:       make(map[*worker]struct{}),
				}
				q.workers[&worker{filter: func(*model.Task) (bool, int) { return matched, 1 }}] = struct{}{}
				q.waitingOnDeps.PushBack(&model.Task{ID: "early", ConcurrencyGroup: "deploy", ConcurrencyLimit: 1})
				for i := range size {
					q.pending.PushBack(&model.Task{
						ID:               fmt.Sprintf("task-%d", i),
						PipelineID:       int64(i + 1),
						Created:          int64(i + 1),
						ConcurrencyGroup: "deploy",
						ConcurrencyLimit: 1,
					})
				}
				b.ReportAllocs()
				for b.Loop() {
					if element, _ := q.assignToWorker(); element != nil {
						b.Fatal("the backlog must stay blocked")
					}
				}
			})
		}
	}
}

// BenchmarkSchedule measures one full scheduling pass, the time the queue lock
// is held per tick, on a backlog of five-workflow pipelines: two independent
// builds, a test, a deploy limited to one at a time across pipelines and a
// notification. Pipelines alternate between two platforms. An older deploy is
// running, so every deploy and notification waits.
//
// busy: all agents are occupied, the steady state under load.
// dispatch: eight agents, four per platform, are idle and get work.
func BenchmarkSchedule(b *testing.B) {
	for _, idle := range []int{0, 8} {
		name := "busy"
		if idle > 0 {
			name = "dispatch"
		}
		for _, size := range []int{100, 1000, 10000} {
			b.Run(fmt.Sprintf("%s/%d", name, size), func(b *testing.B) {
				b.ReportAllocs()
				for b.Loop() {
					b.StopTimer()
					q := newPipelineBacklog(size/5, idle)
					running := len(q.running)
					b.StartTimer()

					q.schedule()

					b.StopTimer()
					if len(q.running) != running+idle {
						b.Fatalf("dispatched %d tasks, want %d", len(q.running)-running, idle)
					}
					b.StartTimer()
				}
			})
		}
	}
}

// newPipelineBacklog returns a queue that already ran a scheduling pass over
// the given number of pipelines, with the given number of idle agents.
func newPipelineBacklog(pipelines, idle int) *fifo {
	platform := func(i int) string { return []string{"amd64", "arm64"}[i%2] }
	q := &fifo{
		pending:       list.New(),
		waitingOnDeps: list.New(),
		running:       make(map[string]*entry),
		workers:       make(map[*worker]struct{}),
		extension:     time.Hour,
		queuedIDs:     make(map[string]struct{}, 128),
	}
	q.running["deploy-old"] = &entry{
		item:     &model.Task{ID: "deploy-old", Name: "deploy", ConcurrencyGroup: "deploy", ConcurrencyLimit: 1},
		done:     make(chan bool),
		deadline: time.Now().Add(time.Hour),
	}
	for i := range idle {
		want := platform(i)
		q.workers[&worker{
			agentID: int64(i + 1),
			channel: make(chan *model.Task, 1),
			filter:  func(task *model.Task) (bool, int) { return task.Labels["platform"] == want, 1 },
		}] = struct{}{}
	}
	for p := range pipelines {
		id := func(name string) string { return fmt.Sprintf("%d-%s", p, name) }
		for _, w := range []struct {
			name string
			deps []string
		}{
			{"build", nil},
			{"lint", nil},
			{"test", []string{id("build")}},
			{"deploy", []string{id("test"), id("lint")}},
			{"notify", []string{id("deploy")}},
		} {
			task := &model.Task{
				ID:           id(w.name),
				Name:         w.name,
				PipelineID:   int64(p + 1),
				Created:      int64(p + 1),
				Labels:       map[string]string{"platform": platform(p)},
				Dependencies: w.deps,
				DepStatus:    make(map[string]model.StatusValue),
			}
			if w.name == "deploy" {
				task.ConcurrencyGroup = "deploy"
				task.ConcurrencyLimit = 1
			}
			q.pending.PushBack(task)
		}
	}
	q.filterWaiting()
	return q
}

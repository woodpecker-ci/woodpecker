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

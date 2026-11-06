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

	"github.com/rs/zerolog"

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

// BenchmarkAssignToWorker measures one scheduling attempt, including index
// construction, on backlogs blocked by labels, reservations, or running slots.
func BenchmarkAssignToWorker(b *testing.B) {
	for _, shape := range []string{"unmatched", "reserved", "unlimited", "many-groups", "same-pipeline", "running-full", "no-workers", "shuffled"} {
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
				if shape == "unmatched" || shape == "reserved" {
					q.waitingOnDeps.PushBack(&model.Task{ID: "early", ConcurrencyGroup: "deploy", ConcurrencyLimit: 1})
				}
				if shape == "running-full" {
					q.running["active"] = &entry{item: &model.Task{ID: "active", ConcurrencyGroup: "deploy"}}
				}
				if shape == "no-workers" {
					clear(q.workers)
				}
				for i := range size {
					task := &model.Task{
						ID:               fmt.Sprintf("task-%d", i),
						PipelineID:       int64(i + 1),
						Created:          int64(i + 1),
						ConcurrencyGroup: "deploy",
						ConcurrencyLimit: 1,
					}
					switch shape {
					case "unlimited":
						task.ConcurrencyLimit = 0
					case "many-groups":
						task.ConcurrencyGroup = fmt.Sprintf("group-%d", i%32)
					case "same-pipeline":
						task.PipelineID = 1
					case "shuffled":
						// Coprime to every fixture size: a deterministic permutation
						// exercises sorting rather than an already ordered backlog.
						task.Created = int64((i*7919)%size + 1)
					}
					q.pending.PushBack(task)
				}
				b.ReportAllocs()
				for b.Loop() {
					if element, _ := q.assignToWorker(q.newConcurrencyIndex()); element != nil {
						b.Fatal("the backlog must stay blocked")
					}
				}
			})
		}
	}
}

// BenchmarkScheduleReady includes index construction and all starts in a tick.
// Fixture restoration is excluded, but assignment/removal/running-entry costs
// are timed. The scan variant uses the pre-index concurrency rules as a baseline.
func BenchmarkScheduleReady(b *testing.B) {
	level := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	b.Cleanup(func() { zerolog.SetGlobalLevel(level) })
	for _, limited := range []bool{false, true} {
		for _, size := range []int{100, 1000, 10000} {
			for _, scan := range []bool{false, true} {
				b.Run(fmt.Sprintf("limited=%t/%d/scan=%t", limited, size, scan), func(b *testing.B) {
					q := &fifo{
						pending: list.New(), waitingOnDeps: list.New(),
						running: make(map[string]*entry), workers: make(map[*worker]struct{}),
					}
					tasks := make([]*model.Task, size)
					for i := range tasks {
						tasks[i] = &model.Task{ID: fmt.Sprintf("task-%d", i), PipelineID: int64(i), Created: int64(i)}
						if limited {
							tasks[i].ConcurrencyGroup = fmt.Sprintf("group-%d", i%8)
							tasks[i].ConcurrencyLimit = 8
						}
					}
					workers := make([]*worker, 64)
					for i := range workers {
						workers[i] = &worker{filter: filterFnTrue}
					}
					b.ReportAllocs()
					for b.Loop() {
						b.StopTimer()
						q.pending.Init()
						clear(q.running)
						clear(q.workers)
						for _, task := range tasks {
							q.pending.PushBack(task)
						}
						for _, worker := range workers {
							q.workers[worker] = struct{}{}
						}
						b.StartTimer()
						index := q.newConcurrencyIndex()
						assigned := 0
						for {
							var element *list.Element
							var worker *worker
							if scan {
								element, worker = scanAssignToWorker(q)
							} else {
								element, worker = q.assignToWorker(index)
							}
							if element == nil {
								break
							}
							task, _ := element.Value.(*model.Task)
							if !scan {
								index.start(task)
							}
							q.pending.Remove(element)
							delete(q.workers, worker)
							q.running[task.ID] = &entry{item: task, done: make(chan bool)}
							assigned++
						}
						if assigned != len(workers) {
							b.Fatalf("assigned %d tasks, want %d", assigned, len(workers))
						}
					}
				})
			}
		}
	}
}

// BenchmarkSchedulingContention runs heartbeats concurrently with blocked
// scheduling passes under the queue mutex. Use -mutexprofile to inspect waiting
// time, and compare the maximum observed heartbeat latency between variants.
func BenchmarkSchedulingContention(b *testing.B) {
	for _, scan := range []bool{false, true} {
		b.Run(fmt.Sprintf("scan=%t", scan), func(b *testing.B) {
			q := &fifo{
				pending: list.New(), waitingOnDeps: list.New(), running: make(map[string]*entry),
				workers: map[*worker]struct{}{{filter: filterFnTrue}: {}},
			}
			q.running["heartbeat"] = &entry{item: &model.Task{ID: "heartbeat"}}
			q.waitingOnDeps.PushBack(&model.Task{ID: "early", ConcurrencyGroup: "deploy"})
			for i := range 10000 {
				q.pending.PushBack(&model.Task{
					ID: fmt.Sprintf("task-%d", i), Created: int64(i + 1), PipelineID: int64(i + 1),
					ConcurrencyGroup: "deploy", ConcurrencyLimit: 1,
				})
			}
			stop, done := make(chan struct{}), make(chan struct{})
			started := make(chan struct{})
			ctx := b.Context()
			var maximum time.Duration
			go func() {
				defer close(done)
				close(started)
				for {
					select {
					case <-stop:
						return
					default:
						start := time.Now()
						_ = q.Extend(ctx, 0, "heartbeat")
						maximum = max(maximum, time.Since(start))
					}
				}
			}()
			<-started
			b.ReportAllocs()
			for b.Loop() {
				q.Lock()
				if scan {
					scanAssignToWorker(q)
				} else {
					q.assignToWorker(q.newConcurrencyIndex())
				}
				q.Unlock()
			}
			close(stop)
			<-done
			b.ReportMetric(float64(maximum.Nanoseconds()), "max-heartbeat-ns")
		})
	}
}

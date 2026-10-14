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
	"context"
	"fmt"
	"slices"
	"sync"
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

// BenchmarkQueueDispatchUnderBacklog includes the real scheduling interval and
// concurrent polling/completions. A running parent keeps the backlog blocked,
// while four independent workflows are dispatched and completed in each batch.
func BenchmarkQueueDispatchUnderBacklog(b *testing.B) {
	const workers = 4
	level := zerolog.GlobalLevel()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	b.Cleanup(func() { zerolog.SetGlobalLevel(level) })
	for _, size := range []int{1000, 10000} {
		b.Run(fmt.Sprint(size), func(b *testing.B) {
			ctx, cancel := context.WithCancelCause(b.Context())
			b.Cleanup(func() { cancel(nil) })
			q, ok := NewMemoryQueue(ctx).(*fifo)
			if !ok {
				b.Fatal("expected an in-memory FIFO queue")
			}
			q.Lock()
			q.running["root"] = &entry{
				item:     &model.Task{ID: "root"},
				done:     make(chan bool),
				deadline: time.Now().Add(time.Hour),
			}
			for i := range size {
				q.pending.PushBack(&model.Task{
					ID:           fmt.Sprintf("blocked-%d", i),
					Dependencies: []string{"root"},
					DepStatus:    make(map[string]model.StatusValue),
				})
			}
			q.filterWaiting()
			q.Unlock()

			type sample struct {
				dispatch   time.Duration
				completion time.Duration
			}
			results := make(chan sample, workers)
			var dispatch, completion []time.Duration
			batch := 0
			b.ReportAllocs()
			for b.Loop() {
				tasks := make([]*model.Task, workers)
				for i := range tasks {
					tasks[i] = &model.Task{ID: fmt.Sprintf("ready-%d-%d", batch, i)}
				}
				batch++
				pollCtx, stop := context.WithTimeout(ctx, 5*time.Second)
				started := time.Now()
				if err := q.PushAtOnce(ctx, tasks); err != nil {
					b.Fatal(err)
				}
				var wg sync.WaitGroup
				for i := range workers {
					wg.Go(func() {
						task, err := q.Poll(pollCtx, int64(i+1), filterFnTrue)
						if err != nil {
							b.Error(err)
							return
						}
						dispatched := time.Now()
						if err := q.Done(ctx, task.ID, model.StatusSuccess); err != nil {
							b.Error(err)
						}
						results <- sample{dispatched.Sub(started), time.Since(dispatched)}
					})
				}
				wg.Wait()
				stop()
				if b.Failed() {
					b.FailNow()
				}
				for range workers {
					s := <-results
					dispatch = append(dispatch, s.dispatch)
					completion = append(completion, s.completion)
				}
			}
			b.ReportMetric(float64(len(dispatch))/b.Elapsed().Seconds(), "workflows/s")
			for name, samples := range map[string][]time.Duration{"dispatch": dispatch, "completion": completion} {
				slices.Sort(samples)
				for _, percentile := range []int{50, 95, 99} {
					// Nearest-rank percentile, rounded up to a sample index.
					index := (len(samples)*percentile+99)/100 - 1
					b.ReportMetric(float64(samples[index].Nanoseconds())/1e6, fmt.Sprintf("%s-p%d-ms", name, percentile))
				}
			}
		})
	}
}

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

package pipeline

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

// statusKey names the forge status a post lands on: a workflow's commit
// status, or, for a deploy pipeline, its one deployment, whose status every
// workflow's post sets from the pipeline's state.
type statusKey struct {
	deploy bool
	id     int64
}

func statusKeyOf(pipeline *model.Pipeline, workflow *model.Workflow) statusKey {
	if pipeline.Event == model.EventDeploy {
		return statusKey{deploy: true, id: pipeline.ID}
	}
	return statusKey{id: workflow.ID}
}

// keyedMutex is one mutex per statusKey, kept only while someone holds or
// waits for it.
type keyedMutex struct {
	mu    sync.Mutex
	locks map[statusKey]*keyedLock
}

type keyedLock struct {
	sync.Mutex
	refs int
}

func (k *keyedMutex) lock(key statusKey) (unlock func()) {
	k.mu.Lock()
	if k.locks == nil {
		k.locks = make(map[statusKey]*keyedLock)
	}
	l := k.locks[key]
	if l == nil {
		l = &keyedLock{}
		k.locks[key] = l
	}
	l.refs++
	k.mu.Unlock()

	l.Lock()
	return func() {
		l.Unlock()
		k.mu.Lock()
		l.refs--
		if l.refs == 0 {
			delete(k.locks, key)
		}
		k.mu.Unlock()
	}
}

// statusLocks orders the posts of each forge status.
var statusLocks keyedMutex

// PostWorkflowStatus posts the forge status of workflow, with the workflow's
// and the pipeline's state as persisted when the post is sent. Every post of
// a status goes through here, one at a time: a forge keeps the status posted
// last, and posts of one status may be made concurrently (a pipeline posts its
// workflows' statuses after queueing them, while an agent reports a workflow
// it already runs). Each state change is persisted before it is posted, so the
// post that reaches the forge last carries the latest state.
func PostWorkflowStatus(ctx context.Context, _store store.Store, _forge forge.Forge, user *model.User, repo *model.Repo, pipeline *model.Pipeline, workflow *model.Workflow) error {
	unlock := statusLocks.lock(statusKeyOf(pipeline, workflow))
	defer unlock()

	// copies: the caller's pipeline and workflow are not ours to change
	p, w := *pipeline, *workflow
	if current, err := _store.WorkflowLoad(workflow.ID); err != nil {
		log.Error().Err(err).Msgf("cannot load workflow %d to post its status; posting the state it was handed", workflow.ID)
	} else {
		w.State = current.State
	}
	if current, err := _store.GetPipeline(pipeline.ID); err != nil {
		log.Error().Err(err).Msgf("cannot load pipeline %d to post its status; posting the state it was handed", pipeline.ID)
	} else {
		p.Status = current.Status
	}
	return _forge.Status(ctx, user, repo, &p, &w)
}

// maxConcurrentStatusUpdates bounds the parallel commit-status calls per
// pipeline. Forges rate limit concurrent writes from a single user much more
// aggressively than sequential ones, so this stays deliberately low: it is
// enough to keep pipelines with many workflows from taking tens of seconds,
// without looking like a burst.
const maxConcurrentStatusUpdates = 4

// updatePipelineStatus posts one commit status per workflow, each with the
// workflow's persisted state (PostWorkflowStatus), not the copy in pipeline.
// Every workflow is attempted even if some fail, so a single forge error
// cannot leave the remaining workflows without a status.
func updatePipelineStatus(ctx context.Context, _store store.Store, forge forge.Forge, pipeline *model.Pipeline, repo *model.Repo, user *model.User) {
	// setting one status per workflow sequentially delays pipelines with many
	// workflows by tens of seconds, so post them with bounded concurrency
	var wg sync.WaitGroup
	var failed atomic.Int64
	sem := make(chan struct{}, maxConcurrentStatusUpdates)

	for _, workflow := range pipeline.Workflows {
		wg.Add(1)
		go func() {
			defer wg.Done()

			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				failed.Add(1)
				return
			}

			if err := PostWorkflowStatus(ctx, _store, forge, user, repo, pipeline, workflow); err != nil {
				failed.Add(1)
				log.Error().Err(err).Msgf("error setting commit status for %s/%d", repo.FullName, pipeline.Number)
			}
		}()
	}
	wg.Wait()

	// individual failures are logged above, but a pipeline whose statuses all
	// failed is invisible on the forge and must not be silent
	if n := failed.Load(); n > 0 {
		log.Error().Msgf("failed to set %d of %d commit statuses for %s/%d", n, len(pipeline.Workflows), repo.FullName, pipeline.Number)
	}
}

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

// Package forgestatus keeps the statuses a forge shows of each workflow in
// line with the store.
//
// A forge keeps only the last status per commit and context, so a lost
// final status leaves a check pending for good, and an older status landing
// last shows a wrong result. The server therefore does not send statuses
// where they change: it marks the workflow, and a sender of its pipeline
// reads the workflow from the store and sends what is there at that moment.
// A status that failed is sent again, with growing waits, for up to a
// window; by then the store may hold a newer one, which goes instead, with
// a window of its own. So:
//
//   - the newest status always goes last: a pipeline has one sender at a
//     time, and it sends what the store holds when it sends;
//   - changes made while a forge was failing collapse into the last one;
//   - nobody waits for the forge: marking costs a map write.
//
// A status the forge took is not sent again: the server marks a workflow
// several times while its state stays the same.
//
// Marks live in memory; a status not yet through when the server stops is
// not sent.
package forgestatus

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v7"
	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	store_types "go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

// Policy is how hard a status is pushed through.
type Policy struct {
	Attempt   time.Duration // bounds each request to the forge
	FirstWait time.Duration // after the first failure; grows up to MaxWait
	MaxWait   time.Duration
	Window    time.Duration // how long a failing status is tried
	Senders   int           // pipelines synced at once
}

// DefaultPolicy is the policy of the server.
var DefaultPolicy = Policy{
	Attempt:   10 * time.Second, //nolint:mnd
	FirstWait: time.Second,
	MaxWait:   time.Minute,
	Window:    10 * time.Minute, //nolint:mnd
	Senders:   8,                //nolint:mnd
}

// withDefaults fills the fields left zero from DefaultPolicy.
func (p Policy) withDefaults() Policy {
	d := DefaultPolicy
	for _, f := range []struct {
		v   *time.Duration
		def time.Duration
	}{{&p.Attempt, d.Attempt}, {&p.FirstWait, d.FirstWait}, {&p.MaxWait, d.MaxWait}, {&p.Window, d.Window}} {
		if *f.v <= 0 {
			*f.v = f.def
		}
	}
	if p.Senders <= 0 {
		p.Senders = d.Senders
	}
	return p
}

// maxSent is how many sent statuses are remembered; forgetting them costs
// at most one repeated request each.
const maxSent = 10000

// Store is what the sender reads.
type Store interface {
	GetPipeline(id int64) (*model.Pipeline, error)
	GetRepo(id int64) (*model.Repo, error)
	GetUser(id int64) (*model.User, error)
	WorkflowList(pipeline *model.Pipeline) ([]*model.Workflow, error)
}

// Sender sends the statuses of marked workflows to their forges.
type Sender struct {
	ctx     context.Context
	store   Store
	forgeOf func(*model.Repo) (forge.Forge, error)
	refresh func(context.Context, forge.Forge, *model.User)
	policy  Policy
	slots   chan struct{}

	mu     sync.Mutex
	marked map[int64]*marks // pipeline ID -> workflows to sync
	busy   map[int64]bool   // pipelines with a sender
	sent   map[int64]string // workflow ID -> the status its forge took
	wg     sync.WaitGroup   // senders, for tests
}

// marks are the workflows of one pipeline to sync.
type marks struct {
	all       bool
	workflows map[int64]bool
}

func (m *marks) add(o *marks) {
	m.all = m.all || o.all
	for id := range o.workflows {
		m.workflows[id] = true
	}
}

func (m *marks) empty() bool { return !m.all && len(m.workflows) == 0 }

func newMarks() *marks { return &marks{workflows: make(map[int64]bool)} }

// New returns a sender that sends statuses until ctx is done.
func New(ctx context.Context, s store.Store, forgeOf func(*model.Repo) (forge.Forge, error), policy Policy) *Sender {
	return newSender(ctx, s, forgeOf, func(ctx context.Context, f forge.Forge, u *model.User) {
		forge.Refresh(ctx, f, s, u)
	}, policy)
}

func newSender(ctx context.Context, s Store, forgeOf func(*model.Repo) (forge.Forge, error), refresh func(context.Context, forge.Forge, *model.User), policy Policy) *Sender {
	policy = policy.withDefaults()
	return &Sender{
		ctx:     ctx,
		store:   s,
		forgeOf: forgeOf,
		refresh: refresh,
		policy:  policy,
		slots:   make(chan struct{}, policy.Senders),
		marked:  make(map[int64]*marks),
		busy:    make(map[int64]bool),
		sent:    make(map[int64]string),
	}
}

// Pipeline marks every workflow of the pipeline. A nil sender does nothing.
func (s *Sender) Pipeline(pipelineID int64) {
	if s == nil {
		return
	}
	m := newMarks()
	m.all = true
	s.mark(pipelineID, m)
}

// Workflow marks one workflow of the pipeline. A nil sender does nothing.
func (s *Sender) Workflow(pipelineID, workflowID int64) {
	if s == nil {
		return
	}
	m := newMarks()
	m.workflows[workflowID] = true
	s.mark(pipelineID, m)
}

func (s *Sender) mark(pipelineID int64, m *marks) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if prev, ok := s.marked[pipelineID]; ok {
		prev.add(m)
	} else {
		s.marked[pipelineID] = m
	}
	if !s.busy[pipelineID] {
		s.busy[pipelineID] = true
		s.wg.Add(1)
		go s.run(pipelineID)
	}
}

// take returns the marks of the pipeline added to todo, and clears them.
// With nothing to do it lets the pipeline go, in the same step, so a mark
// added meanwhile always finds a sender.
func (s *Sender) take(pipelineID int64, todo *marks) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.marked[pipelineID]; ok {
		todo.add(m)
		delete(s.marked, pipelineID)
	}
	if todo.empty() || s.ctx.Err() != nil {
		delete(s.busy, pipelineID)
		return false
	}
	return true
}

// run sends the statuses of the pipeline until none is marked or failing.
func (s *Sender) run(pipelineID int64) {
	defer s.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("statuses of pipeline %d: sender panicked: %v", pipelineID, r)
			s.restart(pipelineID)
		}
	}()

	waits := backoff.NewExponentialBackOff()
	waits.InitialInterval, waits.MaxInterval = s.policy.FirstWait, s.policy.MaxWait
	p := &pipelineSync{id: pipelineID, failing: make(map[int64]*failure)}
	todo := newMarks()
	for s.take(pipelineID, todo) {
		retry, after, ok := s.syncInSlot(p, todo)
		if !ok {
			continue // the server stops; take lets go
		}
		todo = retry
		if todo.empty() {
			waits.Reset()
			continue
		}
		wait := waits.NextBackOff()
		if after > 0 {
			// as asked, plus a little, so that senders stalled by one rate
			// limit do not all retry the same instant; not past the window,
			// which would hold back newer statuses that long
			wait = min(after+time.Duration(rand.Int64N(int64(s.policy.FirstWait)+1)), s.policy.Window) //nolint:gosec
		}
		sleep(s.ctx, wait)
	}
}

// restart gives the pipeline a new sender if marks came while the one
// before panicked; the marks it was sending are dropped, so that a status
// that panics is not sent forever.
func (s *Sender) restart(pipelineID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.marked[pipelineID]; ok && !m.empty() && s.ctx.Err() == nil {
		s.wg.Add(1)
		go s.run(pipelineID)
		return
	}
	delete(s.busy, pipelineID)
}

// pipelineSync is what the sender of one pipeline keeps between attempts.
type pipelineSync struct {
	id      int64
	loading *failure           // failures to read the pipeline
	failing map[int64]*failure // workflow ID -> failures of its status
}

// failure is how a status, or reading the pipeline, has failed so far.
type failure struct {
	value    string // the status that failed
	since    time.Time
	attempts int
}

// fail records a failed attempt at value. A status other than the one
// failing so far gets a window of its own.
func fail(f *failure, value string) *failure {
	if f == nil || f.value != value {
		f = &failure{value: value, since: time.Now()}
	}
	f.attempts++
	return f
}

func (f *failure) expired(window time.Duration) bool { return time.Since(f.since) > window }

// syncInSlot is sync within one of the slots; ok is false when the server
// stopped before a slot was free.
func (s *Sender) syncInSlot(p *pipelineSync, todo *marks) (retry *marks, after time.Duration, ok bool) {
	select {
	case s.slots <- struct{}{}:
	case <-s.ctx.Done():
		return newMarks(), 0, false
	}
	defer func() { <-s.slots }()
	retry, after = s.sync(p, todo)
	return retry, after, true
}

// outcome is what a status attempt leads to.
type outcome int

const (
	taken   outcome = iota // the forge has the status
	refused                // the forge will not take it; do not send it again
	failed                 // it may pass if sent again
)

// classify tells what the error of a status attempt leads to, and how long
// the forge asked to wait before the next.
func classify(err error) (outcome, time.Duration) {
	var statusErr *types.StatusError
	switch {
	case err == nil, errors.Is(err, types.ErrNotImplemented):
		return taken, 0
	case errors.As(err, &statusErr) && statusErr.Final:
		return refused, 0
	case errors.As(err, &statusErr):
		return failed, statusErr.RetryAfter
	default:
		return failed, 0
	}
}

// sync sends the statuses of the marked workflows of the pipeline. It
// returns the workflows to try again, and how long a forge asked to wait.
func (s *Sender) sync(p *pipelineSync, todo *marks) (*marks, time.Duration) {
	retry := newMarks()
	pipeline, user, repo, f, workflows, err := s.load(p.id)
	if errors.Is(err, errGone) {
		return retry, 0
	}
	if err != nil {
		// nothing could be sent; try it all again, within the window
		p.loading = fail(p.loading, "")
		if p.loading.expired(s.policy.Window) {
			log.Error().Err(err).Msgf("statuses of pipeline %d: gave up after %d attempts", p.id, p.loading.attempts)
			p.loading = nil
			return retry, 0
		}
		if p.loading.attempts == 1 {
			log.Warn().Err(err).Msgf("statuses of pipeline %d: trying again", p.id)
		}
		return todo, 0
	}
	p.loading = nil

	var after time.Duration
	for _, w := range workflows {
		if !todo.all && !todo.workflows[w.ID] {
			continue
		}
		value := statusValue(pipeline, w)
		if s.sentValue(w.ID) == value {
			continue
		}

		ctx, cancel := context.WithTimeout(s.ctx, s.policy.Attempt)
		err := f.Status(ctx, user, repo, pipeline, w)
		cancel()
		what := fmt.Sprintf("status %s of %s#%d %s", w.State, repo.FullName, pipeline.Number, w.Name)

		result, wait := classify(err)
		if result == failed && s.ctx.Err() != nil {
			return newMarks(), 0 // the server stops
		}
		switch result {
		case taken:
			s.setSent(w.ID, value)
			if fl := p.failing[w.ID]; fl != nil && fl.value == value {
				log.Info().Msgf("%s: sent after %d failed attempts in %s", what, fl.attempts, time.Since(fl.since).Round(time.Second))
			}
			delete(p.failing, w.ID)
		case refused:
			s.setSent(w.ID, "")
			delete(p.failing, w.ID)
			log.Error().Err(err).Msgf("%s: refused by the forge", what)
		case failed:
			s.setSent(w.ID, "") // the forge may have stored it anyway
			fl := fail(p.failing[w.ID], value)
			p.failing[w.ID] = fl
			if fl.expired(s.policy.Window) {
				log.Error().Err(err).Msgf("%s: gave up after %d attempts in %s", what, fl.attempts, time.Since(fl.since).Round(time.Second))
				delete(p.failing, w.ID)
				continue
			}
			if fl.attempts == 1 {
				log.Warn().Err(err).Msgf("%s: failed, trying again", what)
			}
			after = max(after, wait)
			retry.workflows[w.ID] = true
		}
	}
	return retry, after
}

// statusValue is what decides the status a forge shows of the workflow:
// its state, and the pipeline status, which a deployment shows. The rest
// of what a forge sends (context, link, description) follows from them or
// does not change.
func statusValue(pipeline *model.Pipeline, w *model.Workflow) string {
	return fmt.Sprintf("%s %s", w.State, pipeline.Status)
}

// errGone: what the statuses of a pipeline need is not in the store, so
// they cannot be sent; load logged why.
var errGone = errors.New("gone")

// load reads what the statuses of the pipeline need, as the store holds it.
func (s *Sender) load(pipelineID int64) (*model.Pipeline, *model.User, *model.Repo, forge.Forge, []*model.Workflow, error) {
	pipeline, err := s.store.GetPipeline(pipelineID)
	if errors.Is(err, store_types.ErrRecordNotExist) {
		log.Debug().Msgf("statuses of pipeline %d: the pipeline was deleted", pipelineID)
		return nil, nil, nil, nil, nil, errGone
	}
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	repo, err := s.store.GetRepo(pipeline.RepoID)
	if errors.Is(err, store_types.ErrRecordNotExist) {
		log.Error().Msgf("statuses of pipeline %d: its repository %d is not in the store", pipelineID, pipeline.RepoID)
		return nil, nil, nil, nil, nil, errGone
	}
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	user, err := s.store.GetUser(repo.UserID)
	if errors.Is(err, store_types.ErrRecordNotExist) {
		log.Error().Msgf("statuses of pipeline %d: user %d of repository %s is not in the store", pipelineID, repo.UserID, repo.FullName)
		return nil, nil, nil, nil, nil, errGone
	}
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	workflows, err := s.store.WorkflowList(pipeline)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	f, err := s.forgeOf(repo)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}
	ctx, cancel := context.WithTimeout(s.ctx, s.policy.Attempt)
	defer cancel()
	s.refresh(ctx, f, user)
	return pipeline, user, repo, f, workflows, nil
}

func (s *Sender) sentValue(workflowID int64) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sent[workflowID]
}

// setSent records the status the forge took for the workflow; "" forgets it.
func (s *Sender) setSent(workflowID int64, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if value == "" {
		delete(s.sent, workflowID)
		return
	}
	if len(s.sent) >= maxSent {
		clear(s.sent)
	}
	s.sent[workflowID] = value
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
	case <-ctx.Done():
	}
}

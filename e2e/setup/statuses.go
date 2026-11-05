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

//go:build test

package setup

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/forgestatus"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// statusPolicy is the server's, in milliseconds.
//
//nolint:mnd
var statusPolicy = forgestatus.Policy{
	Attempt:   5 * time.Second,
	FirstWait: 10 * time.Millisecond,
	MaxWait:   50 * time.Millisecond,
	Window:    10 * time.Second,
	Senders:   4,
}

// Statuses is the mock forge's side of commit statuses: it records the
// statuses the forge took, and can fail requests.
type Statuses struct {
	mu       sync.Mutex
	taken    map[int64][]model.StatusValue // workflow ID -> states taken, in order
	failNext int
}

func newStatuses() *Statuses {
	return &Statuses{taken: make(map[int64][]model.StatusValue)}
}

// FailNext makes the next n status requests fail as a forge that is down.
func (s *Statuses) FailNext(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failNext = n
}

func (s *Statuses) status(_ context.Context, _ *model.User, _ *model.Repo, _ *model.Pipeline, w *model.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failNext > 0 {
		s.failNext--
		return errors.New("forge unavailable")
	}
	s.taken[w.ID] = append(s.taken[w.ID], w.State)
	return nil
}

// Taken returns the states the forge took for the workflow, in order.
func (s *Statuses) Taken(workflowID int64) []model.StatusValue {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]model.StatusValue(nil), s.taken[workflowID]...)
}

// WaitForLast waits until the last status the forge took for the workflow
// is state.
func (s *Statuses) WaitForLast(t *testing.T, workflowID int64, state model.StatusValue, timeout time.Duration) {
	t.Helper()
	require.Eventually(t, func() bool {
		taken := s.Taken(workflowID)
		return len(taken) > 0 && taken[len(taken)-1] == state
	}, timeout, defaultInterval, "last status of workflow %d: want %s, the forge took %v", workflowID, state, s.Taken(workflowID))
}

// FailuresLeft is how many of the requests FailNext asked to fail have not
// come yet.
func (s *Statuses) FailuresLeft() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.failNext
}

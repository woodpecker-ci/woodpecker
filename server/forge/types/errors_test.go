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

package types

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestErrIgnoreEvent(t *testing.T) {
	t.Parallel()

	t.Run("message without reason", func(t *testing.T) {
		t.Parallel()
		err := &ErrIgnoreEvent{Event: "push"}
		assert.Equal(t, "explicit ignored event 'push'", err.Error())
	})

	t.Run("message with reason", func(t *testing.T) {
		t.Parallel()
		err := &ErrIgnoreEvent{Event: "push", Reason: "deleted branch"}
		assert.Equal(t, "explicit ignored event 'push', reason: deleted branch", err.Error())
	})

	t.Run("errors.Is matches any ErrIgnoreEvent", func(t *testing.T) {
		t.Parallel()
		err := &ErrIgnoreEvent{Event: "push"}
		assert.ErrorIs(t, err, &ErrIgnoreEvent{})
		assert.NotErrorIs(t, errors.New("other"), &ErrIgnoreEvent{})
	})
}

func TestErrConfigNotFound(t *testing.T) {
	t.Parallel()

	t.Run("message lists configs", func(t *testing.T) {
		t.Parallel()
		err := &ErrConfigNotFound{Configs: []string{".woodpecker.yml", ".woodpecker/"}}
		assert.Equal(t, "configs not found: .woodpecker.yml, .woodpecker/", err.Error())
	})

	t.Run("errors.Is matches any ErrConfigNotFound", func(t *testing.T) {
		t.Parallel()
		err := &ErrConfigNotFound{Configs: []string{"x"}}
		assert.ErrorIs(t, err, &ErrConfigNotFound{})
		assert.NotErrorIs(t, errors.New("other"), &ErrConfigNotFound{})
	})
}

func TestErrConfigFetchTimeout(t *testing.T) {
	t.Parallel()

	t.Run("message names timeout, attempts and configs", func(t *testing.T) {
		t.Parallel()
		err := &ErrConfigFetchTimeout{Timeout: 10 * time.Second, Attempts: 3, Configs: []string{".woodpecker/", ".woodpecker.yaml", ".woodpecker.yml"}}
		assert.Equal(t, "forge did not answer within 10s (tried 3 times) while fetching '.woodpecker/', '.woodpecker.yaml' and '.woodpecker.yml'; the forge may be overloaded, try again later. Admins can raise WOODPECKER_FORGE_TIMEOUT", err.Error())
	})

	t.Run("message with one config and one attempt", func(t *testing.T) {
		t.Parallel()
		err := &ErrConfigFetchTimeout{Timeout: 500 * time.Millisecond, Attempts: 1, Configs: []string{"ci.yaml"}}
		assert.Equal(t, "forge did not answer within 500ms (tried 1 time) while fetching 'ci.yaml'; the forge may be overloaded, try again later. Admins can raise WOODPECKER_FORGE_TIMEOUT", err.Error())
	})

	t.Run("errors.Is matches any ErrConfigFetchTimeout and the deadline", func(t *testing.T) {
		t.Parallel()
		err := &ErrConfigFetchTimeout{Timeout: time.Second, Attempts: 1, Configs: []string{"x"}}
		assert.ErrorIs(t, err, &ErrConfigFetchTimeout{})
		assert.ErrorIs(t, err, context.DeadlineExceeded)
		assert.NotErrorIs(t, err, &ErrConfigNotFound{})
		assert.NotErrorIs(t, errors.New("other"), &ErrConfigFetchTimeout{})
	})
}

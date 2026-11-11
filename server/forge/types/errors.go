// Copyright 2022 Woodpecker Authors
// Copyright 2018 Drone.IO Inc.
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
	"fmt"
	"strings"
	"time"
)

var (
	ErrNotImplemented          = errors.New("not implemented")
	ErrRepoNotFound            = errors.New("repo not found")
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
)

type ErrIgnoreEvent struct {
	Event  string
	Reason string
}

func (err *ErrIgnoreEvent) Error() string {
	if err.Reason != "" {
		return fmt.Sprintf("explicit ignored event '%s', reason: %s", err.Event, err.Reason)
	}
	return fmt.Sprintf("explicit ignored event '%s'", err.Event)
}

func (*ErrIgnoreEvent) Is(target error) bool {
	_, ok := target.(*ErrIgnoreEvent)
	return ok
}

type ErrConfigNotFound struct {
	Configs []string
}

func (m *ErrConfigNotFound) Error() string {
	return fmt.Sprintf("configs not found: %s", strings.Join(m.Configs, ", "))
}

func (*ErrConfigNotFound) Is(target error) bool {
	_, ok := target.(*ErrConfigNotFound)
	return ok
}

// ErrConfigFetchTimeout is returned when the forge did not deliver the pipeline config
// within the configured forge timeout on any attempt.
type ErrConfigFetchTimeout struct {
	Timeout  time.Duration
	Attempts uint
	Configs  []string
}

func (m *ErrConfigFetchTimeout) Error() string {
	configs := make([]string, 0, len(m.Configs))
	for _, conf := range m.Configs {
		configs = append(configs, fmt.Sprintf("'%s'", strings.Trim(conf, "'")))
	}

	return fmt.Sprintf("forge did not answer within %s (tried %d times) while fetching %s; "+
		"the forge may be overloaded, try again later. Admins can raise WOODPECKER_FORGE_TIMEOUT",
		m.Timeout, m.Attempts, strings.Join(configs, ", "))
}

func (*ErrConfigFetchTimeout) Is(target error) bool {
	_, ok := target.(*ErrConfigFetchTimeout)
	return ok
}

func (*ErrConfigFetchTimeout) Unwrap() error {
	return context.DeadlineExceeded
}

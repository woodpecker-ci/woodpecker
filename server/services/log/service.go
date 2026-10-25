// Copyright 2024 Woodpecker Authors
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

package log

import (
	"context"
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// StreamDelayAllowed is how long after a pipeline or step finished the agent may still send log lines.
// Once it has passed, a step's log is complete (see Service.StepFinished).
const StreamDelayAllowed = 5 * time.Minute

// AppendTimeout bounds appending log lines (rpc.Log): an append admitted inside the late-log window
// is over before the log finalizer, which waits StreamDelayAllowed plus AppendTimeout, takes the log.
const AppendTimeout = time.Minute

// releaseTimeout bounds handing back a claimed step log.
const releaseTimeout = 10 * time.Second

// ReleaseContext returns a context to hand back a claimed step log with even once ctx is done,
// else the claim stays until the log finalizer recovers it half an hour later.
func ReleaseContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), releaseTimeout)
}

// Service stores step logs. Implementations should give up once ctx is done.
type Service interface {
	LogFind(ctx context.Context, step *model.Step) ([]*model.LogEntry, error)
	LogAppend(ctx context.Context, step *model.Step, logEntries []*model.LogEntry) error
	LogDelete(ctx context.Context, step *model.Step) error
	// StepFinished is called once no further log entries can be appended for the step:
	// its pipeline and the step have stopped and the late-log window (StreamDelayAllowed) has passed.
	// Steps that never started have no log and get no call.
	// Implementations may pack, compress or move the log.
	// It must be idempotent: it is called again after a failure or a crash.
	// It may come for a log that was just deleted, which has to stay deleted.
	// LogFind must return the complete log before, during and after the call.
	StepFinished(ctx context.Context, step *model.Step) error
}

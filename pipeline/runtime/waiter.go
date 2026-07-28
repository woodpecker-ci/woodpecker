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

package runtime

import (
	"context"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// WorkflowWaiter blocks a step on its cross-workflow dependencies. The agent
// injects an implementation backed by the server's WaitWorkflow RPC; when no
// waiter is configured (e.g. woodpecker-cli exec), cross-workflow
// dependencies are ignored with a warning.
type WorkflowWaiter interface {
	// WaitWorkflowDependency blocks until dep reaches a terminal state.
	// ok reports whether the dependency is satisfied; when it is not, msg
	// explains why (target failed, was skipped, or a required target is
	// missing). err is reserved for transport failures and context
	// cancellation.
	WaitWorkflowDependency(ctx context.Context, dep backend_types.WorkflowDependency) (ok bool, msg string, err error)
}

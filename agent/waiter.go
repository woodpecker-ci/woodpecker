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

package agent

import (
	"context"
	"fmt"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	pipeline_runtime "go.woodpecker-ci.org/woodpecker/v3/pipeline/runtime"
	"go.woodpecker-ci.org/woodpecker/v3/rpc"
)

// statusSuccess mirrors server/model.StatusSuccess; the agent package does
// not depend on server code.
const statusSuccess = "success"

// workflowWaiter implements pipeline_runtime.WorkflowWaiter on top of the
// server's WaitWorkflow RPC for one workflow execution.
type workflowWaiter struct {
	client     rpc.Peer
	workflowID string
}

func (r *Runner) createWorkflowWaiter(workflow *rpc.Workflow) pipeline_runtime.WorkflowWaiter {
	return &workflowWaiter{client: r.client, workflowID: workflow.ID}
}

func (w *workflowWaiter) WaitWorkflowDependency(ctx context.Context, dep backend_types.WorkflowDependency) (bool, string, error) {
	result, err := w.client.WaitWorkflow(ctx, w.workflowID, dep.Workflow, dep.Step)
	if err != nil {
		return false, "", err
	}

	target := fmt.Sprintf("workflow '%s'", dep.Workflow)
	if dep.Step != "" {
		target = fmt.Sprintf("step '%s' of workflow '%s'", dep.Step, dep.Workflow)
	}

	if !result.Found {
		if dep.Optional {
			return true, "", nil
		}
		// the pipeline builder rejects missing required targets, so this can
		// only happen in edge cases like partial restarts
		return false, fmt.Sprintf("%s does not exist in this pipeline", target), nil
	}

	if result.Status == statusSuccess {
		return true, "", nil
	}
	// any other terminal state — including skipped, which means the target's
	// own dependencies failed — fails the dependency
	return false, fmt.Sprintf("%s finished with status %s", target, result.Status), nil
}

// Copyright 2023 Woodpecker Authors
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

package state

import (
	backend "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// State is used to signal the current workflow and step state.
// Only steps using the trace func report back what's going on.
// And the workflow is updated alongside it.
type State struct {
	// Global state of the currently running Workflow.
	Workflow struct {
		// Workflow start time
		Started int64 `json:"time"`
		// Current pipeline error state
		Error error `json:"error"`
	}

	// Current step that updates the step and workflow state
	CurrStep *backend.Step `json:"step"`

	// Current Step state.
	CurrStepState backend.State
}

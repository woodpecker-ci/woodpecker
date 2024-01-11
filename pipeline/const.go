// Copyright 2023 Woodpecker Authors
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

// StatusValue represent pipeline states woodpecker know
type StatusValue string //	@name StatusValue

const (
	StatusSkipped  StatusValue = "skipped"  // skipped as another step failed
	StatusPending  StatusValue = "pending"  // pending to be executed
	StatusRunning  StatusValue = "running"  // currently running
	StatusSuccess  StatusValue = "success"  // successfully finished
	StatusFailure  StatusValue = "failure"  // failed to finish (exit code != 0)
	StatusKilled   StatusValue = "killed"   // killed by user
	StatusError    StatusValue = "error"    // error with the config / while parsing / some other system problem
	StatusBlocked  StatusValue = "blocked"  // waiting for approval
	StatusDeclined StatusValue = "declined" // blocked and declined
	StatusCreated  StatusValue = "created"  // created / internal use only
)

const ExitCodeKilled int = 137

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

package tenki

import "errors"

var (
	ErrMissingAPIKey         = errors.New("no Tenki API key was set")
	ErrNoProjectResolved     = errors.New("could not resolve a Tenki project from the API key identity; set backend-tenki-project-id")
	ErrWorkflowStateNotFound = errors.New("workflow state not found")
	ErrStepStateNotFound     = errors.New("step state not found")
	ErrStepReaderNotFound    = errors.New("could not find log reader for step")
	ErrUnsupportedStepType   = errors.New("unsupported step type")
	ErrNoCommands            = errors.New("step has no commands to run")
)

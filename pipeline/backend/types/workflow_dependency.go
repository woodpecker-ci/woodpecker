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

package types

// WorkflowDependency identifies another workflow of the same pipeline (or a
// single step of it) that a step waits for before it starts. When Optional
// is true and the referenced workflow or step does not exist, the dependency
// is treated as satisfied.
type WorkflowDependency struct {
	Workflow string `json:"workflow"`
	Step     string `json:"step,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

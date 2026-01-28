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

package types

// State defines a container state.
type State struct {
	// Unix start time
	Started int64 `json:"started"`
	// Container exit code
	ExitCode int `json:"exit_code"`
	// Container exited, true or false
	Exited bool `json:"exited"`
	// Container is oom killed, true or false
	// TODO (6024): well known errors as string enum into ./errors.go
	OOMKilled bool `json:"oom_killed"`
	// Container error
	Error error
}

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

package docker

import "strings"

func isErrContainerNotFoundOrNotRunning(err error) bool {
	// Error response from daemon: Cannot kill container: ...: No such container: ...
	// Error response from daemon: Cannot kill container: ...: Container ... is not running"
	// Error response from podman daemon: can only kill running containers. ... is in state exited
	// Error response from daemon: removal of container ... is already in progress
	// Error: No such container: ...
	return err != nil &&
		(strings.Contains(err.Error(), "No such container") ||
			strings.Contains(err.Error(), "is not running") ||
			strings.Contains(err.Error(), "can only kill running containers") ||
			(strings.Contains(err.Error(), "removal of container") && strings.Contains(err.Error(), "is already in progress")))
}

func isErrVolumeInUse(err error) bool {
	return err != nil &&
		strings.Contains(err.Error(), "volume is in use")
}

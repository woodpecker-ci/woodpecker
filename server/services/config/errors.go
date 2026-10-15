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

package config

import "fmt"

// ErrConfigExtension is returned when a configuration extension explicitly
// rejects a pipeline by responding with HTTP status 422 Unprocessable Entity.
// Message contains the response body of the extension and is shown to the
// user as pipeline error.
type ErrConfigExtension struct {
	Message string
}

func (e *ErrConfigExtension) Error() string {
	return fmt.Sprintf("config extension error: %s", e.Message)
}

func (*ErrConfigExtension) Is(target error) bool {
	_, ok := target.(*ErrConfigExtension)
	return ok
}

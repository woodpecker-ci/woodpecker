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

package types

import "time"

// StatusError is a failed status update that says whether, and when, to try
// it again. Other errors of Forge.Status are taken as ones that may pass.
type StatusError struct {
	Err error
	// Final: the forge refused the status, sending it again cannot help.
	Final bool
	// RetryAfter is how long the forge asked to wait first; 0 if it did not.
	RetryAfter time.Duration
}

func (e *StatusError) Error() string { return e.Err.Error() }

func (e *StatusError) Unwrap() error { return e.Err }

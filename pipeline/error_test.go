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

package pipeline

import (
	"testing"
)

func TestExitError(t *testing.T) {
	err := ExitError{
		Name: "build",
		Code: 255,
	}
	got, want := err.Error(), "build : exit code 255"
	if got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}

func TestOomError(t *testing.T) {
	err := OomError{
		Name: "build",
	}
	got, want := err.Error(), "build : received oom kill"
	if got != want {
		t.Errorf("Want error message %q, got %q", want, got)
	}
}
